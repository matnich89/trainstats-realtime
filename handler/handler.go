package handler

import (
	"github.com/gorilla/websocket"
	"github.com/matnich89/network-rail-client/client"
	"log"
	"net/http"
	"time"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type Handler struct {
	nrClient *client.NetworkRailClient
}

func NewHandler(nrClient *client.NetworkRailClient) *Handler {
	return &Handler{nrClient: nrClient}
}

func (h *Handler) ConnectWS(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Error upgrading connection:", err)
		return
	}
	defer conn.Close()

	if err := conn.WriteMessage(websocket.TextMessage, []byte("Connected to WebSocket server")); err != nil {
		log.Println("Error sending initial message:", err)
		return
	}

	/*
		 In the service broadcasting would be invoked by data being received
		from network rail client, this is just for the template, so we can
		confirm the calling client is receiving data from the websocket
	*/
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := conn.WriteMessage(websocket.TextMessage, []byte("Periodic update: train stuff")); err != nil {
				log.Println("Error sending periodic message:", err)
				return
			}
		}
	}
}
