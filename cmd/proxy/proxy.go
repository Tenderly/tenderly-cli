package proxy

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/tenderly/tenderly-cli/ethereum/client"
	"github.com/tenderly/tenderly-cli/jsonrpc2"
)

type Proxy struct {
	client *client.Client

	target *url.URL
	proxy  *httputil.ReverseProxy
}

func NewProxy(target string) (*Proxy, error) {
	targetUrl, _ := url.Parse(target)

	client, err := client.Dial(target)
	if err != nil {
		return nil, fmt.Errorf("Failed calling target ethereum blockchain on %s", target)
	}

	return &Proxy{
		client: client,

		target: targetUrl,
		proxy:  httputil.NewSingleHostReverseProxy(targetUrl),
	}, nil
}

func (p *Proxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Printf("Failed reading request body: %s\n", err)
		return
	}

	messages, err := unmarshalMessages(data)
	if err != nil {
		fmt.Printf("Failed parsing request body: %s\n", err)
		return
	}

	for _, message := range messages {
		err := p.client.Proxy(message)
		if err != nil {
			fmt.Printf("Failed processing proxy request: %s\n", err)
		}
	}

	var respData []byte
	if isBatchRequest(data) {
		respData, err = json.Marshal(messages)
	} else {
		respData, err = json.Marshal(messages[0])
	}
	if err != nil {
		fmt.Printf("Failed formatting proxy response: %s\n", err)
		return
	}

	// Send the final response to the caller.
	_, err = io.Copy(w, bytes.NewReader(respData))
	if err != nil {
		fmt.Printf("Failed sending proxy response: %s\n", err)
		return
	}

	fmt.Printf("%s\n", respData)
}

func unmarshalMessages(data []byte) ([]*jsonrpc2.Message, error) {
	var messages []*jsonrpc2.Message
	var err error
	if isBatchRequest(data) {
		err = json.Unmarshal(data, &messages)

		return messages, err
	}

	var message jsonrpc2.Message
	err = json.Unmarshal(data, &message)

	messages = append(messages, &message)

	return messages, err
}

func isBatchRequest(data []byte) bool {
	for _, b := range data {
		if b == 0x20 || b == 0x09 || b == 0x0a || b == 0x0d {
			continue
		}

		return b == '['
	}

	return false
}

func Start(targetSchema, targetHost, targetPort, proxyHost, proxyPort, path, network string) error {
	flag.Parse()

	fmt.Println(fmt.Sprintf("server will run on %s:%s", proxyHost, proxyPort))
	fmt.Println(fmt.Sprintf("redirecting to %s:%s", targetHost, targetPort))

	proxy, err := NewProxy(targetSchema + "://" + targetHost + ":" + targetPort)
	if err != nil {
		fmt.Printf("Failed initiating target proxy %s\n", err)
	}

	return http.ListenAndServe(proxyHost+":"+proxyPort, proxy)
}
