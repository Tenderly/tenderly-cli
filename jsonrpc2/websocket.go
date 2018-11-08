package jsonrpc2

import (
	"fmt"
	"io"

	"github.com/gorilla/websocket"
)

type websocketConnection struct {
	ws *websocket.Conn
}

func DialWebsocketConnection(host string) (Connection, error) {
	ws, _, err := websocket.DefaultDialer.Dial(host, nil)
	if err != nil {
		return nil, fmt.Errorf("open websocket connection: %s", err)
	}

	return &websocketConnection{
		ws: ws,
	}, nil
}

func (conn *websocketConnection) Write(r *Request) error {
	err := conn.ws.WriteJSON(r)
	if err != nil {
		return fmt.Errorf("write websocket: %s", err)
	}

	return nil
}

func (conn *websocketConnection) Read() (*Message, error) {
	var msg Message

	err := conn.ws.ReadJSON(&msg)
	if err == io.ErrUnexpectedEOF {
		return nil, io.EOF
	}
	if err != nil {
		return nil, fmt.Errorf("read websocket: %s", err)
	}

	return &msg, nil
}

func (conn *websocketConnection) Close() error {
	return conn.ws.Close()
}
