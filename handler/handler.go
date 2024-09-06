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
	"time"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type Client interface {
	SubRTPPM() (chan *realtime.RTPPMDataMsg, error)
}

type Handler struct {
	nrClient         Client
	realTimeDataChan chan *realtime.RTPPMDataMsg
	latestData       *model.NationalData
}

func NewHandler(nrClient *client.NetworkRailClient) (*Handler, error) {
	realTimeDataChan, err := nrClient.SubRTPPM()
	if err != nil {
		return nil, err
	}
	latestData := &model.NationalData{
		OnTime:              0,
		CancelledOrVeryLate: 0,
		Late:                0,
		Total:               0,
	}
	return &Handler{nrClient: nrClient, realTimeDataChan: realTimeDataChan, latestData: latestData}, nil
}

func (h *Handler) Listen() {
	go func() {
		for {
			select {
			case data := <-h.realTimeDataChan:
				nrData, err := buildNationalRailData(data.RTPPMDataMsgV1.RTPPMData.NationalPage.NationalPPM)
				if err != nil {
					log.Println("Error building NationalRailData:", err)
				} else {
					h.latestData = nrData
				}
			case <-time.After(5 * time.Minute):
				log.Println("Warning: No data received in the last 5 minutes, Topic could be down")
			}
		}
	}()
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
		b, err := json.Marshal(h.latestData)
		if err != nil {
			log.Println("Error marshalling data:", err)
		}
		if err := conn.WriteMessage(websocket.TextMessage, b); err != nil {
			log.Println("Error sending message to client:", err)
			return
		}
	}
}

func buildNationalRailData(ppm realtime.NationalPPM) (*model.NationalData, error) {
	onTime, err := strconv.Atoi(ppm.OnTime)
	if err != nil {
		return nil, err
	}

	canclledOrVeryLate, err := strconv.Atoi(ppm.CancelVeryLate)
	if err != nil {
		return nil, err
	}

	late, err := strconv.Atoi(ppm.Late)
	if err != nil {
		return nil, err
	}
	late = late - canclledOrVeryLate

	total, err := strconv.Atoi(ppm.Total)
	if err != nil {
		return nil, err
	}

	return &model.NationalData{
		OnTime:              onTime,
		CancelledOrVeryLate: canclledOrVeryLate,
		Late:                late,
		Total:               total,
	}, nil

}
