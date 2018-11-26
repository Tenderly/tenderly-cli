package client

import (
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"os"
)

func Request(method, path, token string, body io.ReadWriter) io.Reader {
	req, err := http.NewRequest(method, fmt.Sprintf("%s/%s", "http://api.tenderly.love", path), body)
	if err != nil {
		// print error and exit
		fmt.Fprintln(os.Stderr, "Failed to create request")
		os.Exit(0)
	}

	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: false}

	if token != "" {
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
