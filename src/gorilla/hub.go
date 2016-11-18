package gorilla

import (
	"errors"
)

type HubManager struct {
	hubs map[string]*Hub
}

func newHubManager() *HubManager {
	return &HubManager{
		hubs: make(map[string]*Hub),
	}
}

func (hm *HubManager) getHub(id string) *Hub {

	hub, exist := hm.hubs[id]
	if !exist {
		hub := newHub(id)
		hm.registerHub(id, hub)

		hub.run()
	}
	return hub
}

func (hm *HubManager) registerHub(id string, h *Hub) error {

	if _, exist := hm.hubs[id]; !exist {
		hm.hubs[id] = h
		return nil
	}
	err := errors.New("Failed to register Hub - ID [" + id + "] already exists")
	return err
}

func (hm *HubManager) unregisterHub(id string) {

	if hub, ok := hm.hubs[id]; ok {
		close(hub.register)
		close(hub.unregister)
		close(hub.broadcast)
		delete(hm.hubs, id)
	}
}

type Hub struct {

	// Hub ID
	id string

	// Registered clients.
	clients map[*Client]bool

	// Register requests from the clients.
	register chan *Client

	// Unregister requests from clients.
	unregister chan *Client

	// Inbound messages from the clients.
	broadcast chan []byte
}

func newHub(id string) *Hub {
	return &Hub{
		id:         id,
		clients:    make(map[*Client]bool),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan []byte),
	}
}

func (h *Hub) run() {
	defer func() {
		hm.unregisterHub(h)
	}()

	for {
		select {
		case client := <-h.register:
			h.clients[client] = true
		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)

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
