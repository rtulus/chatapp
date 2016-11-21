package gorilla

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/julienschmidt/httprouter"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

var uid int64 = 1
var hm *HubManager

type WebsocketMessage struct {
	HubID   int64  `json:"hub_id"`
	From    int64  `json:"from, omitempty"`
	Message string `json:"message"`
}

func InitWebsocket() {
	hm = newHubManager()
}

func ServeWebsocket(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	log.Println("serve client websocket")
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	// uid, err := strconv.ParseInt(params.ByName("user_id"), 10, 64)
	// if err != nil {
	// 	log.Println("Error parsing user_id")
	// }
	client := &Client{
		userID:     uid,
		conn:       conn,
		send:       make(chan *WebsocketMessage),
		joinedHubs: make(map[int64]*Hub),
	}
	uid++
	log.Println("uid", uid)

	// hubIDs := params.ByName("hub_ids")
	// hubIDarr := strings.Split(hubIDs, ",")
	// for _, hid := range hubIDarr {
	// 	if hubID, err := strconv.ParseInt(hid, 10, 64); err == nil {
	// 		client.joinHub(hubID)
	// 	}
	// }
	log.Println("joining hub 1")
	client.joinHub(1)

	go client.writePump()
	client.readPump()
}
