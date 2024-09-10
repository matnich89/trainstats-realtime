package traincompany

import (
	"github.com/gorilla/websocket"
	"github.com/matnich89/network-rail-client/model/realtime"
	"github.com/matnich89/trainstats-realtime/model"
	"log"
	"net/http"
	"strconv"
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
	trainCompanyDataChan chan *realtime.OperatorData
	networkRailService   RailService
}

func NewHandler(trainCompanyDataChan chan *realtime.OperatorData) *Handler {
	return &Handler{trainCompanyDataChan: trainCompanyDataChan}
}

func (h *Handler) Listen(shutdownCh <-chan struct{}) {
	for {
		select {
		case data := <-h.trainCompanyDataChan:
			operatorData, err := buildTrainOperatorData(data)
			if err != nil {
				log.Println("Error building train data", err)
			}
			log.Println(operatorData.Name + ":" + strconv.Itoa(operatorData.Percentage))
		case <-shutdownCh:
			log.Println("Shutting down national data handler")
			return
		}
	}
}

func (h *Handler) HandleOperatorData(w http.ResponseWriter, r *http.Request) {
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
		if err != nil {
			log.Println("Error marshalling data:", err)
		}
		if err := conn.WriteMessage(websocket.TextMessage, []byte("todo")); err != nil {
			log.Println("Error sending message to client:", err)
			return
		}
	}
}

func buildTrainOperatorData(ppm *realtime.OperatorData) (*model.TrainOperator, error) {
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
	return &model.TrainOperator{
		Name:                ppm.Name,
		Total:               total,
		Late:                0,
		CancelledOrVeryLate: 0,
		Percentage:          0,
		Position:            0,
		OnTime:              onTime,
	}, nil
}
