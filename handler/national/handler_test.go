package national

import (
	"encoding/json"
	"github.com/gorilla/websocket"
	"github.com/matnich89/network-rail-client/model/realtime"
	"github.com/matnich89/trainstats-realtime/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

type MockClient struct {
	mock.Mock
}

func (m *MockClient) SubRTPPM() (chan *realtime.RTPPMDataMsg, error) {
	args := m.Called()
	return args.Get(0).(chan *realtime.RTPPMDataMsg), args.Error(1)
}

func TestNewHandler(t *testing.T) {
	mockClient := new(MockClient)
	mockChan := make(chan *realtime.RTPPMDataMsg)
	mockClient.On("SubRTPPM").Return(mockChan, nil)

	handler, err := NewHandler(mockClient)

	assert.NoError(t, err)
	assert.NotNil(t, handler)
	assert.Equal(t, mockClient, handler.nrClient)
	assert.Equal(t, mockChan, handler.realTimeDataChan)
	assert.NotNil(t, handler.latestNationalRailData)
}

func TestListen(t *testing.T) {
	mockClient := new(MockClient)
	mockChan := make(chan *realtime.RTPPMDataMsg)
	mockClient.On("SubRTPPM").Return(mockChan, nil)

	handler, err := NewHandler(mockClient)
	require.NoError(t, err)

	go handler.Listen()

	mockChan <- &realtime.RTPPMDataMsg{
		RTPPMDataMsgV1: realtime.RTPPMDataMsgV1{
			RTPPMData: realtime.RTPPMData{
				NationalPage: realtime.NationalPage{
					NationalPPM: realtime.NationalPPM{
						OnTime:         "80",
						CancelVeryLate: "10",
						Late:           "20",
						Total:          "100",
					},
				},
			},
		},
	}

	time.Sleep(100 * time.Millisecond)

	assert.Equal(t, 80, handler.latestNationalRailData.OnTime)
	assert.Equal(t, 10, handler.latestNationalRailData.CancelledOrVeryLate)
	assert.Equal(t, 10, handler.latestNationalRailData.Late)
	assert.Equal(t, 100, handler.latestNationalRailData.Total)
}

func TestHandleNationalData(t *testing.T) {
	mockClient := new(MockClient)
	mockChan := make(chan *realtime.RTPPMDataMsg)
	mockClient.On("SubRTPPM").Return(mockChan, nil)

	handler, err := NewHandler(mockClient)
	require.NoError(t, err)

	server := httptest.NewServer(http.HandlerFunc(handler.HandleNationalData))
	defer server.Close()

	wsURL := "ws" + server.URL[4:]
	ws, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	require.NoError(t, err)
	defer ws.Close()

	_, message, err := ws.ReadMessage()
	require.NoError(t, err)
	assert.Equal(t, "Connected to WebSocket server", string(message))

	_, message, err = ws.ReadMessage()
	require.NoError(t, err)

	var receivedData model.NationalData
	err = json.Unmarshal(message, &receivedData)
	require.NoError(t, err)

	assert.Equal(t, handler.latestNationalRailData, &receivedData)
}

func TestBuildNationalRailData(t *testing.T) {
	testCases := []struct {
		name     string
		input    realtime.NationalPPM
		expected *model.NationalData
		hasError bool
	}{
		{
			name: "Valid input",
			input: realtime.NationalPPM{
				OnTime:         "80",
				CancelVeryLate: "10",
				Late:           "20",
				Total:          "100",
			},
			expected: &model.NationalData{
				OnTime:              80,
				CancelledOrVeryLate: 10,
				Late:                10,
				Total:               100,
			},
			hasError: false,
		},
		{
			name: "Invalid OnTime",
			input: realtime.NationalPPM{
				OnTime:         "invalid",
				CancelVeryLate: "10",
				Late:           "20",
				Total:          "100",
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
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expected, result)
			}
		})
	}
}
