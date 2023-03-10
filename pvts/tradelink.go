// Copyright 2021-2023
// SPDX-License-Identifier: Apache-2.0
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package pvts

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/olekukonko/tablewriter"
	"github.com/penny-vault/tradestation/tradestation"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

type TradeLink struct {
	PortfolioID   string
	AccountID     string
	LastTradeDate time.Time
	NextTradeDate time.Time
}

type Transaction struct {
	ID             string
	Cleared        bool
	Commission     float64
	CompositeFIGI  string
	Date           string
	GainLoss       float64
	Kind           string
	Memo           string
	Predicted      bool
	PricePerShare  float64
	Shares         float64
	SourceID       string
	TaxDisposition string
	Ticker         string
	TotalValue     float64
}

type Allocation struct {
	Date    string
	Members map[string]float64
}

type PVRebalance struct {
	Allocation    *Allocation
	NextTradeDate string
	Transactions  []*Transaction
}

type PVPosition struct {
	CompositeFIGI string
	Ticker        string
	Shares        float64
}

type PVSecurity struct {
	CompositeFIGI string `json:"compositeFigi"`
	Ticker        string `json:"ticker"`
}

func pvTicker2TradeStation(ticker string) string {
	ticker = strings.ReplaceAll(ticker, "/", ".")
	if ticker == "BRK.A" {
		ticker = "BRK.B"
	}
	return ticker
}

// securityFromSymbol given `symbol` get a security object from PV-API
func securityFromSymbol(client *resty.Client, symbol string) (*PVSecurity, error) {
	security := &PVSecurity{}
	query := strings.ReplaceAll(symbol, ".", "%2F")

	resp, err := client.R().
		SetHeader("Content-Type", "application/json").
		SetResult(security).
		Get(fmt.Sprintf("/v1/security/%s/", query))
	if err != nil {
		log.Error().Err(err).Str("Ticker", symbol).Msg("could not get security")
		return nil, err
	}
	if resp.StatusCode() >= 400 {
		log.Error().Int("StatusCode", resp.StatusCode()).Str("URI", resp.RawResponse.Request.RequestURI).Str("Body", string(resp.Body())).Msg("HTTP error returned when communicating with pv-api")
		return nil, fmt.Errorf("%d status code returned from pvapi", resp.StatusCode())
	}

	return security, nil
}

func (tl *TradeLink) convertPositionsToPV(positions []*tradestation.Position) ([]*PVPosition, error) {
	client := resty.New()
	client.SetHeader("X-Pv-Api", viper.GetString("pv.apikey"))
	client.SetDebug(viper.GetBool("debug"))
	pvPos := make([]*PVPosition, len(positions))

	for idx, pos := range positions {
		myPos := &PVPosition{
			Shares: float64(pos.Quantity),
		}
		security := &PVSecurity{}
		symbol := strings.ReplaceAll(pos.Symbol, ".", "%2F")
		pvApiUrl := viper.GetString("pv.url")
		if pvApiUrl == "" {
			pvApiUrl = "https://api.pennyvault.com"
		}
		resp, err := client.R().
			SetHeader("Content-Type", "application/json").
			SetResult(security).
			Get(fmt.Sprintf("%s/v1/security/%s/", pvApiUrl, symbol))
		if err != nil {
			log.Error().Err(err).Str("Ticker", pos.Symbol).Msg("could not get security")
			return nil, err
		}
		if resp.StatusCode() >= 400 {
			log.Error().Str("PortfolioID", tl.PortfolioID).Int("StatusCode", resp.StatusCode()).Str("URI", resp.RawResponse.Request.RequestURI).Str("Body", string(resp.Body())).Msg("HTTP error returned when communicating with pv-api")
			return nil, fmt.Errorf("%d status code returned from pvapi", resp.StatusCode())
		}

		myPos.CompositeFIGI = security.CompositeFIGI
		myPos.Ticker = security.Ticker
		pvPos[idx] = myPos
	}

	return pvPos, nil
}

func (tl *TradeLink) createOrderRequests(strategyPlan *PVRebalance, balance *tradestation.Balance) []*tradestation.OrderRequest {
	orders := make([]*tradestation.OrderRequest, 0, len(strategyPlan.Transactions))

	// create tradestation orders
	for _, trx := range strategyPlan.Transactions {
		ticker := pvTicker2TradeStation(trx.Ticker)
		o := &tradestation.OrderRequest{
			AccountID:      tl.AccountID,
			LimitPrice:     trx.PricePerShare,
			OrderType:      tradestation.LIMIT,
			Quantity:       int64(trx.Shares),
			Symbol:         ticker,
			TimeInForceDur: tradestation.DAY,
		}

		switch trx.Kind {
		case "SELL":
			o.TradeAction = tradestation.SELL
		case "BUY":
			o.TradeAction = tradestation.BUY
		default:
			log.Warn().Str("TradeKind", trx.Kind).Msg("skipping transaction due to unknown transaction kind")
		}

		orders = append(orders, o)
	}

	return orders
}

// pvApiRebalanceRequest calls the pv-api rebalance REST endpoint
func pvApiRebalanceRequest(client *resty.Client, portfolioID string, allocationOnly bool, positions []*PVPosition, prices map[string]float64) (*PVRebalance, error) {
	result := &PVRebalance{
		Allocation: &Allocation{
			Members: make(map[string]float64),
		},
		Transactions: make([]*Transaction, 0, 100),
	}
	resp, err := client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(map[string]any{
			"AllocationOnly": allocationOnly,
			"Positions":      positions,
			"Precision":      0,
			"PriceData":      prices,
		}).
		SetResult(result).
		Post(fmt.Sprintf("/v1/portfolio/%s/rebalance", portfolioID))
	if err != nil {
		log.Error().Err(err).Str("PortfolioID", portfolioID).Msg("could not get strategy rebalance")
		return nil, err
	}
	if resp.StatusCode() >= 400 {
		log.Error().Str("PortfolioID", portfolioID).Int("StatusCode", resp.StatusCode()).Str("Body", string(resp.Body())).Str("URI", resp.RawResponse.Request.RequestURI).Msg("HTTP error returned when communicating with pv-api")
		return nil, fmt.Errorf("%d status code returned from pvapi", resp.StatusCode())
	}
	return result, nil
}

// GetStrategy communicates with pv-api and gets the list of transactions that TradeStation should execute
func (tl *TradeLink) GetStrategy(positions []*PVPosition, balance *tradestation.Balance) (*PVRebalance, error) {
	positions = append(positions, &PVPosition{
		CompositeFIGI: "$CASH",
		Ticker:        "$CASH",
		Shares:        balance.CashBalance,
	})

	// create resty client
	client := resty.New()
	client.SetHeader("X-Pv-Api", viper.GetString("pv.apikey"))
	client.SetDebug(viper.GetBool("debug"))
	pvApiUrl := viper.GetString("pv.url")
	if pvApiUrl == "" {
		pvApiUrl = "https://api.pennyvault.com"
	}
	client.SetBaseURL(pvApiUrl)

	// get list of allocations that portfolio will transition to
	log.Info().Msg("getting allocation from pvapi")
	result, err := pvApiRebalanceRequest(client, tl.PortfolioID, true, make([]*PVPosition, 0), make(map[string]float64))
	if err != nil {
		// error logged by sender
		return nil, err
	}

	log.Info().Int("Num assets in allocation", len(result.Allocation.Members)).Msg("got allocation guidance from pv-api")

	// get price list for all positions and future allocations
	log.Info().Msg("getting price data from tradestation")
	api := tradestation.New()
	tickerMap := make(map[string]bool)
	for figi := range result.Allocation.Members {
		security, err := securityFromSymbol(client, figi)
		if err != nil {
			log.Error().Str("FIGI", figi).Msg("could not find security for given figi")
			return nil, err
		}
		ticker := pvTicker2TradeStation(security.Ticker)
		tickerMap[ticker] = true
	}
	for _, pos := range positions {
		if pos.Ticker != "$CASH" {
			ticker := pvTicker2TradeStation(pos.Ticker)
			tickerMap[ticker] = true
		}
	}
	tickers := make([]string, 0, len(tickerMap))
	for t := range tickerMap {
		tickers = append(tickers, t)
	}
	quotes, err := api.GetQuotes(tickers)
	if err != nil {
		log.Error().Err(err).Strs("tickers", tickers).Msg("could not get quotes for tickers")
		return nil, err
	}

	// Get rebalance plan with current prices
	log.Info().Msg("translating tickers to figi's")
	prices := make(map[string]float64)
	for _, q := range quotes {
		security, err := securityFromSymbol(client, q.Symbol)
		if err != nil {
			log.Error().Err(err).Str("ticker", q.Symbol).Msg("could not translate ticker to figi")
			return nil, err
		}
		prices[security.CompositeFIGI] = q.Bid + ((q.Ask - q.Bid) / 2)
		if q.Symbol == "BRK.B" {
			// use BRK.B price for BRK.A
			security, err := securityFromSymbol(client, "BRK.A")
			if err != nil {
				log.Error().Err(err).Str("ticker", q.Symbol).Msg("could not translate ticker to figi")
				return nil, err
			}
			prices[security.CompositeFIGI] = q.Bid + ((q.Ask - q.Bid) / 2)
		}
	}
	result, err = pvApiRebalanceRequest(client, tl.PortfolioID, false, positions, prices)
	if err != nil {
		log.Error().Err(err).Msg("failed to get rebalance plan from pvapi")
		return nil, err
	}

	log.Info().Int("NumTransactions", len(result.Transactions)).Msg("got transaction plan from pv-api")

	return result, nil
}

// Sync gets a list of transactions from penny-vault and executes them in Trade Station
func (tl *TradeLink) Sync(autoConfirm bool) error {
	subLog := log.With().Str("AccountID", tl.AccountID).Str("PortfolioID", tl.PortfolioID).Logger()

	// check if the account should be synchronized
	now := time.Now()
	if (!tl.LastTradeDate.Equal(time.Time{}) && now.Before(tl.NextTradeDate)) {
		log.Info().Msg("no trades necessary - next trade date has not arrived")
		return nil
	}

	// get current positions in account
	api := tradestation.New()
	account, err := api.GetAccount(tl.AccountID)
	if err != nil {
		subLog.Error().Err(err).Msg("could not get account from tradestation")
		return err
	}

	positions, err := account.GetPositions()
	if err != nil {
		subLog.Error().Err(err).Msg("could not load account positions")
		return err
	}

	pvPositions, err := tl.convertPositionsToPV(positions)
	if err != nil {
		subLog.Error().Err(err).Msg("could not convert positions to pv api format")
		return err
	}

	balance, err := account.GetBalances()
	if err != nil {
		subLog.Error().Err(err).Str("AccountID", account.AccountID).Msg("could not get account balances")
		return err
	}

	// get instructions from pv-api
	strategyPlan, err := tl.GetStrategy(pvPositions, balance)
	if err != nil {
		subLog.Error().Err(err).Msg("could not get strategy instructions")
		return err
	}

	// create order requests for each transaction
	orderReqs := tl.createOrderRequests(strategyPlan, balance)

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"", "Symbol", "Action", "Shares", "Limit Price", "Expected Cost"})
	table.SetBorder(false)

	cashLeft := balance.CashBalance
	for idx, o := range orderReqs {
		if o.TradeAction == "BUY" {
			cashLeft -= (o.LimitPrice * float64(o.Quantity))
		} else {
			cashLeft += (o.LimitPrice * float64(o.Quantity))
		}
		row := []string{fmt.Sprintf("%d", idx), o.Symbol, string(o.TradeAction), fmt.Sprintf("%d", o.Quantity), fmt.Sprintf("%.2f", o.LimitPrice), fmt.Sprintf("%.2f", o.LimitPrice*float64(o.Quantity))}
		table.Append(row)
	}

	table.Render()
	fmt.Printf("Cash Left: %.2f\n", cashLeft)

	confirmed := false
	if autoConfirm {
		confirmed = true
	} else {
		fmt.Println("\nDo you wish to execute the suggested transactions (Y/n)? ")
		var confirmInput string
		fmt.Scanln(&confirmInput)
		confirmInput = strings.ToUpper(confirmInput)
		if confirmInput == "Y" {
			confirmed = true
		}
	}

	if !confirmed {
		return errors.New("user did not confirm transactions")
	}

	orders, err := account.PlaceGroupOrder(orderReqs)
	if err != nil {
		log.Error().Err(err).Msg("error placing orders")
	}

	table = tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"", "Symbol", "Action", "Order ID", "Status", "# Filled", "# Remaining"})
	table.SetBorder(false)
	for idx, o := range orders {
		if len(o.Legs) == 0 {
			row := []string{fmt.Sprintf("%d.0", idx+1), "-", "-", o.OrderID, o.StatusDescription, "-", "-"}
			table.Append(row)
		}
		for idx2, leg := range o.Legs {
			row := []string{fmt.Sprintf("%d.%d", idx+1, idx2), leg.Symbol, leg.BuyOrSell, o.OrderID, o.StatusDescription, fmt.Sprintf("%d", leg.ExecQuantity), fmt.Sprintf("%d", leg.QuantityRemaining)}
			table.Append(row)
		}
	}

	table.Render()

	return nil
}
