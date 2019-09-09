package proxy

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/logrusorgru/aurora"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/tenderly/tenderly-cli/userError"

	"github.com/tenderly/tenderly-cli/ethereum"
	"github.com/tenderly/tenderly-cli/ethereum/client"
	"github.com/tenderly/tenderly-cli/jsonrpc2"
)

type Proxy struct {
	client *client.Client

	target *url.URL
	proxy  *httputil.ReverseProxy
}

var projectPath string
var buildDirectory string

func NewProxy(target string) (*Proxy, error) {
	targetUrl, _ := url.Parse(target)

	c, err := client.Dial(target)
	if err != nil {
		return nil, userError.NewUserError(
			fmt.Errorf("failed calling target ethereum blockchain on %s", target),
			fmt.Sprintf("Couldn't connect to target Ethereum blockchain at: %s", target),
		)
	}

	return &Proxy{
		client: c,

		target: targetUrl,
		proxy:  httputil.NewSingleHostReverseProxy(targetUrl),
	}, nil
}

func (p *Proxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		userError.LogErrorf("failed reading request body: %s",
			userError.NewUserError(
				err,
				"Failed reading proxy request",
			),
		)
		return
	}

	messages, err := unmarshalMessages(data)
	if err != nil {
		userError.LogErrorf("failed parsing request body: %s",
			userError.NewUserError(
				err,
				"Failed parsing proxy request body",
			),
		)
		return
	}

	for _, message := range messages {
		err := p.client.Call(message)
		if err != nil {
			userError.LogErrorf("failed processing proxy request: %s",
				userError.NewUserError(
					err,
					fmt.Sprintf("Failed processing proxy request: %s", err),
				),
			)

			continue
		}

		// @TODO: Extract into a more manageable format.
		if message.Method == "eth_getTransactionReceipt" {
			receipt, err := p.GetTraceReceipt(string(message.Params[2:68]), false)
			if err != nil {
				userError.LogErrorf("couldn't extract trace from: %s",
					userError.NewUserError(
						err,
						"Couldn't extract trace",
					),
				)
				continue
			}

			message.Result, err = json.Marshal(receipt)
			if err != nil {
				userError.LogErrorf("failed encoding transaction receipt: %s",
					userError.NewUserError(
						err,
						fmt.Sprintf("Failed encoding transaction receipt: %s", err),
					),
				)
				continue
			}

			if strings.HasPrefix(receipt.Status(), "0x0") {
				message.Error = &jsonrpc2.Error{
					Message: receipt.Status(),
				}
			}
		}

		if (message.Method == "eth_sendRawTransaction" || message.Method == "eth_sendTransaction") && message.Error != nil {
			var failedTx string
			err = json.Unmarshal(message.Result, &failedTx)
			if err != nil {
				userError.LogErrorf("couldn't extract transaction hash: %s",
					userError.NewUserError(
						err,
						"Couldn't extract transaction hash",
					),
				)
				continue
			}
			receipt, err := p.GetTraceReceipt(failedTx, true)
			if err != nil {
				userError.LogErrorf("couldn't extract trace: %s",
					userError.NewUserError(
						err,
						"Couldn't extract trace",
					),
				)
				continue
			}

			if message.Error != nil {
				message.Error.Message = receipt.Status()
			}
		}
	}

	var respData []byte
	if isBatchRequest(data) {
		respData, err = json.Marshal(messages)
	} else {
		respData, err = json.Marshal(messages[0])
	}
	if err != nil {
		userError.LogErrorf("failed formatting proxy response: %s",
			userError.NewUserError(
				err,
				"Failed formatting proxy response",
			),
		)
		return
	}

	// Send the final response to the caller.
	_, err = io.Copy(w, bytes.NewReader(respData))
	if err != nil {
		userError.LogErrorf("failed sending proxy response: %s",
			userError.NewUserError(
				err,
				"Failed sending proxy response",
			),
		)
		return
	}

	logrus.Infof("%s", respData)
}

func (p *Proxy) GetTraceReceipt(tx string, wait bool) (ethereum.TransactionReceipt, error) {
	receipt, err := p.waitForReceipt(tx, wait)
	if err != nil {
		return nil, err
	}

	if receipt.Status() != "0x0" {
		return receipt, nil
	}

	err = p.Trace(receipt, projectPath, buildDirectory)
	if err != nil {
		return nil, userError.NewUserError(
			fmt.Errorf("get transaction trace: %s", err),
			"Couldn't get transaction trace",
		)
	}

	return receipt, nil
}

func (p *Proxy) waitForReceipt(tx string, wait bool) (ethereum.TransactionReceipt, error) {
	attempts := 200
	waitFor := 500 * time.Millisecond

	var receipt ethereum.TransactionReceipt
	var err error

	for {
		receipt, err = p.getReceipt(tx)
		if !wait || err == nil {
			return receipt, err
		}

		attempts--
		if attempts == 0 {
			return nil, err
		}

		logrus.Info("Waiting for transaction receipt...")
		time.Sleep(waitFor)
	}

	return receipt, err
}

func (p *Proxy) getReceipt(tx string) (ethereum.TransactionReceipt, error) {
	receipt, err := p.client.GetTransactionReceipt(tx)
	if err != nil {
		return nil, userError.NewUserError(
			fmt.Errorf("get transaction receipt: %s", err),
			"Error getting transaction receipt",
		)
	}

	if receipt.Hash() == "" {
		return nil, userError.NewUserError(
			errors.New("transaction status is missing the hash"),
			"Transaction status is missing the hash",
		)
	}

	return receipt, nil
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

func Start(targetSchema, targetHost, targetPort, proxyHost, proxyPort, path, buildDir string) error {
	logrus.Infof("Proxy starting on %s:%s", proxyHost, proxyPort)

	projectPath = path
	buildDirectory = buildDir

	host := getTargetHost(targetHost, targetSchema, targetPort)
	logrus.Infof("Redirecting calls to %s", host)

	proxy, err := NewProxy(host)
	if err != nil {
		userError.LogErrorf("failed starting proxy: %s", err)
		os.Exit(1)
	}

	return http.ListenAndServe(proxyHost+":"+proxyPort, proxy)
}

func getTargetHost(targetHost, targetSchema, targetPort string) string {
	initialSchema := "http"
	if strings.HasPrefix(targetHost, "https") {
		initialSchema = "https"
	}

	re := regexp.MustCompile(`^(https?://|www\.)+`)

	host := fmt.Sprintf("%s://%s", initialSchema, re.ReplaceAllString(targetHost, ""))

	parsedUrl, err := url.Parse(host)
	if err != nil {
		userError.LogErrorf("couldn't parse target host: %s", userError.NewUserError(
			err,
			aurora.Sprintf("Couldn't parse target host: %s", aurora.Bold(aurora.Red(host))),
		))
		os.Exit(1)
	}

	if len(targetSchema) > 0 {
		parsedUrl.Scheme = targetSchema
	}
	if len(targetPort) > 0 {
		parsedUrl.Host = fmt.Sprintf("%s:%s", parsedUrl.Hostname(), targetPort)
	}

	return parsedUrl.String()
}
