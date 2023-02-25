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
	"strings"
	"time"

	"github.com/rs/zerolog/log"
)

type MarketFlags struct {
	IsBats         bool
	IsDelayed      bool
	IsHalted       bool
	IsHardToBorrow bool
}

type tsQuote struct {
	Ask                 string
	AskSize             string
	Bid                 string
	BidSize             string
	Close               string
	High                string
	Low                 string
	High52Week          string
	High52WeekTimestamp string
	Last                string
	MinPrice            string
	MaxPrice            string
	FirstNoticeDate     string
	LastTradingDate     string
	Low52Week           string
	Low52WeekTimestamp  string
	Flags               *MarketFlags
	NetChange           string
	NetChangePct        string
	Open                string
	PreviousClose       string
	PreviousVolume      string
	Restrictions        []string
	Symbol              string
	TickSizeTier        string
	TradeTime           string
	Volume              string
	LastSize            string
	LastVenue           string
	VWAP                string
}

type tsQuoteError struct {
	Symbol string
	Error  string
}

type quoteResponse struct {
	Quotes []*tsQuote
	Errors []*tsQuoteError
}

type Quote struct {
	Ask                 float64
	AskSize             int64
	Bid                 float64
	BidSize             int64
	Close               float64
	High                float64
	Low                 float64
	High52Week          float64
	High52WeekTimestamp time.Time
	Last                float64
	MinPrice            float64
	MaxPrice            float64
	FirstNoticeDate     time.Time
	LastTradingDate     time.Time
	Low52Week           float64
	Low52WeekTimestamp  time.Time
	Flags               *MarketFlags
	NetChange           float64
	NetChangePct        float64
	Open                float64
	PreviousClose       float64
	PreviousVolume      int64
	Restrictions        []string
	Symbol              string
	TickSizeTier        int64
	TradeTime           time.Time
	Volume              int64
	LastSize            int64
	LastVenue           string
	VWAP                float64
}

func (api *API) GetQuotes(tickers []string) ([]*Quote, error) {
	api.CheckAuth()
	nyc, err := time.LoadLocation("America/New_York")
	if err != nil {
		log.Error().Err(err).Msg("cannot load America/New_York timezone")
		return nil, err
	}

	quotes := quoteResponse{
		Quotes: make([]*tsQuote, 0, len(tickers)),
		Errors: make([]*tsQuoteError, 0, len(tickers)),
	}
	resp, err := api.client.R().
		SetResult(&quotes).
		Get(fmt.Sprintf("/marketdata/quotes/%s", strings.Join(tickers, ",")))
	if err != nil {
		log.Error().Err(err).Msg("account request failed")
		return nil, err
	}
	if resp.StatusCode() >= 400 {
		log.Error().Int("StatusCode", resp.StatusCode()).Strs("Tickers", tickers).Msg("invalid response from /marketdata/quotes")
		return nil, fmt.Errorf("%s %d", resp.Request.URL, resp.StatusCode())
	}
	if len(quotes.Errors) > 0 {
		for _, err := range quotes.Errors {
			log.Error().Str("ErrorMsg", err.Error).Str("Ticker", err.Symbol).Msg("quote request failed")
		}
		return nil, errors.New("quote download failed")
	}

	// set api on returned accounts
	res := make([]*Quote, len(quotes.Quotes))
	for idx, quote := range quotes.Quotes {
		q := &Quote{
			Flags:        quote.Flags,
			Restrictions: quote.Restrictions,
			Symbol:       quote.Symbol,
			LastVenue:    quote.LastVenue,
		}

		if quote.Ask != "" {
			if q.Ask, err = strconv.ParseFloat(quote.Ask, 64); err != nil {
				log.Error().Err(err).Msg("error converting Ask to float64")
				return nil, err
			}
		}

		if quote.AskSize != "" {
			if q.AskSize, err = strconv.ParseInt(quote.AskSize, 10, 64); err != nil {
				log.Error().Err(err).Msg("error converting AskSize to int64")
				return nil, err
			}
		}

		if quote.Bid != "" {
			if q.Bid, err = strconv.ParseFloat(quote.Bid, 64); err != nil {
				log.Error().Err(err).Msg("error converting Bid to float64")
				return nil, err
			}
		}

		if quote.BidSize != "" {
			if q.BidSize, err = strconv.ParseInt(quote.BidSize, 10, 64); err != nil {
				log.Error().Err(err).Msg("error converting BidSize to int64")
				return nil, err
			}
		}

		if quote.Close != "" {
			if q.Close, err = strconv.ParseFloat(quote.Close, 64); err != nil {
				log.Error().Err(err).Msg("error converting Close to float64")
				return nil, err
			}
		}

		if quote.High != "" {
			if q.High, err = strconv.ParseFloat(quote.High, 64); err != nil {
				log.Error().Err(err).Msg("error converting High to float64")
				return nil, err
			}
		}

		if quote.Low != "" {
			if q.Low, err = strconv.ParseFloat(quote.Low, 64); err != nil {
				log.Error().Err(err).Msg("error converting Low to float64")
				return nil, err
			}
		}

		if quote.High52Week != "" {
			if q.High52Week, err = strconv.ParseFloat(quote.High52Week, 64); err != nil {
				log.Error().Err(err).Msg("error converting High52Week to float64")
				return nil, err
			}
		}

		if quote.Last != "" {
			if q.Last, err = strconv.ParseFloat(quote.Last, 64); err != nil {
				log.Error().Err(err).Msg("error converting Last to float64")
				return nil, err
			}
		}

		if quote.MinPrice != "" {
			if q.MinPrice, err = strconv.ParseFloat(quote.MinPrice, 64); err != nil {
				log.Error().Err(err).Msg("error converting MinPrice to float64")
				return nil, err
			}
		}

		if quote.MaxPrice != "" {
			if q.MaxPrice, err = strconv.ParseFloat(quote.MaxPrice, 64); err != nil {
				log.Error().Err(err).Msg("error converting MaxPrice to float64")
				return nil, err
			}
		}

		if quote.Low52Week != "" {
			if q.Low52Week, err = strconv.ParseFloat(quote.Low52Week, 64); err != nil {
				log.Error().Err(err).Msg("error converting Low52Week to float64")
				return nil, err
			}
		}

		if quote.NetChange != "" {
			if q.NetChange, err = strconv.ParseFloat(quote.NetChange, 64); err != nil {
				log.Error().Err(err).Msg("error converting NetChange to float64")
				return nil, err
			}
		}

		if quote.NetChangePct != "" {
			if q.NetChangePct, err = strconv.ParseFloat(quote.NetChangePct, 64); err != nil {
				log.Error().Err(err).Msg("error converting NetChangePct to float64")
				return nil, err
			}
		}

		if quote.Open != "" {
			if q.Open, err = strconv.ParseFloat(quote.Open, 64); err != nil {
				log.Error().Err(err).Msg("error converting Open to float64")
				return nil, err
			}
		}

		if quote.PreviousClose != "" {
			if q.PreviousClose, err = strconv.ParseFloat(quote.PreviousClose, 64); err != nil {
				log.Error().Err(err).Msg("error converting PreviousClose to float64")
				return nil, err
			}
		}

		if quote.VWAP != "" {
			if q.VWAP, err = strconv.ParseFloat(quote.VWAP, 64); err != nil {
				log.Error().Err(err).Msg("error converting VWAP to float64")
				return nil, err
			}
		}

		if quote.PreviousVolume != "" {
			if q.PreviousVolume, err = strconv.ParseInt(quote.PreviousVolume, 10, 64); err != nil {
				log.Error().Err(err).Msg("error converting PreviousVolume to int64")
				return nil, err
			}
		}

		if quote.High52WeekTimestamp != "" {
			if q.High52WeekTimestamp, err = time.Parse("2006-01-02T15:04:05Z", quote.High52WeekTimestamp); err != nil {
				log.Error().Err(err).Msg("error converting High52WeekTimestamp to time")
				return nil, err
			}
			q.High52WeekTimestamp = q.High52WeekTimestamp.In(nyc)
		}

		if quote.FirstNoticeDate != "" {
			if q.FirstNoticeDate, err = time.Parse("2006-01-02T15:04:05Z", quote.FirstNoticeDate); err != nil {
				log.Error().Err(err).Msg("error converting FirstNoticeDate to time")
				return nil, err
			}
			q.FirstNoticeDate = q.FirstNoticeDate.In(nyc)
		}

		if quote.LastTradingDate != "" {
			if q.LastTradingDate, err = time.Parse("2006-01-02T15:04:05Z", quote.LastTradingDate); err != nil {
				log.Error().Err(err).Msg("error converting LastTradingDate to time")
				return nil, err
			}
			q.LastTradingDate = q.LastTradingDate.In(nyc)
		}

		if quote.Low52WeekTimestamp != "" {
			if q.Low52WeekTimestamp, err = time.Parse("2006-01-02T15:04:05Z", quote.Low52WeekTimestamp); err != nil {
				log.Error().Err(err).Msg("error converting Low52WeekTimestamp to time")
				return nil, err
			}
			q.Low52WeekTimestamp = q.Low52WeekTimestamp.In(nyc)
		}

		if quote.TradeTime != "" {
			if q.TradeTime, err = time.Parse("2006-01-02T15:04:05Z", quote.TradeTime); err != nil {
				log.Error().Err(err).Msg("error converting TradeTime to time")
				return nil, err
			}
			q.TradeTime = q.TradeTime.In(nyc)
		}

		res[idx] = q
	}

	return res, nil
}
