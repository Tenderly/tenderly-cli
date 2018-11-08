package proxy

import (
	"flag"
	"fmt"
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
	w.Header().Set("X-GoProxy", "GoProxy")
	p.proxy.Transport = &myTransport{}

	p.proxy.ServeHTTP(w, r)

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
