package service

import (
	"testing"
	"time"

	"github.com/matnich89/network-rail-client/model/realtime"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockClient struct {
	mock.Mock
}

func (m *MockClient) SubRTPPM() (chan *realtime.RTPPMDataMsg, error) {
	args := m.Called()
	return args.Get(0).(chan *realtime.RTPPMDataMsg), args.Error(1)
}

func TestNewNetworkRail(t *testing.T) {
	mockClient := new(MockClient)
	mockChannel := make(chan *realtime.RTPPMDataMsg)

	mockClient.On("SubRTPPM").Return(mockChannel, nil)

	networkRail, err := NewNetworkRail(mockClient)

	assert.NoError(t, err)
	assert.NotNil(t, networkRail)
	assert.Equal(t, mockClient, networkRail.client)
	assert.Equal(t, mockChannel, networkRail.DataChan)
	assert.NotNil(t, networkRail.NationalChan)

	mockClient.AssertExpectations(t)
}

func TestNetworkRail_ProcessData(t *testing.T) {
	mockClient := new(MockClient)
	mockChannel := make(chan *realtime.RTPPMDataMsg)

	mockClient.On("SubRTPPM").Return(mockChannel, nil)

	networkRail, _ := NewNetworkRail(mockClient)

	shutdownCh := make(chan struct{})
	go networkRail.ProcessData(shutdownCh)

	// Create test data
	testData := &realtime.RTPPMDataMsg{
		RTPPMDataMsgV1: realtime.RTPPMDataMsgV1{
			RTPPMData: realtime.RTPPMData{
				NationalPage: realtime.NationalPage{
					NationalPPM: realtime.NationalPPM{
						Total: "100",
					},
				},
				OperatorPage: []realtime.OperatorPage{
					{
						Operator: realtime.OperatorData{
							Code: "TEST",
						},
					},
				},
			},
		},
	}

	// Send test data
	networkRail.DataChan <- testData

	select {
	case nationalData := <-networkRail.NationalChan:
		assert.Equal(t, "100", nationalData.Total)
	case <-time.After(time.Second):
		t.Fatal("Timeout waiting for national data")
	}
	// Test shutdown
	close(shutdownCh)
	time.Sleep(100 * time.Millisecond) // Give some time for the goroutine to stop
}

func TestNetworkRail_processNational(t *testing.T) {
	networkRail := &NetworkRail{
		NationalChan: make(chan *realtime.NationalPPM, 1),
	}

	testData := &realtime.RTPPMDataMsg{
		RTPPMDataMsgV1: realtime.RTPPMDataMsgV1{
			RTPPMData: realtime.RTPPMData{
				NationalPage: realtime.NationalPage{
					NationalPPM: realtime.NationalPPM{
						Total: "100",
					},
				},
			},
		},
	}

	networkRail.processNational(testData)

	select {
	case nationalData := <-networkRail.NationalChan:
		assert.Equal(t, "100", nationalData.Total)
	case <-time.After(time.Second):
		t.Fatal("Timeout waiting for national data")
	}
}
