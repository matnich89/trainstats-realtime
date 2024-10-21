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

	initialData, err := json.Marshal(h.latestNationalRailData)
	if err != nil {
		log.Println("Error marshalling initial data:", err)
		return
	}
	if err := conn.WriteMessage(websocket.TextMessage, initialData); err != nil {
		log.Println("Error sending initial data:", err)
		return
	}

	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()

	done := make(chan struct{})
	go func() {
		defer close(done)
		for {
			if _, _, err := conn.NextReader(); err != nil {
				log.Println("Client disconnected:", err)
				return
			}
		}
	}()

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
		case <-done:
			// Client has disconnected
			log.Println("Client has disconnected :(")
			return
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

	// Calculate percentages
	onTimePercentage := float64(onTime) / float64(total) * 100
	cancelledOrVeryLatePercentage := float64(cancelledOrVeryLate) / float64(total) * 100
	latePercentage := float64(late) / float64(total) * 100

	return &model.NationalData{
		OnTime:                        onTime,
		CancelledOrVeryLate:           cancelledOrVeryLate,
		Late:                          late,
		Total:                         total,
		OnTimePercentage:              onTimePercentage,
		CancelledOrVeryLatePercentage: cancelledOrVeryLatePercentage,
		LatePercentage:                latePercentage,
	}, nil
}
