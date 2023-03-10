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
	"os"

	"github.com/pelletier/go-toml/v2"
	"github.com/penny-vault/tradestation/pvts"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var confirm bool

// syncCmd represents the sync command
var syncCmd = &cobra.Command{
	Use:   "sync <sync config>",
	Short: "Execute trades according to the specified strategy in PV-API",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		// load sync file
		tl := &pvts.TradeLink{}
		data, err := os.ReadFile(args[0])
		if err != nil {
			log.Error().Err(err).Str("Sync Config", args[0]).Msg("could not read sync config file")
			return
		}
		if err := toml.Unmarshal(data, tl); err != nil {
			log.Error().Err(err).Str("Sync Config", args[0]).Msg("could not parse sync config file")
			return
		}

		if err := tl.Sync(confirm); err != nil {
			log.Error().Err(err).Str("Sync Config", args[0]).Msg("could not sync tradestation account with pv api")
			return
		}
	},
}

func init() {
	rootCmd.AddCommand(syncCmd)
	syncCmd.PersistentFlags().BoolVarP(&confirm, "confirm-yes", "y", false, "Auto-confirm all prompts during sync process")
}
