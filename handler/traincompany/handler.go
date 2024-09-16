package traincompany

import (
	"context"
	"encoding/json"
	"github.com/go-redis/redis/v8"
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
	trainCompanyDataChan chan *realtime.OperatorData
	networkRailService   RailService
	redisClient          *redis.Client
}

func NewHandler(trainCompanyDataChan chan *realtime.OperatorData, redisClient *redis.Client) *Handler {

	return &Handler{
		trainCompanyDataChan: trainCompanyDataChan,
		redisClient:          redisClient,
	}
}

func (h *Handler) Listen(shutdownCh <-chan struct{}) {
	ctx := context.Background()
	for {
		select {
		case data := <-h.trainCompanyDataChan:
			operatorData, err := buildTrainOperatorData(data)
			if err != nil {
				log.Println("Error building train data", err)
				continue
			}
			h.updateOperatorData(ctx, operatorData)
		case <-shutdownCh:
			log.Println("Shutting down national data handler")
			return
		}
	}
}

func (h *Handler) updateOperatorData(ctx context.Context, data *model.TrainOperator) {
	// Store operator data
	jsonData, err := json.Marshal(data)
	if err != nil {
		log.Println("Error marshalling operator data:", err)
		return
	}
	err = h.redisClient.Set(ctx, "operator:"+data.Name, jsonData, 0).Err()
	if err != nil {
		log.Println("Error storing operator data in Redis:", err)
		return
	}

	// Update league table
	err = h.redisClient.ZAdd(ctx, "leagueTable", &redis.Z{
		Score:  float64(data.Percentage),
		Member: data.Name,
	}).Err()
	if err != nil {
		log.Println("Error updating league table in Redis:", err)
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

	ctx := context.Background()
	for {
		data, err := h.getLatestData(ctx)
		if err != nil {
			log.Println("Error getting latest data:", err)
			continue
		}

		jsonData, err := json.Marshal(data)
		if err != nil {
			log.Println("Error marshalling data:", err)
			continue
		}

		if err := conn.WriteMessage(websocket.TextMessage, jsonData); err != nil {
			log.Println("Error sending message to client:", err)
			return
		}

		time.Sleep(5 * time.Second)
	}
}

func (h *Handler) getLatestData(ctx context.Context) (map[string]interface{}, error) {
	// Get all operator data
	keys, err := h.redisClient.Keys(ctx, "operator:*").Result()
	if err != nil {
		return nil, err
	}

	currentData := make(map[string]*model.TrainOperator)
	for _, key := range keys {
		jsonData, err := h.redisClient.Get(ctx, key).Result()
		if err != nil {
			log.Println("Error getting operator data from Redis:", err)
			continue
		}

		var operator model.TrainOperator
		err = json.Unmarshal([]byte(jsonData), &operator)
		if err != nil {
			log.Println("Error unmarshalling operator data:", err)
			continue
		}

		currentData[operator.Name] = &operator
	}

	// Get league table
	leagueTable, err := h.redisClient.ZRevRangeWithScores(ctx, "leagueTable", 0, -1).Result()
	if err != nil {
		return nil, err
	}

	leagueTableData := make([]map[string]interface{}, len(leagueTable))
	for i, z := range leagueTable {
		leagueTableData[i] = map[string]interface{}{
			"name":       z.Member,
			"percentage": z.Score,
		}
	}

	return map[string]interface{}{
		"currentData": currentData,
		"leagueTable": leagueTableData,
	}, nil
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

	percentage, err := strconv.Atoi(ppm.PPM.Text)
	if err != nil {
		return nil, err
	}

	return &model.TrainOperator{
		Name:                ppm.Name,
		Total:               total,
		Late:                late,
		CancelledOrVeryLate: cancelledOrVeryLate,
		Percentage:          percentage,
		OnTime:              onTime,
	}, nil
}
