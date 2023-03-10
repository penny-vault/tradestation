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

// balanceCmd represents the balance command
var balanceCmd = &cobra.Command{
	Use:   "balance",
	Short: "Get current ticker balances",
	Run: func(cmd *cobra.Command, args []string) {
		api := tradestation.New()

		accounts, err := api.GetAccounts()
		if err != nil {
			log.Error().Err(err).Msg("Error getting accounts")
		}

		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"Account ID", "Account Type", "Cash Balance", "Buying Power"})
		table.SetBorder(false) // Set Border to false

		for _, account := range accounts {
			// Get account balances
			balance, err := account.GetBalances()
			if err != nil {
				log.Error().Err(err).Str("AccountID", account.AccountID).Msg("error getting orders for account")
				continue
			}

			row := []string{account.AccountID, account.AccountType, fmt.Sprintf("%.2f", balance.CashBalance), fmt.Sprintf("%.2f", balance.BuyingPower)}
			table.Append(row)
		}

		table.Render()
		fmt.Printf("\n")
	},
}

func init() {
	rootCmd.AddCommand(balanceCmd)
}
