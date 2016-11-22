package gorilla

import (
	"errors"
	"fmt"
	"log"
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
	log.Println("getting new hub", id)
	hub, exist := hm.hubs[id]
	if !exist {
		hub = newHub(id)
		hm.registerHub(id, hub)
		go hub.run()
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

	if _, ok := hm.hubs[id]; ok {
		// log.Println("closing register")
		// close(hub.register)
		// log.Println("closing unregister")
		// close(hub.unregister)
		// log.Println("closing broadcast")
		// close(hub.broadcast)
		log.Println("delete hub from hub manager")
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
			log.Println("register user", client.userID, "to hub", h.id)
			h.clients[client] = true

		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				log.Println("unregister client", client.userID, "from hub", h.id)
				// delete client from hub
				delete(h.clients, client)
				close(client.send)

				// if no more client exist, destroy hub
				if len(h.clients) <= 0 {
					hm.unregisterHub(h.id)
				}
			}

		case message := <-h.broadcast:
			log.Println("broadcast message on hub", h.id)
			log.Printf("message: %+v", message)
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
