package proxy

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"time"
)

type myTransport struct {
	// Uncomment this if you want to capture the transport
	// CapturedTransport http.RoundTripper
}
type Montioringpath struct {
	Path        string
	Count       int64
	Duration    int64
	AverageTime int64
}

var globalMap = make(map[string]Montioringpath)

func (t *myTransport) RoundTrip(request *http.Request) (*http.Response, error) {

	fmt.Println("---------------------------New Request--------------------------------------------------")
	buf, _ := ioutil.ReadAll(request.Body)
	rdr1 := ioutil.NopCloser(bytes.NewBuffer(buf))
	rdr2 := ioutil.NopCloser(bytes.NewBuffer(buf))

	fmt.Println("Request body : ", rdr1)
	request.Body = rdr2 // OK since rdr2 implements the

	start := time.Now()
	response, err := http.DefaultTransport.RoundTrip(request)
	if err != nil {
		print("\n\ncame in error resp here", err)
		return nil, err //Server is not reachable. Server not working
	}
	elapsed := time.Since(start)

	key := request.Method + "-" + request.URL.Path //for example for POST Method with /path1 as url path key=POST-/path1

	if val, ok := globalMap[key]; ok {
		val.Count = val.Count + 1
		val.Duration += elapsed.Nanoseconds()
		val.AverageTime = val.Duration / val.Count
		globalMap[key] = val
		//do something here
	} else {
		var m Montioringpath
		m.Path = request.URL.Path
		m.Count = 1
		m.Duration = elapsed.Nanoseconds()
		m.AverageTime = m.Duration / m.Count
		globalMap[key] = m
	}
	b, err := json.MarshalIndent(globalMap, "", "  ")
	if err != nil {
		fmt.Println("error:", err)
	}

	body, err := httputil.DumpResponse(response, true)
	if err != nil {
		print("\n\nerror in dumb response")
		// copying the response body did not work
		return nil, err
	}

	log.Println("Response Body : ", string(body))
	log.Println("Response Time:", elapsed.Nanoseconds())

	fmt.Println("Analysis on different Path :", string(b))

	return response, err
}
