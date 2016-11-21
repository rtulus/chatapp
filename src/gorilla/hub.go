package gorilla

import (
	"errors"
	"fmt"
)

type HubManager struct {
	hubs map[int64]*Hub
}

func newHubManager() *HubManager {
	return &HubManager{
		hubs: make(map[int64]*Hub),
	}
}

func (hm *HubManager) getHub(id int64) *Hub {

	hub, exist := hm.hubs[id]
	if !exist {
		hub := newHub(id)
		hm.registerHub(id, hub)

		hub.run()
	}
	return hub
}

func (hm *HubManager) registerHub(id int64, h *Hub) error {

	if _, exist := hm.hubs[id]; !exist {
		hm.hubs[id] = h
		return nil
	}
	err := errors.New(fmt.Sprint("Failed to register Hub - ID [%d] already exists", id))
	return err
}

func (hm *HubManager) unregisterHub(id int64) {

	if hub, ok := hm.hubs[id]; ok {
		close(hub.register)
		close(hub.unregister)
		close(hub.broadcast)
		delete(hm.hubs, id)
	}
}

type Hub struct {

	// Hub ID
	id int64

	// Registered clients.
	clients map[*Client]bool

	// Register requests from the clients.
	register chan *Client

	// Unregister requests from clients.
	unregister chan *Client

	// Inbound messages from the clients.
	broadcast chan *WebsocketMessage
}

func newHub(id int64) *Hub {
	return &Hub{
		id:         id,
		clients:    make(map[*Client]bool),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan *WebsocketMessage),
	}
}

func (h *Hub) run() {
	defer func() {
		hm.unregisterHub(h.id)
	}()

	for {
		select {
		case client := <-h.register:
			h.clients[client] = true
		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {

				// delete client from hub
				delete(h.clients, client)
				close(client.send)

				// if no more client exist, destroy hub
				if len(h.clients) <= 0 {
					hm.unregisterHub(h.id)
				}
			}
		case message := <-h.broadcast:
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(h.clients, client)
				}
			}
		}
	}
}
