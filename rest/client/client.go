package client

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"github.com/tenderly/tenderly-cli/rest/payloads"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/tenderly/tenderly-cli/config"
	"github.com/tenderly/tenderly-cli/userError"
)

const (
	sessionLimitErrorSlug = "session_limit_exceeded"
	defaultApiBaseURL     = "https://api.tenderly.co"
)

func Request(method, path string, body []byte) io.Reader {
	apiBaseURL := resolveApiBaseURL()
	requestURL := resolveRequestURL(apiBaseURL, path)
	req, err := http.NewRequest(
		method,
		requestURL,
		bytes.NewReader(body),
	)
	if err != nil {
		userError.LogErrorf("failed to create request: %s", userError.NewUserError(
			err,
			"Failed to create request. Please try again.",
		))
		os.Exit(1)
	}

	ensureTLS()

	if key := config.GetAccessKey(); key != "" {
		// set access key
		req.Header.Add("x-access-key", key)
	} else if token := config.GetToken(); token != "" {
		// set auth token
		req.Header.Add("Authorization", "Bearer "+token)

		urlPath := fmt.Sprintf("api/v1/account/%s/token", config.GetAccountId())
		if requestURL != fmt.Sprintf("%s/%s", apiBaseURL, urlPath) {
			var request payloads.GenerateAccessTokenRequest
			request.Name = "CLI access token"

			body, err := json.Marshal(request)
			if err != nil {
				logrus.Debug("failed to marshall generate access token request", logrus.Fields{
					"url_path":   urlPath,
					"account_id": config.GetAccountId(),
				})
			} else {
				reader := Request(
					"POST",
					urlPath,
					body,
				)

				var tokenResp payloads.TokenResponse
				err := json.NewDecoder(reader).Decode(&tokenResp)

				if err != nil || tokenResp.Error != nil {
					if tokenResp.Error.Slug == sessionLimitErrorSlug {
						config.SetGlobalConfig(config.Token, "")
						err = config.WriteGlobalConfig()
						if err != nil {
							userError.LogErrorf(
								"write global config: %s",
								userError.NewUserError(err, "Couldn't write global config file"),
							)

							return nil
						}
						logrus.Info("Maximum number of active sessions exceeded. " +
							"You are allowed to have no more than 3 simultaneously active sessions. \n" +
							"Please use login again with tenderly login command in order to generate new session.")

						os.Exit(1)
					}

					logrus.Debug("failed creating token, user has been logged out")
				}

				config.SetGlobalConfig(config.AccessKey, tokenResp.Token)
				config.SetGlobalConfig(config.AccessKeyId, tokenResp.ID)

				//@TODO(filip): remove this once we
				err = config.WriteGlobalConfig()
				if err != nil {
					userError.LogErrorf(
						"write global config: %s",
						userError.NewUserError(err, "Couldn't write global config file"),
					)

					return nil
				}

				req.Header.Add("x-access-key", tokenResp.Token)
				req.Header.Del("Authorization")
			}
		}
	}

	logrus.WithFields(
		logrus.Fields{
			"request_url":  requestURL,
			"request_body": string(body),
		},
	).Debug("Making request")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		userError.LogErrorf("failed making request: %s", userError.NewUserError(
			err,
			"Failed making request. Please try again.",
		))
		os.Exit(1)
	}
	defer func() {
		_ = res.Body.Close()
	}()

	data, err := io.ReadAll(res.Body)
	logrus.WithField("response_body", string(data)).Debug("Got response with body")

	handleResponseStatus(res, data, err)

	if err != nil {
		userError.LogErrorf("failed reading response body: %s", userError.NewUserError(
			err,
			"Failed reading response body. Please try again.",
		))
		os.Exit(1)
	}

	return bytes.NewReader(data)
}

// handleResponseStatus handles the response status code.
func handleResponseStatus(res *http.Response, resBodyData []byte, err error) {
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		if res.StatusCode >= 500 {
			userError.LogErrorf("request failed: %s", userError.NewUserError(
				err,
				fmt.Sprintf(
					"The request failed with a status code of %d and status message '%s'. Please try again.",
					res.StatusCode,
					res.Status,
				),
			))
		} else {
			userError.LogErrorf("request failed: %s", userError.NewUserError(
				err,
				fmt.Sprintf(
					"The request failed with a status code of %d and message '%s'",
					res.StatusCode,
					extractErrorMessage(resBodyData, res.Status),
				),
			))
		}
		os.Exit(1)
	}
}

// extractErrorMessage tries to extract an error message from the (JSON) response body.
// Falls back to the HTTP status if the response is not valid JSON or has no error message.
func extractErrorMessage(resBodyData []byte, status string) string {
	var errorResp struct {
		Error struct {
			ID      string                 `json:"id"`
			Slug    string                 `json:"slug"`
			Message string                 `json:"message"`
			Data    map[string]interface{} `json:"data"`
		} `json:"error"`
	}

	if json.Unmarshal(resBodyData, &errorResp) == nil && errorResp.Error.Message != "" {
		return errorResp.Error.Message
	}

	return status
}

// ensureTLS configures the default http transport to use TLS.
func ensureTLS() {
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{
		InsecureSkipVerify: false,
	}
}

// resolveRequestURL resolves the request URL based on the API base URL and the path.
func resolveRequestURL(apiBaseURL string, path string) string {
	requestURL := fmt.Sprintf("%s/%s", apiBaseURL, strings.TrimPrefix(path, "/"))
	return requestURL
}

// resolveApiBaseURL resolves the API base URL based on the config.
func resolveApiBaseURL() string {
	apiBase := defaultApiBaseURL
	if apiBaseOverride := config.MaybeGetString("api_base"); len(apiBaseOverride) != 0 {
		apiBase = apiBaseOverride
	}
	return apiBase
}
