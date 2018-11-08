package jsonrpc2

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"sync"
)

type httpConnection struct {
	addr string

	respCh chan *Message

	once   sync.Once
	ctx    context.Context
	cancel context.CancelFunc
}

func DialHttpConnection(addr string) (Connection, error) {
	ctx, cancel := context.WithCancel(context.Background())

	conn := &httpConnection{
		addr: addr,

		respCh: make(chan *Message, 256),

		ctx:    ctx,
		cancel: cancel,
	}

	msg := NewRequest("ping")

	err := conn.Write(msg)
	if err != nil {
		return nil, fmt.Errorf("connect via http: %s", err)
	}

	_, err = conn.Read()
	if err != nil {
		return nil, fmt.Errorf("connect via http: %s", err)
	}

	return conn, nil
}

func (conn *httpConnection) Write(msg *Request) error {
	// Setup request, possibly wasteful.
	req, err := http.NewRequest(http.MethodPost, conn.addr, nil)
	if err != nil {
		return fmt.Errorf("write request setup: %s", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("write request body: %s", err)
	}

	req.Body = ioutil.NopCloser(bytes.NewReader(data))
	req.ContentLength = int64(len(data))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("write request send: %s", err)
	}
	if resp.StatusCode < 200 && resp.StatusCode >= 300 {
		return fmt.Errorf("write request unsuccessful: %s", err)
	}

	var respMsg Message

	err = json.NewDecoder(resp.Body).Decode(&respMsg)
	if err != nil {
		return fmt.Errorf("write request read: %s", err)
	}

	conn.respCh <- &respMsg

	err = resp.Body.Close()
	if err != nil {
		return fmt.Errorf("write request cleanup: %s", err)
	}

	return nil
}

func (conn *httpConnection) Read() (*Message, error) {
	select {
	case resp := <-conn.respCh:
		return resp, nil
	case <-conn.ctx.Done():
		return nil, io.EOF
	}
}

func (conn *httpConnection) Close() error {
	conn.once.Do(func() {
		conn.cancel()
	})

	return nil
}
