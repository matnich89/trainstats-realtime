package national

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/matnich89/network-rail-client/model/realtime"
	"github.com/matnich89/trainstats-realtime/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewHandler(t *testing.T) {
	ch := make(chan *realtime.NationalPPM)
	handler := NewHandler(ch)

	assert.NotNil(t, handler)
	assert.Equal(t, ch, handler.nationalDataChan)
	assert.NotNil(t, handler.latestNationalRailData)
}

func TestHandler_Listen(t *testing.T) {
	ch := make(chan *realtime.NationalPPM)
	handler := NewHandler(ch)

	shutdownCh := make(chan struct{})
	go handler.Listen(shutdownCh)

	// Send test data
	testData := &realtime.NationalPPM{
		Total:          "100",
		OnTime:         "80",
		Late:           "15",
		CancelVeryLate: "5",
	}
	ch <- testData

	// Allow some time for processing
	time.Sleep(100 * time.Millisecond)

	// Check if the data was processed correctly
	assert.Equal(t, 80, handler.latestNationalRailData.OnTime)
	assert.Equal(t, 5, handler.latestNationalRailData.CancelledOrVeryLate)
	assert.Equal(t, 10, handler.latestNationalRailData.Late)
	assert.Equal(t, 100, handler.latestNationalRailData.Total)

	// Test shutdown
	close(shutdownCh)
	time.Sleep(100 * time.Millisecond)
}

func TestHandler_HandleNationalData(t *testing.T) {
	ch := make(chan *realtime.NationalPPM)
	handler := NewHandler(ch)

	server := httptest.NewServer(http.HandlerFunc(handler.HandleNationalData))
	defer server.Close()

	// Replace "http" with "ws" in the URL
	wsURL := "ws" + server.URL[4:]

	// Connect to the WebSocket server
	ws, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	require.NoError(t, err)
	defer ws.Close()

	_, msg, err := ws.ReadMessage()
	require.NoError(t, err)
	assert.Equal(t, "Connected to WebSocket server", string(msg))

	_, msg, err = ws.ReadMessage()
	require.NoError(t, err)

	var receivedData model.NationalData
	err = json.Unmarshal(msg, &receivedData)
	require.NoError(t, err)

	assert.Equal(t, handler.latestNationalRailData, &receivedData)
}

func TestBuildNationalRailData(t *testing.T) {
	testCases := []struct {
		name     string
		input    *realtime.NationalPPM
		expected *model.NationalData
		hasError bool
	}{
		{
			name: "Valid input",
			input: &realtime.NationalPPM{
				Total:          "100",
				OnTime:         "80",
				Late:           "15",
				CancelVeryLate: "5",
			},
			expected: &model.NationalData{
				Total:               100,
				OnTime:              80,
				Late:                10,
				CancelledOrVeryLate: 5,
			},
			hasError: false,
		},
		{
			name: "Invalid input",
			input: &realtime.NationalPPM{
				Total:          "invalid",
				OnTime:         "80",
				Late:           "15",
				CancelVeryLate: "5",
			},
			expected: nil,
			hasError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := buildNationalRailData(tc.input)

			if tc.hasError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expected, result)
			}
		})
	}
}
