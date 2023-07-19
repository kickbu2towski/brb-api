package main

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/kickbu2towski/brb-api/internal/data"
)

type Hub struct {
	clients map[*Client]bool
	broadcast chan data.Message
}

func NewHub() *Hub {
	return &Hub{
		clients: make(map[*Client]bool),
		broadcast: make(chan data.Message),
	}
}

func (h *Hub) run() {
	for {
		msg := <-h.broadcast
		for client := range h.clients {
			err := client.conn.WriteJSON(msg)
			if err != nil {
				log.Println("error writing ws json:", err)
				delete(h.clients, client)
			}
		}
	}
}

type Client struct {
	hub *Hub
	conn *websocket.Conn
}

func (c *Client) read() {
	for {
		var msg data.Message
		err := c.conn.ReadJSON(&msg)
		if err != nil {
			log.Println("error reading ws json:", err)
			delete(c.hub.clients, c)
			break
		}
		c.hub.broadcast <- msg
	}
}

func (app *application) wsHandler(w http.ResponseWriter, r *http.Request) {
	u := websocket.Upgrader{
		ReadBufferSize: 1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool { return true },
	}

	conn, err := u.Upgrade(w, r, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	defer conn.Close()

	client := &Client{
		hub: app.hub,
		conn: conn,
	}
	app.hub.clients[client] = true

	go client.read()
}
