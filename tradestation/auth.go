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
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"net"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/hydrogen18/stoppableListener"
	"github.com/lestrrat-go/jwx/v2/jwt"
	"github.com/pkg/browser"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
	"lukechampine.com/blake3"
)

type OAuthToken struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	IDToken      string `json:"id_token"`
	TokenType    string `json:"token_type"`
	Scope        string `json:"scope"`
	ExpiresIn    int    `json:"expires_in"`
}

func EncryptAES(plaintext string) string {
	// create cipher
	key := encryptionKey()
	if len(key) != 32 {
		log.Error().Msg("encryption key invalid length")
		return ""
	}
	c, err := aes.NewCipher(key)
	if err != nil {
		log.Error().Err(err).Msg("could not create AES cipher for encryption")
	}

	// gcm or Galois/Counter Mode, is a mode of operation
	// for symmetric key cryptographic block ciphers
	// - https://en.wikipedia.org/wiki/Galois/Counter_Mode
	gcm, err := cipher.NewGCM(c)
	// if any error generating new GCM
	// handle them
	if err != nil {
		log.Error().Err(err).Msg("error creating gcm")
		return ""
	}

	// creates a new byte array the size of the nonce
	// which must be passed to Seal
	nonce := make([]byte, gcm.NonceSize())
	// populates our nonce with a cryptographically secure
	// random sequence
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		log.Error().Err(err).Msg("error populating nonce")
		return ""
	}

	// encrypt
	out := gcm.Seal(nonce, nonce, []byte(plaintext), nil)

	// return hex string
	return base64.StdEncoding.EncodeToString(out)
}

func DecryptAES(ct string) string {
	ciphertext, _ := base64.StdEncoding.DecodeString(ct)
	key := encryptionKey()
	if len(key) != 32 {
		log.Error().Msg("encryption key invalid length")
		return ""
	}
	c, err := aes.NewCipher(key)
	if err != nil {
		log.Error().Err(err).Msg("could not create AES cipher for decryption")
		return ""
	}

	gcm, err := cipher.NewGCM(c)
	if err != nil {
		log.Error().Err(err).Msg("could not create gcm")
		return ""
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		log.Error().Int("CipherTextSize", len(ciphertext)).Int("NonceSize", nonceSize).Msg("encrypted text is smaller than the nonce")
		return ""
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		log.Error().Err(err).Msg("unable to decrypt text")
	}

	return string(plaintext)
}

func ApiKey() string {
	username := viper.GetString("auth.apikey")
	return DecryptAES(username)
}

func Secret() string {
	secret := viper.GetString("auth.secret")
	return DecryptAES(secret)
}

func encryptionKey() []byte {
	encryptionKeyPath := viper.GetString("auth.key_file")
	if encryptionKeyPath == "" {
		userHomeDir, err := os.UserHomeDir()
		if err != nil {
			log.Error().Err(err).Msg("cannot get home directory")
			return []byte{}
		}
		encryptionKeyPath = fmt.Sprintf("%s/.ssh/id_rsa", userHomeDir)
	}
	key, err := os.ReadFile(encryptionKeyPath)
	if err != nil {
		log.Error().Err(err).Str("EncryptionKey", encryptionKeyPath).Msg("could not read encryption key")
		return []byte{}
	}

	key32 := blake3.Sum256(key)
	return key32[:]
}

func stateCode() string {
	var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
	b := make([]rune, 6)
	for i := range b {
		idx, err := rand.Int(rand.Reader, big.NewInt(int64(len(letters))))
		if err != nil {
			log.Panic().Err(err).Msg("could not get random digit")
		}
		b[i] = letters[idx.Int64()]
	}
	return string(b)
}

// CheckAuth checks if there is a token that is not expired available
func (api *API) CheckAuth() {
	if api.token == nil || api.token.AccessToken == "" {
		api.loadStateFile()
	}

	// if access token is still blank then authenticate
	if api.token == nil || api.token.AccessToken == "" {
		api.authenticate()
	}

	token, err := jwt.Parse([]byte(api.token.AccessToken), jwt.WithVerify(false))
	if err != nil {
		log.Error().Err(err).Msg("could not verify access token")
	}

	if token.Expiration().Before(time.Now().Add(time.Minute * 1)) {
		log.Info().Time("Expiration", token.Expiration()).Msg("refreshing access token")
		api.authenticate()
	}
}

func (api *API) loadStateFile() {
	stateFn := viper.GetString("state_file")
	state, err := os.ReadFile(stateFn)
	if err != nil {
		log.Warn().Err(err).Str("StateFile", stateFn).Msg("could not read state file")
		return
	}
	stateData := DecryptAES(string(state))
	var token OAuthToken
	err = json.Unmarshal([]byte(stateData), &token)
	if err != nil {
		log.Error().Err(err).Msg("could not unmarshal state json")
		return
	}

	api.client.SetAuthScheme("Bearer")
	api.client.SetAuthToken(token.AccessToken)

	api.token = &token
	log.Debug().Msg("loaded state from file")
}

func (api *API) writeStateFile() {
	stateFn := viper.GetString("state_file")
	data, err := json.Marshal(api.token)
	if err != nil {
		log.Error().Err(err).Msg("could not marshal token to json")
		return
	}

	encryptedData := EncryptAES(string(data))
	err = os.WriteFile(stateFn, []byte(encryptedData), 0600)
	if err != nil {
		log.Error().Err(err).Str("StateFile", stateFn).Msg("could not write state file")
	}
	log.Debug().Msg("wrote state to file")
}

func (api *API) authenticate() {
	// generate a unique stateKey to identify this request
	stateKey := stateCode()

	var oauthCode string
	var oauthState string

	// Setup and start HTTP server for OAUTH2 redirects
	httpListener, err := net.Listen("tcp", "127.0.0.1:31022")
	if err != nil {
		log.Panic().Err(err).Msg("cannot create http server listener")
	}

	listener, err := stoppableListener.New(httpListener)
	if err != nil {
		log.Panic().Err(err).Msg("cannot create stoppable listener")
	}

	// API routes
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.RequestURI == "/favicon.ico" {
			w.WriteHeader(404)
			return
		}
		// extract parameters
		params := r.URL.Query()
		oauthCode = params.Get("code")
		oauthState = params.Get("state")

		// write response to server
		if oauthState != stateKey {
			w.WriteHeader(400)
			io.WriteString(w, "state does not match - authentication failed")
			log.Panic().Msg("state key does not match - exiting authentication attempt")
		} else {
			w.WriteHeader(200)
			io.WriteString(w, "You can close this window; successfully authenticated with TradeStation!\n")
		}

		listener.Stop()
	})

	// Start server on port specified above
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		http.Serve(listener, nil)
	}()

	authUrl := fmt.Sprintf("https://signin.tradestation.com/authorize?response_type=code&client_id=%s&redirect_uri=%s&audience=https://api.tradestation.com&state=%s&scope=openid%%20profile%%20MarketData%%20ReadAccount%%20Trade", ApiKey(), "http://localhost:31022", stateKey)
	log.Debug().Str("Auth URL", authUrl).Msg("authorization url")

	browser.OpenURL(authUrl)
	wg.Wait()

	// exchange code for a token
	token := OAuthToken{}
	curl := resty.New()
	resp, err := curl.R().
		SetFormData(map[string]string{
			"grant_type":    "authorization_code",
			"client_id":     ApiKey(),
			"client_secret": Secret(),
			"code":          oauthCode,
			"redirect_uri":  "http://localhost:31022",
		}).
		SetResult(&token).
		Post("https://signin.tradestation.com/oauth/token")
	if err != nil {
		log.Panic().Err(err).Msg("err exchanging oauth code for a token")
	}
	if resp.StatusCode() >= 300 {
		log.Panic().Int("StatusCode", resp.StatusCode()).Msg("request failed")
	}

	api.client.SetAuthScheme("Bearer")
	api.client.SetAuthToken(token.AccessToken)

	api.token = &token

	api.writeStateFile()
}
