package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/kickbu2towski/brb-api/internal/data"
)

type Hub struct {
	clients   map[*Client]bool
	broadcast chan map[string]any
	models    *data.Models
}

func NewHub(models *data.Models) *Hub {
	return &Hub{
		clients:   make(map[*Client]bool),
		broadcast: make(chan map[string]any),
		models:    models,
	}
}

func (h *Hub) run() {
	for {
		msg := <-h.broadcast
		for client := range h.clients {
			bc, _ := GetBroadcastTo(msg)
			allowed := Includes(bc, client.user.ID)
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

func (c *Client) save(e *data.Event) error {
	b, err := json.Marshal(e.Payload)
	if err != nil {
		log.Println("error marshalling payload:", err)
		return err
	}

	switch e.Type {
	case "Create":
		var m data.Message
		m.UserID = c.user.ID
		err := json.Unmarshal(b, &m)
		if err != nil {
			log.Println("error: dm payload: unmarshalling:", err)
			return err
		}
		err = c.hub.models.Messages.InsertMessage(context.Background(), &m)
		if err != nil {
			log.Println("error: dm payload: inserting:", err)
			return err
		}
	case "Edit", "Delete", "Reaction":
		ctx := context.Background()

		var payload struct {
			ID        string
			Content   string
			Reactions string
		}

		err := json.Unmarshal(b, &payload)
		if err != nil {
			log.Println("error: edit payload: unmarshalling:", err)
			return err
		}

		msg, err := c.hub.models.Messages.GetMessage(ctx, payload.ID)
		if err != nil {
			log.Println("error: edit payload: getting message:", err)
			return err
		}

		if e.Type == "Edit" {
			msg.Content = payload.Content
			msg.IsEdited = true
		} else if e.Type == "Reaction" {
			var reactions map[string][]string
			err := json.Unmarshal([]byte(payload.Reactions), &reactions)
			if err != nil {
				log.Println("error: unmarshalling reactions:", err)
				return err
			}
			msg.Reactions = reactions
		} else if e.Type == "Delete" {
			msg.IsDeleted = true
		}

		err = c.hub.models.Messages.UpdateMessage(ctx, payload.ID, msg)
		if err != nil {
			log.Println("error: edit payload: updating message:", err)
			return err
		}
	}

	return nil
}

func (c *Client) read() {
	defer func() {
		delete(c.hub.clients, c)
	}()
	for {
		var msg map[string]any
		err := c.conn.ReadJSON(&msg)
		if err != nil {
			log.Println("error: reading ws json:", err)
			break
		}

		e, ok := (msg["event"]).(string)
		if ok && e == "DMEvent" {
			m, _ := GetMessageEvent(msg)
			err = c.save(m)
			if err != nil {
				log.Println("error: saving ws json:", err)
				break
			}
			c.hub.broadcast <- msg
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
