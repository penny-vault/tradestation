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

package cmd

import (
	"fmt"
	"os"

	"github.com/olekukonko/tablewriter"
	"github.com/penny-vault/tradestation/tradestation"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

// quoteCmd represents the quote command
var quoteCmd = &cobra.Command{
	Use:   "quote <ticker> ...",
	Short: "Get current ticker quotes",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		api := tradestation.New()
		quotes, err := api.GetQuotes(args)
		if err != nil {
			log.Error().Err(err).Msg("fetching quotes failed")
			return
		}

		if len(quotes) < 1 {
			return
		}

		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"Ticker", "Bid/Ask", "VWAP", "PctChange", "Open", "High", "Low", "Close", "Volume"})
		table.SetBorder(false) // Set Border to false

		for _, q := range quotes {
			row := []string{q.Symbol, fmt.Sprintf("%.2f/%.2f (%.2f)", q.Bid, q.Ask, q.Ask-q.Bid), fmt.Sprintf("%.2f", q.VWAP), fmt.Sprintf("%.2f%%", q.NetChangePct), fmt.Sprintf("$%.2f", q.Open), fmt.Sprintf("$%.2f", q.High), fmt.Sprintf("$%.2f", q.Low), fmt.Sprintf("$%.2f", q.Close), fmt.Sprintf("%d", q.Volume)}
			if q.NetChangePct < 0 {
				table.Rich(row, []tablewriter.Colors{{tablewriter.Normal, tablewriter.FgRedColor}, {tablewriter.Normal, tablewriter.FgRedColor}, {tablewriter.Normal, tablewriter.FgRedColor}, {tablewriter.Normal, tablewriter.FgRedColor}, {tablewriter.Normal, tablewriter.FgRedColor}, {tablewriter.Normal, tablewriter.FgRedColor}, {tablewriter.Normal, tablewriter.FgRedColor}, {tablewriter.Normal, tablewriter.FgRedColor}, {tablewriter.Normal, tablewriter.FgRedColor}})
			} else {
				table.Rich(row, []tablewriter.Colors{{tablewriter.Normal, tablewriter.FgGreenColor}, {tablewriter.Normal, tablewriter.FgGreenColor}, {tablewriter.Normal, tablewriter.FgGreenColor}, {tablewriter.Normal, tablewriter.FgGreenColor}, {tablewriter.Normal, tablewriter.FgGreenColor}, {tablewriter.Normal, tablewriter.FgGreenColor}, {tablewriter.Normal, tablewriter.FgGreenColor}, {tablewriter.Normal, tablewriter.FgGreenColor}, {tablewriter.Normal, tablewriter.FgGreenColor}})
			}
		}

		table.Render()
		fmt.Printf("\nAs of: %s\n", quotes[0].TradeTime.String())
	},
}

func init() {
	rootCmd.AddCommand(quoteCmd)
}
