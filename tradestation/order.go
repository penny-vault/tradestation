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

type Action string

const (
	BUY        Action = "BUY"
	SELL       Action = "SELL"
	BUYTOCOVER Action = "BUYTOCOVER"
	SELLSHORT  Action = "SELLSHORT"
)

type TimeInForceDuration string

const (
	DAY                      TimeInForceDuration = "DAY"
	DAY_PLUS                 TimeInForceDuration = "DYP"
	GOOD_TILL_CANCELLED      TimeInForceDuration = "GTC"
	GOOD_TILL_CANCELLED_PLUS TimeInForceDuration = "GCP"
	GOOD_THROUGH_DATE        TimeInForceDuration = "GDP"
	ON_OPEN                  TimeInForceDuration = "OPG"
	ON_CLOSE                 TimeInForceDuration = "CLO"
	IMMEDIATE                TimeInForceDuration = "IOC"
	FILL_OR_KILL             TimeInForceDuration = "FOK"
	ONE_MIN                  TimeInForceDuration = "1"
	THREE_MIN                TimeInForceDuration = "3"
	FIVE_MIN                 TimeInForceDuration = "5"
)

type tsTimeInForce struct {
	Duration   TimeInForceDuration
	Expiration string `json:"Expiration,omitempty"`
}

type tsOrderConfirmLeg struct {
	ExpirationDate string `json:"ExpirationDate,omitempty"`
	Quantity       string
	Symbol         string
	TradeAction    Action
}

type tsOrderConfirm struct {
	AccountCurrency          string
	AccountID                string
	BaseCurrency             string
	CounterCurrency          string
	Currency                 string
	DebitCreditEstimatedCost string
	EstimatedCommission      string
	EstimatedCost            string
	EstimatedPrice           string
	InitialMarginDisplay     string
	Legs                     []*tsOrderConfirmLeg
	LimitPrice               string
	OrderAssetCategory       string
	OrderConfirmID           string
	ProductCurrency          string
	Route                    string
	StopPrice                string
	SummaryMessage           string
	TimeInForce              tsTimeInForce
}

type OrderConfirm struct {
	AccountCurrency          string
	AccountID                string
	BaseCurrency             string
	CounterCurrency          string
	Currency                 string
	DebitCreditEstimatedCost float64
	EstimatedCommission      float64
	EstimatedCost            float64
	EstimatedPrice           float64
	InitialMarginDisplay     string
	ExpirationDate           time.Time
	Quantity                 int64
	Symbol                   string
	TradeAction              Action
	LimitPrice               float64
	OrderAssetCategory       string
	OrderConfirmID           string
	ProductCurrency          string
	Route                    string
	StopPrice                float64
	SummaryMessage           string
	TimeInForceDur           TimeInForceDuration
	TimeInForceExpiration    time.Time
}

type TSOrderType string

const (
	LIMIT      TSOrderType = "Limit"
	STOP       TSOrderType = "StopMarket"
	MARKET     TSOrderType = "Market"
	STOP_LIMIT TSOrderType = "StopLimit"
)

type tsOrderRequest struct {
	AccountID      string
	LimitPrice     string `json:"LimitPrice,omitempty"`
	OrderConfirmID string `json:"OrderConfirmID,omitempty"`
	OrderType      TSOrderType
	Quantity       string
	StopPrice      string `json:"StopPrice,omitempty"`
	Symbol         string
	TimeInForce    tsTimeInForce
	TradeAction    Action
}

type OrderRequest struct {
	AccountID      string
	LimitPrice     float64
	OrderConfirmID string
	OrderType      TSOrderType
	Quantity       int64
	StopPrice      float64
	Symbol         string
	TimeInForceDur TimeInForceDuration
	TradeAction    Action
}

type confirmOrderResponse struct {
	Confirmations []*tsOrderConfirm
}

func (req *OrderRequest) toTsOrderRequest() *tsOrderRequest {
	return &tsOrderRequest{
		AccountID:      req.AccountID,
		LimitPrice:     fmt.Sprintf("%.2f", req.LimitPrice),
		OrderConfirmID: req.OrderConfirmID,
		OrderType:      req.OrderType,
		Quantity:       fmt.Sprintf("%d", req.Quantity),
		StopPrice:      fmt.Sprintf("%.2f", req.StopPrice),
		Symbol:         req.Symbol,
		TimeInForce: tsTimeInForce{
			Duration: req.TimeInForceDur,
		},
		TradeAction: req.TradeAction,
	}
}

func convertOrderConfirm(order *tsOrderConfirm) (*OrderConfirm, error) {
	var err error
	confirm := &OrderConfirm{
		AccountCurrency:      order.AccountCurrency,
		AccountID:            order.AccountID,
		BaseCurrency:         order.BaseCurrency,
		CounterCurrency:      order.CounterCurrency,
		Currency:             order.Currency,
		InitialMarginDisplay: order.InitialMarginDisplay,
		OrderAssetCategory:   order.OrderAssetCategory,
		OrderConfirmID:       order.OrderConfirmID,
		ProductCurrency:      order.ProductCurrency,
		Route:                order.Route,
		SummaryMessage:       order.SummaryMessage,
		TimeInForceDur:       order.TimeInForce.Duration,
	}

	if order.DebitCreditEstimatedCost != "" {
		if confirm.DebitCreditEstimatedCost, err = strconv.ParseFloat(order.DebitCreditEstimatedCost, 64); err != nil {
			log.Error().Err(err).Msg("error converting DebitCreditEstimatedCost to float64")
			return nil, err
		}
	}

	if order.EstimatedCommission != "" {
		if confirm.EstimatedCommission, err = strconv.ParseFloat(order.EstimatedCommission, 64); err != nil {
			log.Error().Err(err).Msg("error converting EstimatedCommission to float64")
			return nil, err
		}
	}

	if order.EstimatedCost != "" {
		if confirm.EstimatedCost, err = strconv.ParseFloat(order.EstimatedCost, 64); err != nil {
			log.Error().Err(err).Msg("error converting EstimatedCost to float64")
			return nil, err
		}
	}

	if order.EstimatedPrice != "" {
		if confirm.EstimatedPrice, err = strconv.ParseFloat(order.EstimatedPrice, 64); err != nil {
			log.Error().Err(err).Msg("error converting EstimatedPrice to float64")
			return nil, err
		}
	}

	if order.LimitPrice != "" {
		if confirm.LimitPrice, err = strconv.ParseFloat(order.LimitPrice, 64); err != nil {
			log.Error().Err(err).Msg("error converting LimitPrice to float64")
			return nil, err
		}
	}

	if order.StopPrice != "" {
		if confirm.StopPrice, err = strconv.ParseFloat(order.StopPrice, 64); err != nil {
			log.Error().Err(err).Msg("error converting StopPrice to float64")
			return nil, err
		}
	}

	if order.TimeInForce.Expiration != "" {
		if confirm.TimeInForceExpiration, err = time.Parse("2006-01-02T15:04:05Z", order.TimeInForce.Expiration); err != nil {
			log.Error().Err(err).Msg("error converting ExpirationDate to time")
			return nil, err
		}
	}

	switch len(order.Legs) {
	case 0:
	case 1:
		l := order.Legs[0]
		confirm.Symbol = l.Symbol
		confirm.TradeAction = l.TradeAction

		if l.ExpirationDate != "" {
			if confirm.ExpirationDate, err = time.Parse("2006-01-02T15:04:05Z", l.ExpirationDate); err != nil {
				log.Error().Err(err).Msg("error converting ExpirationDate to time")
				return nil, err
			}
		}

		if l.Quantity != "" {
			if confirm.Quantity, err = strconv.ParseInt(l.Quantity, 0, 64); err != nil {
				log.Error().Err(err).Msg("error converting Quantity to int64")
				return nil, err
			}
		}
	default:
		log.Error().Int("Len OrderLegs", len(order.Legs)).Msg("order legs unexpected size, should be 1")
	}

	return confirm, nil
}

// ConfirmOrder returns estimated cost and commission information for an order
// without the order actually being placed. Request valid for Market, Limit,
// Stop Market, Stop Limit, Options, and Order Sends Order (OSO) order types.
func (account *Account) ConfirmOrder(order *OrderRequest) (*OrderConfirm, error) {
	account.api.CheckAuth()

	confirms := confirmOrderResponse{
		Confirmations: make([]*tsOrderConfirm, 0, 1),
	}

	tsOrder := order.toTsOrderRequest()
	tsOrder.AccountID = account.AccountID

	resp, err := account.api.client.R().
		SetBody(tsOrder).
		SetResult(&confirms).
		Post("/orderexecution/orderconfirm")
	if err != nil {
		log.Error().Err(err).Msg("account request failed")
		return nil, err
	}
	if resp.StatusCode() >= 400 {
		log.Error().Int("StatusCode", resp.StatusCode()).Str("Body", string(resp.Body())).Msg("Received invalid status code")
		return nil, fmt.Errorf("%s %d", resp.Request.URL, resp.StatusCode())
	}

	// convert to OrderConfirm object
	return convertOrderConfirm(confirms.Confirmations[0])
}

// ConfirmGroupOrder returns estimated cost and commission information for a group of
// orders without the orders actually being placed. Request valid for Market,
// Limit, Stop Market, Stop Limit, Options, and Order Sends Order (OSO) order
// types.
func (account *Account) ConfirmGroupOrder(orders []*OrderRequest) ([]*OrderConfirm, error) {
	account.api.CheckAuth()

	confirms := confirmOrderResponse{
		Confirmations: make([]*tsOrderConfirm, 0, len(orders)),
	}

	tsOrders := make([]*tsOrderRequest, len(orders))
	for idx, order := range orders {
		tsOrders[idx] = order.toTsOrderRequest()
		tsOrders[idx].AccountID = account.AccountID
	}

	resp, err := account.api.client.R().
		SetBody(map[string]any{
			"Orders": tsOrders,
			"Type":   "NORMAL",
		}).
		SetResult(&confirms).
		Post("/orderexecution/ordergroupconfirm")
	if err != nil {
		log.Error().Err(err).Msg("account request failed")
		return nil, err
	}
	if resp.StatusCode() >= 400 {
		log.Error().Int("StatusCode", resp.StatusCode()).Msg("Received invalid status code")
		return nil, fmt.Errorf("%s %d", resp.Request.URL, resp.StatusCode())
	}

	res := make([]*OrderConfirm, len(confirms.Confirmations))
	for idx, confirm := range confirms.Confirmations {
		c, err := convertOrderConfirm(confirm)
		if err != nil {
			return nil, err
		}
		res[idx] = c
	}
	return res, nil
}

// Creates a new brokerage order. Request valid for all account types. Request
// valid for Market, Limit, Stop Market, Stop Limit, Options and Order Sends
// Order (OSO) order types.
func (account *Account) PlaceOrder(order *OrderRequest) (*Order, error) {
	account.api.CheckAuth()

	orderResp := orderResponse{
		Errors: make([]*tsError, 0, 1),
		Orders: make([]*tsOrder, 0, 1),
	}

	tsOrder := order.toTsOrderRequest()
	tsOrder.AccountID = account.AccountID

	resp, err := account.api.client.R().
		SetBody(tsOrder).
		SetResult(&orderResp).
		Post("/orderexecution/orders")
	if err != nil {
		log.Error().Err(err).Msg("account request failed")
		return nil, err
	}
	if resp.StatusCode() >= 400 {
		log.Error().Int("StatusCode", resp.StatusCode()).Msg("Received invalid status code")
		return nil, fmt.Errorf("%s %d", resp.Request.URL, resp.StatusCode())
	}
	if len(orderResp.Errors) > 0 {
		for _, err := range orderResp.Errors {
			log.Error().Str("ErrorType", err.Error).Msg(err.Message)
		}
		return nil, errors.New("received errors in place order group")
	}

	res, err := convertOrders(orderResp.Orders)
	return res[0], err
}

// Creates a new brokerage order. Request valid for all account types. Request
// valid for Market, Limit, Stop Market, Stop Limit, Options and Order Sends
// Order (OSO) order types.
func (account *Account) PlaceGroupOrder(orders []*OrderRequest) ([]*Order, error) {
	account.api.CheckAuth()

	orderResp := orderResponse{
		Errors: make([]*tsError, 0, 1),
		Orders: make([]*tsOrder, 0, len(orders)),
	}

	tsOrders := make([]*tsOrderRequest, len(orders))
	for idx, order := range orders {
		tsOrders[idx] = order.toTsOrderRequest()
		tsOrders[idx].AccountID = account.AccountID
	}

	resp, err := account.api.client.R().
		SetBody(map[string]any{
			"Orders": tsOrders,
			"Type":   "NORMAL",
		}).
		SetResult(&orderResp).
		Post("/orderexecution/ordergroups")
	if err != nil {
		log.Error().Err(err).Msg("account request failed")
		return nil, err
	}
	if resp.StatusCode() >= 400 {
		log.Error().Int("StatusCode", resp.StatusCode()).Msg("Received invalid status code")
		return nil, fmt.Errorf("%s %d", resp.Request.URL, resp.StatusCode())
	}
	if len(orderResp.Errors) > 0 {
		for _, err := range orderResp.Errors {
			log.Error().Str("ErrorType", err.Error).Msg(err.Message)
		}
		return nil, errors.New("received errors in place order group")
	}

	return convertOrders(orderResp.Orders)
}
