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
	"time"

	"github.com/penny-vault/tradestation/tradestation"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

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
		fmt.Println("# Accounts")
		for _, account := range accounts {
			if account.Alias != "" {
				fmt.Printf("## %s (%s)\n", account.Alias, account.AccountID)
			} else {
				fmt.Printf("## %s (%s)\n", account.AccountID, account.AccountType)
			}

			// Get orders for the last 30 days
			orders, err := account.GetHistoricalOrders(time.Now().AddDate(0, 0, -30))
			if err != nil {
				log.Error().Err(err).Str("AccountID", account.AccountID).Msg("error getting orders for account")
			}
			for ii, order := range orders {
				fmt.Printf(" %d. %s %s %s\n", ii+1, order.OrderID, order.StatusDescription, order.Legs[0].Symbol)
			}

			fmt.Printf("\n")
		}
	},
}

func init() {
	rootCmd.AddCommand(trxCmd)
}
