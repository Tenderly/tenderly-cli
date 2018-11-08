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

	"github.com/tenderly/tenderly-cli/jsonrpc2"
)

type Proxy struct {
	client *http.Client

	target *url.URL
	proxy  *httputil.ReverseProxy
}

func NewProxy(target string) *Proxy {
	targetUrl, _ := url.Parse(target)

	return &Proxy{
		client: &http.Client{},

		target: targetUrl,
		proxy:  httputil.NewSingleHostReverseProxy(targetUrl),
	}
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
		err = p.processMessage(r, message)
		if err != nil {
			fmt.Printf("Failed processing proxy response: %s\n", err)
			return
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

func (p *Proxy) processMessage(r *http.Request, message *jsonrpc2.Message) error {
	// Process JSONRPC request before sending it to the target server.

	// Send JSONRPC request to target server
	proxyReqData, err := json.Marshal(message)
	if err != nil {
		fmt.Errorf("failed formatting proxy request: %s", err)
	}

	proxyReq, err := http.NewRequest(r.Method, p.target.String(), ioutil.NopCloser(bytes.NewBuffer(proxyReqData)))
	if err != nil {
		return fmt.Errorf("failed creating proxy request: %s", err)
	}

	proxyResp, err := p.client.Do(proxyReq)
	if err != nil {
		return fmt.Errorf("failed sending proxy request: %s", err)
	}

	proxyRespData, err := ioutil.ReadAll(proxyResp.Body)
	if err != nil {
		return fmt.Errorf("failed reading proxy response: %s", err)
	}

	err = json.Unmarshal(proxyRespData, &message)
	if err != nil {
		return fmt.Errorf("failed parsing proxy response: %s", err)
	}

	// Process JSONRPC response after receiving it from the target server.

	return nil
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

	proxy := NewProxy(targetSchema + "://" + targetHost + ":" + targetPort)

	return http.ListenAndServe(proxyHost+":"+proxyPort, proxy)
}
