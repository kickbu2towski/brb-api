package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/kickbu2towski/brb-api/internal/data"
)

type BroadcastMessage struct {
	BroadcastTo []int          `json:"-"`
	Data        map[string]any `json:"data"`
	toEveryone  bool           `json:"-"`
}

type Hub struct {
	clients   map[*Client]bool
	broadcast chan *BroadcastMessage
	models    *data.Models
}

func NewHub(models *data.Models) *Hub {
	return &Hub{
		clients:   make(map[*Client]bool),
		broadcast: make(chan *BroadcastMessage),
		models:    models,
	}
}

func (h *Hub) run() {
	for {
		msg := <-h.broadcast
		for client := range h.clients {
			allowed := msg.toEveryone
			if !allowed {
				allowed = Includes(msg.BroadcastTo, client.user.ID)
			}
			if allowed {
				err := client.conn.WriteJSON(msg)
				if err != nil {
					log.Println("error writing ws json:", err)
					delete(h.clients, client)
				}
			}
		}
	}
}

type Client struct {
	user *data.BasicUserResp
	hub  *Hub
	conn *websocket.Conn
}

func (c *Client) save(e *data.Event) (string, error) {
	var msgID string
	b, err := json.Marshal(e.Payload)
	if err != nil {
		return msgID, err
	}

	switch e.Type {
	case "Create":
		var m data.Message
		err = json.Unmarshal(b, &m)
		if err != nil {
			return msgID, err
		}
		msgID = m.ID
		m.UserID = c.user.ID
		err = c.hub.models.Messages.InsertMessage(context.Background(), &m)
		if err != nil {
			return msgID, err
		}
	case "Edit", "Delete", "Reaction":
		ctx := context.Background()

		var err error
		var payload struct {
			ID       string `json:"id"`
			Content  string `json:"content"`
			Reaction string `json:"reaction"`
			ToRemove bool   `json:"toRemove"`
		}

		err = json.Unmarshal(b, &payload)
		if err != nil {
			return msgID, err
		}

		msgID = payload.ID

		if e.Type != "Reaction" {
			msg, err := c.hub.models.Messages.GetMessage(ctx, payload.ID, c.user.ID)
			if err != nil {
				return msgID, err
			}
			if e.Type == "Edit" {
				msg.Content = payload.Content
				msg.IsEdited = true
			} else if e.Type == "Delete" {
				msg.IsDeleted = true
			}
			err = c.hub.models.Messages.UpdateMessage(ctx, payload.ID, msg)
			if err != nil {
				return msgID, err
			}
		} else {
			if payload.ToRemove {
				err = c.hub.models.Reactions.Delete(ctx, payload.Reaction, payload.ID, c.user.ID)
			} else {
				err = c.hub.models.Reactions.Insert(ctx, payload.Reaction, payload.ID, c.user.ID)
			}
		}

		if err != nil {
			return msgID, err
		}
	}

	return msgID, nil
}

func (c *Client) read() {
	defer func() {
		delete(c.hub.clients, c)
	}()
	for {
		_, wsMsg, err := c.conn.ReadMessage()
		if err != nil {
			log.Println("error: reading ws message:", err)
			break
		}

		var e data.Event
		err = json.Unmarshal(wsMsg, &e)
		if err != nil {
			log.Println("error: unmarshalling ws message as DMEvent:", err)
			break
		}

		if e.Name == "DMEvent" {
			if e.UserID != c.user.ID {
				log.Println("forbidden")
				break
			}

			if len(e.BroadcastTo) != 2 {
				log.Println("error: broadcastTo length mismatch")
				break
			}

			isFriends, err := c.hub.models.Users.IsFriends(context.Background(), e.BroadcastTo)
			if err != nil {
				log.Println("error: checking whether broadcastTo participants are friends", err)
				break
			}
			if !isFriends {
				log.Println("participants should be friends")
				break
			}

			msgID, err := c.save(&e)
			m, err := c.hub.models.Messages.GetMessage(context.Background(), msgID, -1)
			if err != nil {
				log.Println("error: getting message after saving DMEvent:", err)
				break
			}

			if err != nil {
				log.Println("error: saving DMEvent from ws message:", err)
				break
			}

			c.hub.broadcast <- &BroadcastMessage{
				BroadcastTo: e.BroadcastTo,
				Data: map[string]any{
					"name":    "PublishEvent",
					"type":    "DM",
					"payload": m,
				},
			}
		}
	}
}

func (app *application) wsHandler(w http.ResponseWriter, r *http.Request) {
	user := app.getUserContext(r)

	u := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin:     func(r *http.Request) bool { return true },
	}

	conn, err := u.Upgrade(w, r, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	// TODO: adding this closing the connection. figure out why.
	// defer conn.Close()

	client := &Client{
		user: &data.BasicUserResp{
			ID:       user.ID,
			Username: user.Username,
			Avatar:   user.Avatar,
		},
		hub:  app.hub,
		conn: conn,
	}
	app.hub.clients[client] = true

	go client.read()
}
