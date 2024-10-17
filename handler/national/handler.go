package national

import (
	"encoding/json"
	"github.com/gorilla/websocket"
	"github.com/matnich89/network-rail-client/model/realtime"
	"github.com/matnich89/trainstats-realtime/model"
	"log"
	"net/http"
	"strconv"
	"time"
)

type RailService interface {
	ProcessData(shutdown <-chan struct{})
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type Handler struct {
	nationalDataChan       chan *realtime.NationalPPM
	latestNationalRailData *model.NationalData
	networkRailService     RailService
}

func NewHandler(nationalDataChan chan *realtime.NationalPPM) *Handler {
	latestData := &model.NationalData{
		OnTime:              0,
		CancelledOrVeryLate: 0,
		Late:                0,
		Total:               0,
	}
	return &Handler{latestNationalRailData: latestData, nationalDataChan: nationalDataChan}
}

func (h *Handler) Listen(shutdownCh <-chan struct{}) {
	for {
		select {
		case data := <-h.nationalDataChan:
			nrData, err := buildNationalRailData(data)
			if err != nil {
				log.Println("Error building NationalRailData:", err)
			} else {
				h.latestNationalRailData = nrData
			}
		case <-shutdownCh:
			log.Println("Shutting down national data handler")
			return
		}
	}
}

func (h *Handler) HandleNationalData(w http.ResponseWriter, r *http.Request) {
	log.Println("connection made...")
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Error upgrading connection:", err)
		return
	}
	defer conn.Close()

	if err := conn.WriteMessage(websocket.TextMessage, []byte("{}")); err != nil {
		log.Println("Error sending initial message:", err)
		return
	}

	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			b, err := json.Marshal(h.latestNationalRailData)
			if err != nil {
				log.Println("Error marshalling data:", err)
				continue
			}
			if err := conn.WriteMessage(websocket.TextMessage, b); err != nil {
				log.Println("Error sending message to client:", err)
				return
			}
		}
	}
}

func buildNationalRailData(ppm *realtime.NationalPPM) (*model.NationalData, error) {
	onTime, err := strconv.Atoi(ppm.OnTime)
	if err != nil {
		return nil, err
	}

	cancelledOrVeryLate, err := strconv.Atoi(ppm.CancelVeryLate)
	if err != nil {
		return nil, err
	}

	late, err := strconv.Atoi(ppm.Late)
	if err != nil {
		return nil, err
	}
	late = late - cancelledOrVeryLate

	total, err := strconv.Atoi(ppm.Total)
	if err != nil {
		return nil, err
	}

	return &model.NationalData{
		OnTime:              onTime,
		CancelledOrVeryLate: cancelledOrVeryLate,
		Late:                late,
		Total:               total,
	}, nil
}
