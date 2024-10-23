package national

import (
	"encoding/json"
	"github.com/gorilla/websocket"
	"github.com/matnich89/network-rail-client/model/realtime"
	"github.com/matnich89/trainstats-realtime/model"
	"log"
	"math"
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
	operatorDataChan        chan []realtime.OperatorPage
	latestPassengerRailData *model.PerformanceData
	latestFreightRailData   *model.PerformanceData
	networkRailService      RailService
	ticker                  *time.Ticker
}

func NewHandler(passengerDataChan chan []realtime.OperatorPage) *Handler {
	latestFrieghtData := &model.PerformanceData{
		OnTime:              0,
		CancelledOrVeryLate: 0,
		Late:                0,
		Total:               0,
	}

	latestPassengerData := &model.PerformanceData{
		OnTime:              0,
		CancelledOrVeryLate: 0,
		Late:                0,
		Total:               0,
	}

	return &Handler{
		latestFreightRailData:   latestFrieghtData,
		latestPassengerRailData: latestPassengerData,
		ticker:                  time.NewTicker(15 * time.Second),
		operatorDataChan:        passengerDataChan,
	}
}

func (h *Handler) Listen(shutdownCh <-chan struct{}) {
	for {
		select {
		case passengerData := <-h.operatorDataChan:
			paData, err := buildNationalPassengerData(passengerData)
			if err != nil {
				log.Println("Error building NationalPassengerData:", err)
			} else {
				h.latestPassengerRailData = paData
			}
		case <-shutdownCh:
			log.Println("Shutting down national data handler")
			return
		}
	}
}

func (h *Handler) HandlePassengerData(w http.ResponseWriter, r *http.Request) {
	log.Println("connection made...")
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Error upgrading connection:", err)
		return
	}
	defer conn.Close()

	initialData, err := json.Marshal(h.latestPassengerRailData)
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
			b, err := json.Marshal(h.latestPassengerRailData)
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

func buildNationalPassengerData(operatorData []realtime.OperatorPage) (*model.PerformanceData, error) {
	var onTimeTotal, cancelledVeryLateTotal, lateTotal, fullTotal int
	var bestOperator, worstOperator *model.OperatorPerformance
	bestScore := math.MaxFloat64
	worstScore := -1.0

	for _, op := range operatorData {

		if val, ok := model.TrainOperators[op.Operator.Code]; ok {
			if val.Carries != model.Passenger {
				continue
			}
		}

		onTime, err := strconv.Atoi(op.Operator.OnTime)
		if err != nil {
			log.Println("Error converting onTime:", err)
			continue
		}

		cancelledVeryLate, err := strconv.Atoi(op.Operator.CancelVeryLate)
		if err != nil {
			log.Println("Error converting cancelVeryLate:", err)
			continue
		}

		late, err := strconv.Atoi(op.Operator.Late)
		if err != nil {
			log.Println("Error converting late:", err)
			continue
		}

		total, err := strconv.Atoi(op.Operator.Total)
		if err != nil {
			log.Println("Error converting total:", err)
			continue
		}

		operatorPerf := model.OperatorPerformance{
			Code:                op.Operator.Code,
			Name:                op.Operator.Name,
			OnTime:              onTime,
			Late:                late - cancelledVeryLate,
			CancelledOrVeryLate: cancelledVeryLate,
			Total:               total,
		}

		if total > 0 {
			operatorPerf.OnTimePercentage = float64(onTime) / float64(total) * 100
			operatorPerf.LatePercentage = float64(late-cancelledVeryLate) / float64(total) * 100
			operatorPerf.CancelledOrVeryLatePercentage = float64(cancelledVeryLate) / float64(total) * 100
		}

		score := calculatePerformanceScore(operatorPerf)
		operatorPerf.PerformanceScore = score

		if score < bestScore && total >= 10 { // Minimum threshold to avoid operators with very few trains
			bestScore = score
			bestOperator = &operatorPerf
		}
		if score > worstScore && total >= 10 {
			worstScore = score
			worstOperator = &operatorPerf
		}

		onTimeTotal += onTime
		cancelledVeryLateTotal += cancelledVeryLate
		lateTotal += late
		fullTotal += total
	}

	onTimePercentage := float64(onTimeTotal) / float64(fullTotal) * 100
	cancelledOrVeryLatePercentage := float64(cancelledVeryLateTotal) / float64(fullTotal) * 100
	latePercentage := float64(lateTotal) / float64(fullTotal) * 100

	if bestOperator != nil && worstOperator != nil {
		return &model.PerformanceData{
			OnTime:                        onTimeTotal,
			CancelledOrVeryLate:           cancelledVeryLateTotal,
			Late:                          lateTotal - cancelledVeryLateTotal,
			Total:                         fullTotal,
			OnTimePercentage:              onTimePercentage,
			CancelledOrVeryLatePercentage: cancelledOrVeryLatePercentage,
			LatePercentage:                latePercentage,
			BestOperator:                  bestOperator,
			WorstOperator:                 worstOperator,
		}, nil
	}

	return &model.PerformanceData{
		OnTime:                        onTimeTotal,
		CancelledOrVeryLate:           cancelledVeryLateTotal,
		Late:                          lateTotal - cancelledVeryLateTotal,
		Total:                         fullTotal,
		OnTimePercentage:              onTimePercentage,
		CancelledOrVeryLatePercentage: cancelledOrVeryLatePercentage,
		LatePercentage:                latePercentage,
	}, nil
}

func calculatePerformanceScore(op model.OperatorPerformance) float64 {
	if op.Total == 0 {
		return math.MaxFloat64
	}

	// Weight cancelled/very late more heavily than regular late trains
	cancelledWeight := 2.0
	lateWeight := 1.0

	score := (float64(op.CancelledOrVeryLate)*cancelledWeight +
		float64(op.Late-op.CancelledOrVeryLate)*lateWeight) / float64(op.Total)

	return score
}
