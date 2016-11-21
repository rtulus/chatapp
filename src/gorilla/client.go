package gorilla

import (
	"log"
	"time"

	"github.com/gorilla/websocket"
)

const (
	// Time allowed to write a message to the peer.
	WRITE_WAIT = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	PONG_WAIT = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	PING_PERIOD = (PONG_WAIT * 9) / 10

	// Maximum message size allowed from peer.
	MAX_MESSAGE_SIZE = 512

	// We should have a system to determine what type of message we got
	// and do actions accordingly.
	// eg.
	// 100 = normal broadcast to hubid attached
	// 200 = create room, with room name
	// 201 = rename room, must have hubid attached, must be admin
	// 300 = leave room
	// 301 = leave all
	MESSAGE_TYPE_BROADCAST = 100
)

var (
	newline = []byte{'\n'}
	space   = []byte{' '}
)

type Client struct {
	userID     int64
	conn       *websocket.Conn
	send       chan *WebsocketMessage
	joinedHubs map[int64]*Hub
}

func (c *Client) joinHub(id int64) {
	hub, isNew := hm.getHub(id)
	if isNew {
		go hub.run()
	}
	log.Println("%+v", hub)
	hub.register <- c
	log.Println("register client to hub", id)
}

// readPump pumps messages from the websocket connection to the hub.
//
// The application runs readPump in a per-connection goroutine. The application
// ensures that there is at most one reader on a connection by executing all
// reads from this goroutine.
func (c *Client) readPump() {
	defer func() {
		for _, hub := range c.joinedHubs {
			hub.unregister <- c
		}
		c.conn.Close()
	}()
	c.conn.SetReadLimit(MAX_MESSAGE_SIZE)
	c.conn.SetReadDeadline(time.Now().Add(PONG_WAIT))
	c.conn.SetPongHandler(
		func(string) error {
			c.conn.SetReadDeadline(time.Now().Add(PONG_WAIT))
			return nil
		})

	for {
		msg := WebsocketMessage{}
		err := c.conn.ReadJSON(&msg)
		msg.From = c.userID
		log.Println("%+v", msg)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway) {
				log.Printf("error: %v", err)
			}
			break
		}
		log.Println(msg)
		hub := c.joinedHubs[msg.HubID]
		hub.broadcast <- &msg
	}
}

// writePump pumps messages from the hub to the websocket connection.
//
// A goroutine running writePump is started for each connection. The
// application ensures that there is at most one writer to a connection by
// executing all writes from this goroutine.
func (c *Client) writePump() {
	ticker := time.NewTicker(PING_PERIOD)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()
	for {
		select {
		case wsMsg, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(WRITE_WAIT))
			if !ok {
				// The hub closed the channel.
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write([]byte(wsMsg.Message))

			// Add queued chat messages to the current websocket message.
			// n := len(c.send)
			// for i := 0; i < n; i++ {
			// 	w.Write(newline)
			// 	w.Write([]byte(<-c.send.Message))
			// }

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(WRITE_WAIT))
			if err := c.conn.WriteMessage(websocket.PingMessage, []byte{}); err != nil {
				return
			}
		}
	}
}
