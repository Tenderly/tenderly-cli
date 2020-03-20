package client

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/tenderly/tenderly-cli/config"
	"github.com/tenderly/tenderly-cli/userError"
	"io"
	"io/ioutil"
	"net/http"
	"os"
)

func Request(method, path string, body []byte) io.Reader {
	apiBase := "http://127.0.0.1:8080"
	if alternativeApiBase := config.MaybeGetString("api_base"); len(alternativeApiBase) != 0 {
		apiBase = alternativeApiBase
	}

	requestUrl := fmt.Sprintf("%s/%s", apiBase, path)
	req, err := http.NewRequest(
		method,
		requestUrl,
		bytes.NewReader(body),
	)
	if err != nil {
		userError.LogErrorf("failed creating request: %s", userError.NewUserError(
			err,
			"Failed creating request. Please try again.",
		))
		os.Exit(1)
	}

	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: false}

	if token := config.GetToken(); token != "" {
		// set auth token
		req.Header.Add("Authorization", "Bearer "+token)
	}

	logrus.WithFields(logrus.Fields{
		"request_url":  requestUrl,
		"request_body": string(body),
	}).Debug("Making request")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		userError.LogErrorf("failed making request: %s", userError.NewUserError(
			err,
			"Failed making request. Please try again.",
		))
		os.Exit(1)
	}

	data, err := ioutil.ReadAll(res.Body)
	defer res.Body.Close()

	if err != nil {
		userError.LogErrorf("failed reading response body: %s", userError.NewUserError(
			err,
			"Failed reading response body. Please try again.",
		))
		os.Exit(1)
	}

	logrus.WithField("response_body", string(data)).Debug("Got response with body")

	return bytes.NewReader(data)
}
