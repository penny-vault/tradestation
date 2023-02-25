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

package tradestation

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/rs/zerolog/log"
)

type accountResponse struct {
	Accounts []*Account
}

type Account struct {
	AccountID   string
	Alias       string
	Currency    string
	Status      string
	AccountType string
	api         *API
}

type tsError struct {
	AccountID string
	Error     string
	Message   string
}

type tsBalance struct {
	AccountID     string
	AccountType   string
	BalanceDetail struct {
		CostOfPositions      string
		DayTrades            string
		MaintenanceRate      string
		OvernightBuyingPower string
		RequiredMargin       string
		RealizedProfitLoss   string
		UnrealizedProfitLoss string
		UnsettledFunds       string
	}
	BuyingPower      string
	CashBalance      string
	Commission       string
	Equity           string
	MarketValue      string
	TodaysProfitLoss string
	UnclearedDeposit string
}

type balanceResponse struct {
	Balances []*tsBalance
	Errors   []*tsError
}

type Balance struct {
	AccountID            string
	AccountType          string
	BuyingPower          float64
	CashBalance          float64
	Commission           float64
	CostOfPositions      float64
	DayTrades            float64
	Equity               float64
	MaintenanceRate      float64
	MarketValue          float64
	OvernightBuyingPower float64
	TodaysProfitLoss     float64
	RealizedProfitLoss   float64
	RequiredMargin       float64
	UnclearedDeposit     float64
	UnrealizedProfitLoss float64
}

type orderResponse struct {
	Orders    []*tsOrder
	Errors    []*tsError
	NextToken string
}

type tsOrder struct {
	AccountID               string
	AdvancedOptions         string
	ClosedDateTime          string
	CommissionFee           string
	ConditionalOrders       []*LinkedOrder
	ConversionRate          string
	Currency                string
	Duration                string
	FilledPrice             string
	GoodTillDate            string
	GroupName               string
	Legs                    []*tsOrderLeg
	MarketActivationRules   []*tsMarketRule
	TimeActivationRules     []*tsTimeRule
	LimitPrice              string
	OpenedDateTime          string
	OrderID                 string
	OrderType               string
	PriceUsedForBuyingPower string
	RejectReason            string
	Routing                 string
	ShowOnlyQuantity        string
	Spread                  string
	Status                  string
	StatusDescription       string
	StopPrice               string
	UnbundledRouteFee       string
}

type tsOrderLeg struct {
	OpenOrClose       string
	QuantityOrdered   string
	ExecQuantity      string
	QuantityRemaining string
	BuyOrSell         string
	Symbol            string
	AssetType         string
}

type tsMarketRule struct {
	RuleType   string
	Symbol     string
	Predicate  string
	TriggerKey string
	Price      string
}

type tsTimeRule struct {
	TimeUtc string
}

type LinkedOrder struct {
	OrderID      string
	Relationship string
}

type OrderLeg struct {
	OpenOrClose       string
	QuantityOrdered   int64
	ExecQuantity      int64
	QuantityRemaining int64
	BuyOrSell         string
	Symbol            string
	AssetType         string
}

type OrderStatus string

const (
	RECEIVED              OrderStatus = "ACK"
	BROKEN                OrderStatus = "BRO"
	CANCELED              OrderStatus = "CAN"
	EXPIRED               OrderStatus = "EXP"
	FILLED                OrderStatus = "FLL"
	PARTIAL_FILL          OrderStatus = "FLP"
	PARTIAL_FILL_ALIVE    OrderStatus = "FPR"
	TOO_LATE_TO_CANCEL    OrderStatus = "LAT"
	OPEN                  OrderStatus = "OPN"
	OUT                   OrderStatus = "OUT"
	REJECTED              OrderStatus = "REJ"
	REPLACED              OrderStatus = "UCH"
	CANCEL_SENT           OrderStatus = "UCN"
	TRADE_SERVER_CANCELED OrderStatus = "TSC"
	CANCEL_REJECTED       OrderStatus = "RJC"
	QUEUED                OrderStatus = "DON"
	REPLACE_SENT          OrderStatus = "RSN"
	CONDITION_MET         OrderStatus = "CND"
	OSO_ORDER             OrderStatus = "OSO"
	SUSPENDED             OrderStatus = "SUS"
)

type MarketRuleType string

const (
	PRICE MarketRuleType = "Price"
)

type MarketRulePredicate string

const (
	LT  MarketRulePredicate = "Lt"
	LTE MarketRulePredicate = "Lte"
	GT  MarketRulePredicate = "Gt"
	GTE MarketRulePredicate = "Gte"
)

type MarketRuleTrigger string

const (
	SINGLE_TRADE_TICK      MarketRuleTrigger = "STT"  // One trade tick must print within your stop price to trigger your stop.
	SINGLE_TRADE_TICK_NBBO MarketRuleTrigger = "STTN" // One trade tick within the National Best Bid or Offer must print within your stop price to trigger your stop.
	SINGLE_BIDASK_TICK     MarketRuleTrigger = "SBA"  // Buy/Cover Orders: One Ask tick must print within your stop price to trigger your stop. Sell/Short Orders: One Bid tick must print within your stop price to trigger your stop.
	SINGLE_ASKBID_TICK     MarketRuleTrigger = "SAB"  // Buy/Cover Orders: One Bid tick must print within your stop price to trigger your stop. Sell/Short Orders: One Ask tick must print within your stop price to trigger your stop.
	DOUBLE_TRADE_TICK      MarketRuleTrigger = "DTT"  // Two consecutive trade ticks must print within your stop price to trigger your stop.
	DOUBLE_TRADE_TICK_NBBO MarketRuleTrigger = "DTTN" // Two consecutive trade ticks within the National Best Bid or Offer must print within your stop price to trigger your stop.
	DOUBLE_BIDASK_TICK     MarketRuleTrigger = "DBA"  // Buy/Cover Orders: Two consecutive Ask ticks must print within your stop price to trigger your stop. Sell/Short Orders: Two consecutive Bid ticks must print within your stop price to trigger your stop.
	DOUBLE_ASKBID_TICK     MarketRuleTrigger = "DAB"  // Buy/Cover Orders: Two consecutive Bid ticks must print within your stop price to trigger your stop. Sell/Short Orders: Two consecutive Ask ticks must print within your stop price to trigger your stop.
	TWICE_TRADE_TICK       MarketRuleTrigger = "TTT"  // Two trade ticks must print within your stop price to trigger your stop.
	TWICE_TRADE_TICK_NBBO  MarketRuleTrigger = "TTTN" // Two trade ticks within the National Best Bid or Offer must print within your stop price to trigger your stop.
	TWICE_BIDASK_TICK      MarketRuleTrigger = "TBA"  // Buy/Cover Orders: Two Ask ticks must print within your stop price to trigger your stop. Sell/Short Orders: Two Bid ticks must print within your stop price to trigger your stop.
	TWICE_ASKBID_TICK      MarketRuleTrigger = "TAB"  // Buy/Cover Orders: Two Bid ticks must print within your stop price to trigger your stop. Sell/Short Orders: Two Ask ticks must print within your stop price to trigger your stop.
)

type MarketRule struct {
	RuleType   MarketRuleType
	Symbol     string
	Predicate  MarketRulePredicate
	TriggerKey MarketRuleTrigger
	Price      float64
}

type TimeRule struct {
	TimeUtc time.Time
}

type Order struct {
	AccountID               string
	AdvancedOptions         string
	ClosedDateTime          time.Time
	CommissionFee           float64
	ConditionalOrders       []*LinkedOrder
	Duration                string
	FilledPrice             float64
	GoodTillDate            time.Time
	GroupName               string
	Legs                    []*OrderLeg
	MarketActivationRules   []*MarketRule
	OrderID                 string
	OpenedDateTime          time.Time
	OrderType               string
	PriceUsedForBuyingPower float64
	RejectReason            string
	Routing                 string
	Status                  OrderStatus
	StatusDescription       string
	TimeActivationRules     []*TimeRule
	UnbundledRouteFee       float64
}

type tsPosition struct {
	AccountID                   string
	AveragePrice                string
	AssetType                   string
	Last                        string
	Bid                         string
	Ask                         string
	PositionID                  string
	LongShort                   string
	Quantity                    string
	Symbol                      string
	Timestamp                   string
	TodaysProfitLoss            string
	TotalCost                   string
	MarketValue                 string
	MarkToMarketPrice           string
	UnrealizedProfitLoss        string
	UnrealizedProfitLossPercent string
	UnrealizedProfitLossQty     string
}

type positionResponse struct {
	Positions []*tsPosition
	Errors    []*tsError
}

type Position struct {
	AccountID                   string
	AveragePrice                float64
	AssetType                   string
	Last                        float64
	Bid                         float64
	Ask                         float64
	PositionID                  string
	LongShort                   string
	Quantity                    int64
	Symbol                      string
	Timestamp                   time.Time
	TodaysProfitLoss            float64
	TotalCost                   float64
	MarketValue                 float64
	MarkToMarketPrice           float64
	UnrealizedProfitLoss        float64
	UnrealizedProfitLossPercent float64
	UnrealizedProfitLossQty     float64
}

func (api *API) GetAccount(accountID string) (*Account, error) {
	accounts, err := api.GetAccounts()
	if err != nil {
		log.Error().Err(err).Str("Requested AccountID", accountID).Msg("error retrieving account")
		return nil, err
	}
	for _, account := range accounts {
		if accountID == account.AccountID {
			return account, nil
		}
	}
	return nil, nil
}

func (api *API) GetAccounts() ([]*Account, error) {
	api.CheckAuth()
	accounts := accountResponse{
		Accounts: make([]*Account, 0, 5),
	}
	resp, err := api.client.R().
		SetResult(&accounts).
		Get("/brokerage/accounts")
	if err != nil {
		log.Error().Err(err).Msg("account request failed")
		return nil, err
	}
	if resp.StatusCode() >= 400 {
		return nil, fmt.Errorf("/brokerage/accounts %d", resp.StatusCode())
	}

	// set api on returned accounts
	for _, acct := range accounts.Accounts {
		acct.api = api
	}

	return accounts.Accounts, nil
}

func (account *Account) GetBalances() (*Balance, error) {
	account.api.CheckAuth()
	balances := balanceResponse{
		Balances: make([]*tsBalance, 0, 1),
		Errors:   make([]*tsError, 0, 1),
	}

	resp, err := account.api.client.R().
		SetResult(&balances).
		Get(fmt.Sprintf("/brokerage/accounts/%s/balances", account.AccountID))
	if err != nil {
		log.Error().Err(err).Msg("account request failed")
		return nil, err
	}
	if resp.StatusCode() >= 400 {
		return nil, fmt.Errorf("/brokerage/accounts/balances %d", resp.StatusCode())
	}

	return parseBalances(balances)
}

func (account *Account) GetBalancesBOD() (*Balance, error) {
	account.api.CheckAuth()

	balances := balanceResponse{
		Balances: make([]*tsBalance, 0, 1),
		Errors:   make([]*tsError, 0, 1),
	}

	resp, err := account.api.client.R().
		SetResult(&balances).
		Get(fmt.Sprintf("/brokerage/accounts/%s/bodbalances", account.AccountID))
	if err != nil {
		log.Error().Err(err).Msg("account request failed")
		return nil, err
	}
	if resp.StatusCode() >= 400 {
		return nil, fmt.Errorf("/brokerage/accounts %d", resp.StatusCode())
	}

	return parseBalances(balances)
}

func parseBalances(balances balanceResponse) (*Balance, error) {
	if len(balances.Errors) > 0 {
		for _, err := range balances.Errors {
			return nil, fmt.Errorf("%s: %s", err.Error, err.Message)
		}
	}

	// convert struct to proper types
	resBalance := make([]*Balance, len(balances.Balances))
	for idx, balance := range balances.Balances {
		var err error
		b := &Balance{
			AccountID:   balance.AccountID,
			AccountType: balance.AccountType,
		}

		if balance.BuyingPower != "" {
			if b.BuyingPower, err = strconv.ParseFloat(balance.BuyingPower, 64); err != nil {
				log.Error().Err(err).Msg("error converting BuyingPower to float64")
				return nil, err
			}
		}

		if balance.CashBalance != "" {
			if b.CashBalance, err = strconv.ParseFloat(balance.CashBalance, 64); err != nil {
				log.Error().Err(err).Msg("error converting CashBalance to float64")
				return nil, err
			}
		}

		if balance.Commission != "" {
			if b.Commission, err = strconv.ParseFloat(balance.Commission, 64); err != nil {
				log.Error().Err(err).Msg("error converting Commission to float64")
				return nil, err
			}
		}

		if balance.BalanceDetail.CostOfPositions != "" {
			if b.CostOfPositions, err = strconv.ParseFloat(balance.BalanceDetail.CostOfPositions, 64); err != nil {
				log.Error().Err(err).Msg("error converting CostOfPositions to float64")
				return nil, err
			}
		}

		if balance.BalanceDetail.DayTrades != "" {
			if b.DayTrades, err = strconv.ParseFloat(balance.BalanceDetail.DayTrades, 64); err != nil {
				log.Error().Err(err).Msg("error converting DayTrades to float64")
				return nil, err
			}
		}

		if balance.Equity != "" {
			if b.Equity, err = strconv.ParseFloat(balance.Equity, 64); err != nil {
				log.Error().Err(err).Msg("error converting Equity to float64")
				return nil, err
			}
		}

		if balance.BalanceDetail.MaintenanceRate != "" {
			if b.MaintenanceRate, err = strconv.ParseFloat(balance.BalanceDetail.MaintenanceRate, 64); err != nil {
				log.Error().Err(err).Msg("error converting MaintenanceRate to float64")
				return nil, err
			}
		}

		if balance.MarketValue != "" {
			if b.MarketValue, err = strconv.ParseFloat(balance.MarketValue, 64); err != nil {
				log.Error().Err(err).Msg("error converting MarketValue to float64")
				return nil, err
			}
		}

		if balance.BalanceDetail.OvernightBuyingPower != "" {
			if b.OvernightBuyingPower, err = strconv.ParseFloat(balance.BalanceDetail.OvernightBuyingPower, 64); err != nil {
				log.Error().Err(err).Msg("error converting OvernightBuyingPower to float64")
				return nil, err
			}
		}

		if balance.BalanceDetail.RealizedProfitLoss != "" {
			if b.RealizedProfitLoss, err = strconv.ParseFloat(balance.BalanceDetail.RealizedProfitLoss, 64); err != nil {
				log.Error().Err(err).Msg("error converting RealizedProfitLoss to float64")
				return nil, err
			}
		}

		if balance.BalanceDetail.RequiredMargin != "" {
			if b.RequiredMargin, err = strconv.ParseFloat(balance.BalanceDetail.RequiredMargin, 64); err != nil {
				log.Error().Err(err).Msg("error converting RequiredMargin to float64")
				return nil, err
			}
		}

		if balance.TodaysProfitLoss != "" {
			if b.TodaysProfitLoss, err = strconv.ParseFloat(balance.TodaysProfitLoss, 64); err != nil {
				log.Error().Err(err).Msg("error converting TodaysProfitLoss to float64")
				return nil, err
			}
		}

		if balance.UnclearedDeposit != "" {
			if b.UnclearedDeposit, err = strconv.ParseFloat(balance.UnclearedDeposit, 64); err != nil {
				log.Error().Err(err).Msg("error converting UnclearedDeposit to float64")
				return nil, err
			}
		}

		if balance.BalanceDetail.UnrealizedProfitLoss != "" {
			if b.UnrealizedProfitLoss, err = strconv.ParseFloat(balance.BalanceDetail.UnrealizedProfitLoss, 64); err != nil {
				log.Error().Err(err).Msg("error converting UnrealizedProfitLoss to float64")
				return nil, err
			}
		}

		resBalance[idx] = b
	}

	return resBalance[0], nil
}

func (account *Account) ordersRequest(url string, nextToken string) (*orderResponse, error) {
	account.api.CheckAuth()

	orders := orderResponse{
		Orders: make([]*tsOrder, 0, 100),
		Errors: make([]*tsError, 0, 1),
	}

	if nextToken != "" {
		url = fmt.Sprintf("%s&nextToken=%s", url, nextToken)
	}

	resp, err := account.api.client.R().
		SetResult(&orders).
		Get(url)
	if err != nil {
		log.Error().Err(err).Msg("account request failed")
		return nil, err
	}
	if resp.StatusCode() >= 400 {
		log.Error().Int("StatusCode", resp.StatusCode()).Msg("Received invalid status code")
		return nil, fmt.Errorf("/brokerage/accounts %d", resp.StatusCode())
	}
	if len(orders.Errors) != 0 {
		errorMsgs := make([]string, len(orders.Errors))
		for idx, errMsg := range orders.Errors {
			errorMsgs[idx] = errMsg.Message
		}
		log.Error().Strs("Errors", errorMsgs).Msg("errors returned by tradestation api")
		return nil, errors.New("tradestation api returned errors")
	}

	return &orders, nil
}

func convertOrders(orders []*tsOrder) ([]*Order, error) {
	var err error
	nyc, err := time.LoadLocation("America/New_York")
	if err != nil {
		return nil, err
	}
	res := make([]*Order, len(orders))
	for idx, order := range orders {
		o := &Order{
			AccountID:         order.AccountID,
			AdvancedOptions:   order.AdvancedOptions,
			ConditionalOrders: order.ConditionalOrders,
			Duration:          order.Duration,
			GroupName:         order.GroupName,
			OrderID:           order.OrderID,
			OrderType:         order.OrderType,
			RejectReason:      order.RejectReason,
			Routing:           order.Routing,
			StatusDescription: order.StatusDescription,
		}

		if order.ClosedDateTime != "" {
			if o.ClosedDateTime, err = time.Parse("2006-01-02T15:04:05Z", order.ClosedDateTime); err != nil {
				log.Error().Err(err).Msg("error converting ClosedDateTime to time")
				return nil, err
			}
			o.ClosedDateTime = o.ClosedDateTime.In(nyc)
		}

		if order.CommissionFee != "" {
			if o.CommissionFee, err = strconv.ParseFloat(order.CommissionFee, 64); err != nil {
				log.Error().Err(err).Msg("error converting CommissionFee to float64")
				return nil, err
			}
		}

		if order.FilledPrice != "" {
			if o.FilledPrice, err = strconv.ParseFloat(order.FilledPrice, 64); err != nil {
				log.Error().Err(err).Msg("error converting FilledPrice to float64")
				return nil, err
			}
		}

		if order.GoodTillDate != "" {
			if o.GoodTillDate, err = time.Parse("2006-01-02T15:04:05Z", order.GoodTillDate); err != nil {
				log.Error().Err(err).Msg("error converting GoodTillDate to time")
				return nil, err
			}
			o.GoodTillDate = o.GoodTillDate.In(nyc)
		}

		o.Legs = make([]*OrderLeg, len(order.Legs))
		for ii, leg := range order.Legs {
			l := &OrderLeg{
				OpenOrClose: leg.OpenOrClose,
				BuyOrSell:   leg.BuyOrSell,
				Symbol:      leg.Symbol,
				AssetType:   leg.AssetType,
			}

			if leg.QuantityOrdered != "" {
				if l.QuantityOrdered, err = strconv.ParseInt(leg.QuantityOrdered, 0, 64); err != nil {
					log.Error().Err(err).Msg("error converting QuantityOrdered to float64")
					return nil, err
				}
			}

			if leg.ExecQuantity != "" {
				if l.ExecQuantity, err = strconv.ParseInt(leg.ExecQuantity, 0, 64); err != nil {
					log.Error().Err(err).Msg("error converting ExecQuantity to float64")
					return nil, err
				}
			}

			if leg.QuantityRemaining != "" {
				if l.QuantityRemaining, err = strconv.ParseInt(leg.QuantityRemaining, 0, 64); err != nil {
					log.Error().Err(err).Msg("error converting QuantityRemaining to float64")
					return nil, err
				}
			}

			o.Legs[ii] = l
		}

		o.MarketActivationRules = make([]*MarketRule, len(order.MarketActivationRules))
		for ii, rule := range order.MarketActivationRules {
			r := &MarketRule{
				RuleType:   MarketRuleType(rule.RuleType),
				Symbol:     rule.Symbol,
				Predicate:  MarketRulePredicate(rule.Predicate),
				TriggerKey: MarketRuleTrigger(rule.TriggerKey),
			}

			if rule.Price != "" {
				if r.Price, err = strconv.ParseFloat(rule.Price, 64); err != nil {
					log.Error().Err(err).Msg("error converting Price to float64")
					return nil, err
				}
			}

			o.MarketActivationRules[ii] = r
		}

		if order.OpenedDateTime != "" {
			if o.OpenedDateTime, err = time.Parse("2006-01-02T15:04:05Z", order.OpenedDateTime); err != nil {
				log.Error().Err(err).Msg("error converting OpenedDateTime to time")
				return nil, err
			}
			o.OpenedDateTime = o.OpenedDateTime.In(nyc)
		}

		if order.PriceUsedForBuyingPower != "" {
			if o.PriceUsedForBuyingPower, err = strconv.ParseFloat(order.PriceUsedForBuyingPower, 64); err != nil {
				log.Error().Err(err).Msg("error converting PriceUsedForBuyingPower to float64")
				return nil, err
			}
		}

		o.Status = OrderStatus(order.Status)

		o.TimeActivationRules = make([]*TimeRule, len(order.TimeActivationRules))
		for ii, rule := range order.TimeActivationRules {
			myTime, err := time.Parse("2006-01-02T15:04:05Z", rule.TimeUtc)
			if err != nil {
				log.Error().Err(err).Msg("error converting time activation rule")
			}
			t := &TimeRule{
				TimeUtc: myTime,
			}
			o.TimeActivationRules[ii] = t
		}

		if order.UnbundledRouteFee != "" {
			if o.UnbundledRouteFee, err = strconv.ParseFloat(order.UnbundledRouteFee, 64); err != nil {
				log.Error().Err(err).Msg("error converting UnbundledRouteFee to float64")
				return nil, err
			}
		}

		res[idx] = o
	}

	return res, nil
}

// GetHistoricalOrders retrieves historical orders from tradestation
//
// since is the earliest date to retrieve orders for
func (account *Account) GetHistoricalOrders(since time.Time) ([]*Order, error) {
	allOrders := make([]*tsOrder, 0, 100)
	url := fmt.Sprintf("/brokerage/accounts/%s/historicalorders?since=%s", account.AccountID, since.Format("2006-01-02"))
	orders, err := account.ordersRequest(url, "")
	if err != nil {
		return nil, err
	}
	allOrders = append(allOrders, orders.Orders...)

	for orders.NextToken != "" {
		orders, err := account.ordersRequest(url, orders.NextToken)
		if err != nil {
			return nil, err
		}
		allOrders = append(allOrders, orders.Orders...)
	}

	return convertOrders(allOrders)
}

// GetOrders retrieves todays orders from tradestation
func (account *Account) GetOrders() ([]*Order, error) {
	allOrders := make([]*tsOrder, 0, 100)
	url := fmt.Sprintf("/brokerage/accounts/%s/orders", account.AccountID)
	orders, err := account.ordersRequest(url, "")
	if err != nil {
		return nil, err
	}
	allOrders = append(allOrders, orders.Orders...)

	for orders.NextToken != "" {
		orders, err := account.ordersRequest(url, orders.NextToken)
		if err != nil {
			return nil, err
		}
		allOrders = append(allOrders, orders.Orders...)
	}
	return convertOrders(allOrders)
}

func (account *Account) GetPositions() ([]*Position, error) {
	account.api.CheckAuth()
	nyc, err := time.LoadLocation("America/New_York")
	if err != nil {
		return nil, err
	}
	positions := positionResponse{
		Positions: make([]*tsPosition, 0, 5),
		Errors:    make([]*tsError, 0, 1),
	}
	resp, err := account.api.client.R().
		SetResult(&positions).
		Get(fmt.Sprintf("/brokerage/accounts/%s/positions", account.AccountID))
	if err != nil {
		log.Error().Err(err).Msg("account request failed")
		return nil, err
	}
	if resp.StatusCode() >= 400 {
		return nil, fmt.Errorf("/brokerage/accounts %d", resp.StatusCode())
	}
	if len(positions.Errors) != 0 {
		errorMsgs := make([]string, len(positions.Errors))
		for idx, errMsg := range positions.Errors {
			errorMsgs[idx] = errMsg.Message
		}
		log.Error().Strs("Errors", errorMsgs).Msg("errors returned by tradestation api")
		return nil, errors.New("tradestation api returned errors")
	}

	// convert positions to native types
	pos := make([]*Position, len(positions.Positions))
	for idx, position := range positions.Positions {
		p := &Position{
			AccountID:  position.AccountID,
			AssetType:  position.AssetType,
			PositionID: position.PositionID,
			LongShort:  position.LongShort,
			Symbol:     position.Symbol,
		}

		if position.AveragePrice != "" {
			if p.AveragePrice, err = strconv.ParseFloat(position.AveragePrice, 64); err != nil {
				log.Error().Err(err).Msg("error converting AveragePrice to float64")
				return nil, err
			}
		}

		if position.Last != "" {
			if p.Last, err = strconv.ParseFloat(position.Last, 64); err != nil {
				log.Error().Err(err).Msg("error converting Last to float64")
				return nil, err
			}
		}

		if position.Bid != "" {
			if p.Bid, err = strconv.ParseFloat(position.Bid, 64); err != nil {
				log.Error().Err(err).Msg("error converting Bid to float64")
				return nil, err
			}
		}

		if position.Ask != "" {
			if p.Ask, err = strconv.ParseFloat(position.Ask, 64); err != nil {
				log.Error().Err(err).Msg("error converting Ask to float64")
				return nil, err
			}
		}

		if position.Quantity != "" {
			if p.Quantity, err = strconv.ParseInt(position.Quantity, 0, 64); err != nil {
				log.Error().Err(err).Msg("error converting Ask to float64")
				return nil, err
			}
		}

		if position.Timestamp != "" {
			if p.Timestamp, err = time.Parse("2006-01-02T15:04:05Z", position.Timestamp); err != nil {
				log.Error().Err(err).Msg("error converting Timestamp to time")
				return nil, err
			}
			p.Timestamp = p.Timestamp.In(nyc)
		}

		if position.TodaysProfitLoss != "" {
			if p.TodaysProfitLoss, err = strconv.ParseFloat(position.TodaysProfitLoss, 64); err != nil {
				log.Error().Err(err).Msg("error converting TodaysProfitLoss to float64")
				return nil, err
			}
		}

		if position.TotalCost != "" {
			if p.TotalCost, err = strconv.ParseFloat(position.TotalCost, 64); err != nil {
				log.Error().Err(err).Msg("error converting TotalCost to float64")
				return nil, err
			}
		}

		if position.MarketValue != "" {
			if p.MarketValue, err = strconv.ParseFloat(position.MarketValue, 64); err != nil {
				log.Error().Err(err).Msg("error converting MarketValue to float64")
				return nil, err
			}
		}

		if position.MarkToMarketPrice != "" {
			if p.MarkToMarketPrice, err = strconv.ParseFloat(position.MarkToMarketPrice, 64); err != nil {
				log.Error().Err(err).Msg("error converting MarkToMarketPrice to float64")
				return nil, err
			}
		}

		if position.UnrealizedProfitLoss != "" {
			if p.UnrealizedProfitLoss, err = strconv.ParseFloat(position.UnrealizedProfitLoss, 64); err != nil {
				log.Error().Err(err).Msg("error converting UnrealizedProfitLoss to float64")
				return nil, err
			}
		}

		if position.UnrealizedProfitLossPercent != "" {
			if p.UnrealizedProfitLossPercent, err = strconv.ParseFloat(position.UnrealizedProfitLossPercent, 64); err != nil {
				log.Error().Err(err).Msg("error converting UnrealizedProfitLossPercent to float64")
				return nil, err
			}
		}

		if position.UnrealizedProfitLossQty != "" {
			if p.UnrealizedProfitLossQty, err = strconv.ParseFloat(position.UnrealizedProfitLossQty, 64); err != nil {
				log.Error().Err(err).Msg("error converting UnrealizedProfitLossQty to float64")
				return nil, err
			}
		}

		pos[idx] = p
	}

	return pos, nil
}
