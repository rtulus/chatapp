package gorilla

import (
	"log"
	"net/http"
	"strings"

	"github.com/gorilla/websocket"
	"github.com/julienschmidt/httprouter"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

var hm *HubManager

type WebsocketMessage struct {
	HubID   string `json:"hub_id"`
	From    int64  `json:"from, omitempty"`
	Message string `json:"message"`
}

func InitWebsocket() {
	hm = newHubManager()
}

func ServeWebsocket(w http.ResponseWriter, r *http.Request, params *httprouter.Params) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	uid := params.ByName("user_id")
	client := &Client{
		userID:     uid,
		conn:       conn,
		send:       make(chan WebsocketMessage),
		joinedHubs: make(map[string]*Hub),
	}

	hubIDs := params.ByName("hub_ids")
	hubIDarr := strings.Split(hubIDs, ",")
	for _, hubID := range hubIDarr {
		client.joinHub(hubID)
	}

	go client.writePump()
	client.readPump()
}
