package handler

import (
	"encoding/json"
	"github.com/gorilla/websocket"
	"github.com/matnich89/network-rail-client/client"
	"github.com/matnich89/network-rail-client/model/realtime"
	"github.com/matnich89/trainstats-service-template/model"
	"log"
	"net/http"
	"strconv"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type Handler struct {
	nrClient         *client.NetworkRailClient
	realTimeDataChan chan *realtime.RTPPMDataMsg
}

func NewHandler(nrClient *client.NetworkRailClient) (*Handler, error) {
	realTimeDataChan, err := nrClient.SubRTPPM()
	if err != nil {
		return nil, err
	}
	return &Handler{nrClient: nrClient, realTimeDataChan: realTimeDataChan}, nil
}

func (h *Handler) HandleNationalData(w http.ResponseWriter, r *http.Request) {
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

	for {
		select {
		case data := <-h.realTimeDataChan:
			log.Println("received data")
			nrData, err := buildNationalRailData(data.RTPPMDataMsgV1.RTPPMData.NationalPage.NationalPPM)
			if err != nil {
				log.Println("Error building National Rail Data:", err)
			}
			b, err := json.Marshal(nrData)
			if err != nil {
				log.Println("Error marshalling data:", err)
			}
			if err := conn.WriteMessage(websocket.TextMessage, b); err != nil {
				log.Println("Error sending periodic message:", err)
				return
			}
		}
	}
}

func buildNationalRailData(ppm realtime.NationalPPM) (model.NationalData, error) {
	onTime, err := strconv.Atoi(ppm.OnTime)
	if err != nil {
		return model.NationalData{}, err
	}

	canclledOrVeryLate, err := strconv.Atoi(ppm.CancelVeryLate)
	if err != nil {
		return model.NationalData{}, err
	}

	late, err := strconv.Atoi(ppm.Late)
	if err != nil {
		return model.NationalData{}, err
	}

	total, err := strconv.Atoi(ppm.Total)
	if err != nil {
		return model.NationalData{}, err
	}

	return model.NationalData{
		OnTime:              onTime,
		CancelledOrVeryLate: canclledOrVeryLate,
		Late:                late,
		Total:               total,
	}, nil

}
