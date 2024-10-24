package service

import (
	"github.com/matnich89/network-rail-client/model/realtime"
	"log"
)

type Client interface {
	SubRTPPM() (chan *realtime.RTPPMDataMsg, error)
}

type NetworkRail struct {
	client       Client
	DataChan     chan *realtime.RTPPMDataMsg
	OperatorChan chan []realtime.OperatorPage
}

func NewNetworkRail(client Client) (*NetworkRail, error) {
	rtppmChannel, err := client.SubRTPPM()
	if err != nil {
		return nil, err
	}
	return &NetworkRail{
		client:       client,
		DataChan:     rtppmChannel,
		OperatorChan: make(chan []realtime.OperatorPage, 10),
	}, nil
}

func (n *NetworkRail) ProcessData(shutdown <-chan struct{}) {
	for {
		select {
		case data := <-n.DataChan:
			log.Println("processing...")
			n.processOperator(data)
		case <-shutdown:
			log.Println("process data stopped")
			return
		}
	}
}

func (n *NetworkRail) processOperator(data *realtime.RTPPMDataMsg) {
	operatorData := data.RTPPMDataMsgV1.RTPPMData.OperatorPage
	n.OperatorChan <- operatorData
}
