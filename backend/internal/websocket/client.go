package websocket

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

var (
	// pongWait is how long we will await a pong response from client
	pongWait = 10 * time.Second
	// pingInterval has to be less than pongWait, We cant multiply by 0.9 to get 90% of time
	// Because that can make decimals, so instead *9 / 10 to get 90%
	// The reason why it has to be less than PingRequency is because otherwise it will send a new Ping before getting response
	pingInterval = (pongWait * 9) / 10
)

type ClientList map[*Client]bool

type Client struct {
	id         uuid.UUID
	connection *websocket.Conn
	manager    *Manager
	// egress is used to avoid concurrent writes on the WebSocket
	egress chan Event
	logger logger
}

func NewClient(conn *websocket.Conn, manager *Manager) (*Client, error) {
	// Max size in bytes
	conn.SetReadLimit(512)

	if err := conn.SetReadDeadline(time.Now().Add(pongWait)); err != nil {
		return nil, errors.New("error while setting ReadDeadLine to connection")
	}
	conn.SetPongHandler(func(pongMsg string) error {
		return conn.SetReadDeadline(time.Now().Add(pongWait))
	})

	id, err := uuid.NewRandom()
	if err != nil {
		return nil, errors.New("error while generating client uuid")
	}
	client := &Client{
		id:         id,
		connection: conn,
		manager:    manager,
		egress:     make(chan Event),
		logger:     manager.logger,
	}

	return client, nil
}

func (c *Client) readMessages() {

	defer func() {
		c.manager.removeClient(c)
	}()

	for {
		_, payload, err := c.connection.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				c.logger.Error("error reading message: " + err.Error())
			}

			return
		}

		var request Event
		if err := json.Unmarshal(payload, &request); err != nil {
			c.logger.Error("error marshalling message: " + err.Error())

			return
		}

		if err := c.manager.routeEvent(request, c); err != nil {
			c.logger.Error("error handling message: " + err.Error())
		}
	}
}

func (c *Client) writeMessages() {

	ticker := time.NewTicker(pingInterval)
	defer func() {
		ticker.Stop()
		c.manager.removeClient(c)
	}()

	for {
		select {
		case message, ok := <-c.egress:
			if !ok {
				if err := c.connection.WriteMessage(websocket.CloseMessage, nil); err != nil {
					c.logger.Info("connection closed: " + err.Error())
				}

				return
			}

			data, err := json.Marshal(message)
			if err != nil {
				c.logger.Error("error while marshaling in writeMessages: " + err.Error())

				return
			}

			if err := c.connection.WriteMessage(websocket.TextMessage, data); err != nil {
				c.logger.Error("error while writing message: " + err.Error())
			}
			c.logger.Info(fmt.Sprintf("message has been sent by: %v", c.id))

		case <-ticker.C:
			if err := c.connection.WriteMessage(websocket.PingMessage, []byte{}); err != nil {
				c.logger.Error("error while writing ping: " + err.Error())

				return
			}
		}
	}
}
