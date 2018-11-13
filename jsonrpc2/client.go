package jsonrpc2

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/nebojsa94/smart-alert/backend/jsonrpc2"
)

var id int64

func nextID() int64 {
	atomic.AddInt64(&id, 1)

	return id
}

type Request struct {
	ID      int64  `json:"id,omitempty"`
	Version string `json:"jsonrpc"`

	Method string        `json:"method"`
	Params []interface{} `json:"params,omitempty"`
}

func NewRequest(method string, params ...interface{}) *Request {
	return &Request{
		ID:      nextID(),
		Version: "2.0",

		Method: method,
		Params: params,
	}
}

type Message struct {
	ID      int64  `json:"id,omitempty"`
	Version string `json:"jsonrpc"`

	Method string          `json:"method,omitempty"`
	Params json.RawMessage `json:"params,omitempty"`

	Result json.RawMessage `json:"result,omitempty"`
	Error  *Error          `json:"error,omitempty"`
}

type Error struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data,omitempty"`
}

func (msg *Message) Reset() {
	msg.ID = 0
	msg.Version = "2.0"
	msg.Method = ""
	msg.Params = nil
	msg.Result = nil
	msg.Error = nil
}

type Connection interface {
	Write(msg *Request) error
	Read() (*Message, error)
	Close() error
}

type Client struct {
	conn Connection

	flying      sync.Map
	subscribers sync.Map
}

func DiscoverAndDial(target string) (client *Client, err error) {
	client, err = Dial(target)
	if err != nil {
		return nil, fmt.Errorf("could not determine protocol")
	}

	return client, nil
}

func Dial(addr string) (*Client, error) {
	conn, err := dialConn(addr)
	if err != nil {
		return nil, fmt.Errorf("dial connection: %s", err)
	}

	client := &Client{
		conn: conn,
	}

	go client.listen()

	return client, nil
}

func dialConn(addr string) (Connection, error) {
	if strings.HasPrefix(addr, "ws") {
		return DialWebsocketConnection(addr)
	}

	if strings.HasPrefix(addr, "http") {
		return DialHttpConnection(addr)
	}

	return nil, fmt.Errorf("unrecognized protocol")
}

func (c *Client) Call(res interface{}, method string, params ...interface{}) error {
	req := NewRequest(method, params...)

	return c.CallRequest(res, req)
}

func (c *Client) CallRequest(res interface{}, req *Request) error {
	resMsg, err := c.SendRawRequest(req)
	if err != nil {
		return err
	}

	if resMsg.Error != nil {
		return fmt.Errorf("request failed: [ %d ] %s", resMsg.Error.Code, resMsg.Error.Message)
	}

	if _, ok := res.(*jsonrpc2.Message); ok {
		res.(*jsonrpc2.Message).Result = resMsg.Result
		return nil
	}

	err = json.Unmarshal(resMsg.Result, res)
	if err != nil {
		return fmt.Errorf("read result: %s", err)
	}

	return nil
}

func (c *Client) SendRawRequest(req *Request) (*Message, error) {
	resCh := make(chan *Message, 1)
	c.setFlying(req.ID, resCh)
	defer func() {
		c.deleteFlying(req.ID)
	}()

	err := c.conn.Write(req)
	if err != nil {
		return nil, fmt.Errorf("write message to socket: %s", err)
	}

	ctx, _ := context.WithTimeout(context.TODO(), 30*time.Second)
	select {
	case <-ctx.Done():
		return nil, fmt.Errorf("request timed out")
	case r := <-resCh:
		return r, nil
	}
}

func (c *Client) Subscribe(id string) (chan *Message, error) {
	subCh := make(chan *Message, 256)

	if c.hasSubscription(id) {
		return nil, fmt.Errorf("subscription %s already exists", id)
	}

	c.setSubscription(id, subCh)

	return subCh, nil
}

func (c *Client) Unsubscribe(id string) error {
	subCh, ok := c.getSubscription(id)
	if !ok {
		return fmt.Errorf("subscription %s does not exist", id)
	}

	c.deleteSubscription(id)
	close(subCh)

	return nil
}

func (c *Client) listen() {
	for {
		msg, err := c.conn.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("failed reading message from connection: %s", err)
		}

		c.processMsg(msg)
	}
}

func (c *Client) processMsg(msg *Message) error {
	if msg.ID == 0 {
		// notification
		c.subscribers.Range(func(_, val interface{}) bool {
			subCh := val.(chan *Message)

			subCh <- msg

			return true
		})

		return nil
	}

	// response
	//@TODO: We don't have to send ID and JSONRPC version back to caller.

	resCh, ok := c.getFlying(msg.ID)

	if !ok {
		log.Printf("dropped message: %d", msg.ID)
		return nil
	}

	resCh <- msg

	return nil
}

func (c *Client) setFlying(id int64, msg chan *Message) {
	c.flying.Store(id, msg)
}

func (c *Client) getFlying(id int64) (chan *Message, bool) {
	msg, ok := c.flying.Load(id)
	if !ok {
		return nil, ok
	}

	return msg.(chan *Message), ok
}

func (c *Client) deleteFlying(id int64) {
	c.flying.Delete(id)
}

func (c *Client) hasSubscription(id string) bool {
	_, ok := c.subscribers.Load(id)

	return ok
}

func (c *Client) setSubscription(id string, msg chan *Message) {
	c.subscribers.Store(id, msg)
}

func (c *Client) getSubscription(id string) (chan *Message, bool) {
	msg, ok := c.subscribers.Load(id)
	if !ok {
		return nil, ok
	}

	return msg.(chan *Message), ok
}

func (c *Client) deleteSubscription(id string) {
	c.subscribers.Delete(id)
}

func (c *Client) Close() error {
	return c.conn.Close()
}
