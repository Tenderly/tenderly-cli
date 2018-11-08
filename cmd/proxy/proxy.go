package proxy

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/tenderly/tenderly-cli/jsonrpc2"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
)

type Prox struct {
	target *url.URL
	proxy  *httputil.ReverseProxy
}

func NewProxy(target string) *Prox {
	url, _ := url.Parse(target)

	return &Prox{target: url, proxy: httputil.NewSingleHostReverseProxy(url)}
}

func (p *Prox) handle(w http.ResponseWriter, r *http.Request) {
	buf, _ := ioutil.ReadAll(r.Body)
	var message jsonrpc2.Message
	err := json.Unmarshal(buf, &message)
	if err != nil {
		print("\n\nerror in unmarshaling response")
		// unmarshaling the response body did not work
		return
	}

	rdr1 := ioutil.NopCloser(bytes.NewBuffer(buf))
	rdr2 := ioutil.NopCloser(bytes.NewBuffer(buf))

	fmt.Println("Request body : ", rdr1)
	r.Body = rdr2 // OK since rdr2 implements the

	response, err := http.DefaultTransport.RoundTrip(r)
	if err != nil {
		print("\n\ncame in error resp here", err)
		return //Server is not reachable. Server not working
	}

	body, err := httputil.DumpResponse(response, true)
	if err != nil {
		print("\n\nerror in dumb response")
		// copying the response body did not work
		return
	}

	err = json.Unmarshal(body, &message)
	if err != nil {
		print("\n\nerror in unmarshaling response")
		// unmarshaling the response body did not work
		return
	}

	//message
}

func Start(targetSchema, targetHost, targetPort, proxyHost, proxyPort, path, network string) {
	flag.Parse()

	fmt.Println(fmt.Sprintf("server will run on %s:%s", proxyHost, proxyPort))
	fmt.Println(fmt.Sprintf("redirecting to %s:%s", targetHost, targetPort))

	// proxy
	proxy := NewProxy(targetSchema + "://" + targetHost + ":" + targetPort)

	http.HandleFunc("/proxyServer", Server)

	// server redirection
	http.HandleFunc("/", proxy.handle)
	log.Fatal(http.ListenAndServe(proxyHost+":"+proxyPort, nil))
}

func Server(w http.ResponseWriter, r *http.Request) {

}
