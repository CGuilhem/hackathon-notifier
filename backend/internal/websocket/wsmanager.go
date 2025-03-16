package websocket

import (
	"errors"
	"fmt"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

type logger interface {
	Error(msg string)
	Info(msg string)
	Warn(msg string)
}

var (
	websocketUpgrader = websocket.Upgrader{
		CheckOrigin:     checkOrigin,
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
)

var (
	ErrEventNotSupported = errors.New("this event type is not supported")
)

type Manager struct {
	clients  ClientList
	handlers map[string]EventHandler
	logger   logger

	sync.RWMutex
}

func NewManager(logger logger) *Manager {
	m := &Manager{
		clients:  make(ClientList),
		handlers: make(map[string]EventHandler),
		logger:   logger,
	}
	m.setupEventHandlers()

	return m
}

// setupEventHandlers configures and adds all handlers
func (m *Manager) setupEventHandlers() {
	m.handlers[EventSendMessage] = sendMessageHandler
}

func (m *Manager) ServeWS(w http.ResponseWriter, r *http.Request) {

	m.logger.Info("New connection")

	// To remove
	websocketUpgrader.CheckOrigin = func(r *http.Request) bool { return true }

	conn, err := websocketUpgrader.Upgrade(w, r, nil)
	if err != nil {
		m.logger.Error(err.Error())

		return
	}

	client, err := NewClient(conn, m)
	if err != nil {
		m.logger.Error(err.Error())
		// To do send error message to client
	}
	m.addClient(client)

	go client.readMessages()
	go client.writeMessages()
}

// routeEvent is used to make sure the correct event goes into the correct handler
func (m *Manager) routeEvent(event Event, c *Client) error {

	if handler, ok := m.handlers[event.Type]; ok {
		if err := handler(event, c); err != nil {

			return err
		}
		return nil

	} else {
		return ErrEventNotSupported
	}
}

func (m *Manager) addClient(client *Client) {
	m.Lock()
	defer m.Unlock()

	m.clients[client] = true
	m.logger.Info(fmt.Sprintf("client has been added: %v", client.id))
}

func (m *Manager) removeClient(client *Client) {
	m.Lock()
	defer m.Unlock()

	if _, ok := m.clients[client]; ok {
		client.connection.Close()
		delete(m.clients, client)
		m.logger.Info(fmt.Sprintf("client has been deleted: %v", client.id))
	}
}

// checkOrigin will check origin and return true if its allowed
// Maybe create a middleware and/or add a reverse-proxy
func checkOrigin(r *http.Request) bool {

	origin := r.Header.Get("Origin")

	switch origin {
	case "http://localhost:5173":
		return true
	default:
		return false
	}
}
