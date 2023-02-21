/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"

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
		for idx, acct := range accounts {
			fmt.Println(idx, acct)
		}
	},
}

func init() {
	rootCmd.AddCommand(acctCmd)
}
