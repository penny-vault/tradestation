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
	Orders    []*Order
	Errors    []*tsError
	NextToken string
}

type OrderLeg struct {
	OpenOrClose       string
	QuantityOrdered   float64
	ExecQuantity      float64
	QuantityRemaining float64
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

type MarketRule struct {
	RuleType   string
	Symbol     string
	Predicate  string
	TriggerKey string
	Price      float64
}

type TimeRule struct {
	TimeUtc time.Time
}

type Order struct {
	AccountID               string
	AdvancedOptions         string
	ClosedDateTime          float64
	CommissionFee           float64
	Duration                string
	FilledPrice             string
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
	Quantity                    float64
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

func (api *API) GetAccounts() ([]*Account, error) {
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

	return accounts.Accounts, nil
}

func (account *Account) GetBalances() (*Balance, error) {
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

		if b.BuyingPower, err = strconv.ParseFloat(balance.BuyingPower, 64); err != nil {
			log.Error().Err(err).Msg("error converting BuyingPower to float64")
			return nil, err
		}

		if b.CashBalance, err = strconv.ParseFloat(balance.CashBalance, 64); err != nil {
			log.Error().Err(err).Msg("error converting CashBalance to float64")
			return nil, err
		}

		if b.Commission, err = strconv.ParseFloat(balance.Commission, 64); err != nil {
			log.Error().Err(err).Msg("error converting Commission to float64")
			return nil, err
		}

		if b.CostOfPositions, err = strconv.ParseFloat(balance.BalanceDetail.CostOfPositions, 64); err != nil {
			log.Error().Err(err).Msg("error converting CostOfPositions to float64")
			return nil, err
		}

		if b.DayTrades, err = strconv.ParseFloat(balance.BalanceDetail.DayTrades, 64); err != nil {
			log.Error().Err(err).Msg("error converting DayTrades to float64")
			return nil, err
		}

		if b.Equity, err = strconv.ParseFloat(balance.Equity, 64); err != nil {
			log.Error().Err(err).Msg("error converting Equity to float64")
			return nil, err
		}

		if b.MaintenanceRate, err = strconv.ParseFloat(balance.BalanceDetail.MaintenanceRate, 64); err != nil {
			log.Error().Err(err).Msg("error converting MaintenanceRate to float64")
			return nil, err
		}

		if b.MarketValue, err = strconv.ParseFloat(balance.MarketValue, 64); err != nil {
			log.Error().Err(err).Msg("error converting MarketValue to float64")
			return nil, err
		}

		if b.OvernightBuyingPower, err = strconv.ParseFloat(balance.BalanceDetail.OvernightBuyingPower, 64); err != nil {
			log.Error().Err(err).Msg("error converting OvernightBuyingPower to float64")
			return nil, err
		}

		if b.RealizedProfitLoss, err = strconv.ParseFloat(balance.BalanceDetail.RealizedProfitLoss, 64); err != nil {
			log.Error().Err(err).Msg("error converting RealizedProfitLoss to float64")
			return nil, err
		}

		if b.RequiredMargin, err = strconv.ParseFloat(balance.BalanceDetail.RequiredMargin, 64); err != nil {
			log.Error().Err(err).Msg("error converting RequiredMargin to float64")
			return nil, err
		}

		if b.TodaysProfitLoss, err = strconv.ParseFloat(balance.TodaysProfitLoss, 64); err != nil {
			log.Error().Err(err).Msg("error converting TodaysProfitLoss to float64")
			return nil, err
		}

		if b.UnclearedDeposit, err = strconv.ParseFloat(balance.UnclearedDeposit, 64); err != nil {
			log.Error().Err(err).Msg("error converting UnclearedDeposit to float64")
			return nil, err
		}

		if b.UnrealizedProfitLoss, err = strconv.ParseFloat(balance.BalanceDetail.UnrealizedProfitLoss, 64); err != nil {
			log.Error().Err(err).Msg("error converting UnrealizedProfitLoss to float64")
			return nil, err
		}

		resBalance[idx] = b
	}

	return resBalance[0], nil
}

func (account *Account) GetOrders() ([]*Order, error) {
	accounts := accountResponse{
		Accounts: make([]*Account, 0, 5),
	}
	resp, err := account.api.client.R().
		SetResult(&accounts).
		Get("/brokerage/accounts")
	if err != nil {
		log.Error().Err(err).Msg("account request failed")
		return nil, err
	}
	if resp.StatusCode() >= 400 {
		return nil, fmt.Errorf("/brokerage/accounts %d", resp.StatusCode())
	}

	return nil, nil
}

func (account *Account) GetPositions() ([]*Position, error) {
	accounts := accountResponse{
		Accounts: make([]*Account, 0, 5),
	}
	resp, err := account.api.client.R().
		SetResult(&accounts).
		Get("/brokerage/accounts")
	if err != nil {
		log.Error().Err(err).Msg("account request failed")
		return nil, err
	}
	if resp.StatusCode() >= 400 {
		return nil, fmt.Errorf("/brokerage/accounts %d", resp.StatusCode())
	}

	return nil, nil
}
