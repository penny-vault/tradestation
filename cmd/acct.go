/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"os"

	"github.com/olekukonko/tablewriter"
	"github.com/penny-vault/tradestation/tradestation"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

// acctCmd represents the acct command
var acctCmd = &cobra.Command{
	Use:   "acct",
	Short: "Download account details",
	Run: func(cmd *cobra.Command, args []string) {
		api := tradestation.New()
		accounts, err := api.GetAccounts()
		if err != nil {
			log.Error().Err(err).Msg("account download failed")
		}

		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"Account ID", "Type", "Alias", "Status"})
		table.SetBorder(false) // Set Border to false

		for _, acct := range accounts {
			table.Append([]string{acct.AccountID, acct.AccountType, acct.Alias, acct.Status})
		}

		table.Render()
	},
}

func init() {
	rootCmd.AddCommand(acctCmd)
}
