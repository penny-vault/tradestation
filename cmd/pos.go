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

// posCmd represents the pos command
var posCmd = &cobra.Command{
	Use:   "pos",
	Short: "Download positions from tradestation",
	Run: func(cmd *cobra.Command, args []string) {
		api := tradestation.New()
		accounts, err := api.GetAccounts()
		if err != nil {
			log.Error().Err(err).Msg("Error getting accounts")
		}

		for _, account := range accounts {
			if account.Alias != "" {
				fmt.Printf("# Positions for account %s (%s)\n\n", account.Alias, account.AccountID)
			} else {
				fmt.Printf("# Positions for account %s (%s)\n\n", account.AccountID, account.AccountType)
			}

			// Get account positions
			positions, err := account.GetPositions()
			if err != nil {
				log.Error().Err(err).Str("AccountID", account.AccountID).Msg("error getting orders for account")
				continue
			}

			table := tablewriter.NewWriter(os.Stdout)
			table.SetHeader([]string{"Ticker", "Pos ID", "Acquired On", "Qty", "Mkt Val", "Total Cost", "P&L", "Today's P&L"})
			table.SetBorder(false) // Set Border to false

			for _, p := range positions {
				row := []string{p.Symbol, p.PositionID, p.Timestamp.Format("2006-01-02 15:04:05 EST"), fmt.Sprintf("%d", p.Quantity), fmt.Sprintf("$%.2f", p.MarketValue), fmt.Sprintf("$%.2f", p.TotalCost), fmt.Sprintf("%.2f (%.2f%%)", p.UnrealizedProfitLoss, p.UnrealizedProfitLossPercent), fmt.Sprintf("%.2f", p.TodaysProfitLoss)}
				rowColor := tablewriter.FgGreenColor
				todaysColor := tablewriter.FgGreenColor
				if p.UnrealizedProfitLossPercent < 0 {
					rowColor = tablewriter.FgRedColor
				}
				if p.TodaysProfitLoss < 0 {
					todaysColor = tablewriter.FgRedColor
				}

				table.Rich(row, []tablewriter.Colors{{tablewriter.Normal, rowColor}, {tablewriter.Normal, rowColor}, {tablewriter.Normal, rowColor}, {tablewriter.Normal, rowColor}, {tablewriter.Normal, rowColor}, {tablewriter.Normal, rowColor}, {tablewriter.Normal, rowColor}, {tablewriter.Normal, rowColor}, {tablewriter.Normal, todaysColor}})
			}

			table.Render()
			fmt.Printf("\n")
		}
	},
}

func init() {
	rootCmd.AddCommand(posCmd)
}
