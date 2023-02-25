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
	"github.com/go-resty/resty/v2"
	"github.com/spf13/viper"
)

type API struct {
	token   *OAuthToken
	baseUrl string
	client  *resty.Client
}

func New() *API {
	api := &API{
		token:   nil,
		baseUrl: viper.GetString("sim"),
		client:  resty.New(),
	}
	if viper.GetString("mode") == "live" {
		api.baseUrl = viper.GetString("live")
	}
	api.client = api.client.SetBaseURL(api.baseUrl)
	api.client.SetDebug(viper.GetBool("debug"))
	return api
}
