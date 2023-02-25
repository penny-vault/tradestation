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
	"time"

	"github.com/olekukonko/tablewriter"
	"github.com/penny-vault/tradestation/tradestation"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var today bool

// trxCmd represents the trx command
var trxCmd = &cobra.Command{
	Use:   "trx",
	Short: "Download transactions from tradestation",
	Run: func(cmd *cobra.Command, args []string) {
		api := tradestation.New()
		accounts, err := api.GetAccounts()
		if err != nil {
			log.Error().Err(err).Msg("Error getting accounts")
		}

		for _, account := range accounts {
			if account.Alias != "" {
				fmt.Printf("# Orders for account %s (%s)\n", account.Alias, account.AccountID)
			} else {
				fmt.Printf("# Orders for account %s (%s)\n", account.AccountID, account.AccountType)
			}

			// Get account orders
			var orders []*tradestation.Order
			if today {
				orders, err = account.GetOrders()
				fmt.Println("showing todays transactions")
			} else {
				orders, err = account.GetHistoricalOrders(time.Now().AddDate(0, 0, -30))
			}
			if err != nil {
				log.Error().Err(err).Str("AccountID", account.AccountID).Msg("error getting orders for account")
			}

			table := tablewriter.NewWriter(os.Stdout)
			table.SetHeader([]string{"Closed Date", "Type", "Ticker", "Status", "Order ID", "Filled Price", "Exec Qty", "Req Qty", "Commission"})
			table.SetBorder(false) // Set Border to false

			for _, o := range orders {
				row := []string{o.ClosedDateTime.Format("2006-01-02 15:04:05 EST"), o.Legs[0].BuyOrSell, o.Legs[0].Symbol, o.StatusDescription, o.OrderID, fmt.Sprintf("$%.2f", o.FilledPrice), fmt.Sprintf("%d", o.Legs[0].ExecQuantity), fmt.Sprintf("%d", o.Legs[0].QuantityOrdered), fmt.Sprintf("%.2f", o.CommissionFee)}
				table.Append(row)
			}

			table.Render()
			fmt.Printf("\n")
		}
	},
}

func init() {
	rootCmd.AddCommand(trxCmd)
	trxCmd.Flags().BoolVar(&today, "today", false, "load todays orders")
}
