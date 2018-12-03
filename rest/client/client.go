package client

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"github.com/tenderly/tenderly-cli/config"
	"io"
	"net/http"
	"os"
)

func Request(method, path string, body []byte) io.Reader {
	req, err := http.NewRequest(
		method,
		fmt.Sprintf("%s/%s", "http://api.tenderly.love", path),
		bytes.NewReader(body),
	)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to create request")
		os.Exit(1)
	}

	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: false}

	if token := config.GetToken(); token != "" {
		// set auth token
		req.Header.Add("Authorization", "Bearer "+token)
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		// print error and exit
		fmt.Fprintf(os.Stderr, "Failed to make request - %s\n", err)
		os.Exit(0)
	}

	return res.Body
}
