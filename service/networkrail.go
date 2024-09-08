package service

import (
	"github.com/matnich89/network-rail-client/client"
	"github.com/matnich89/network-rail-client/model/realtime"
)

type NetworkRail struct {
	client            *client.NetworkRailClient
	DataChan          chan *realtime.RTPPMDataMsg
	NationalChan      chan *realtime.NationalPPM
	TrainOperatorChan chan *realtime.OperatorData
}

func NewNetworkRail(client *client.NetworkRailClient) (*NetworkRail, error) {
	rtppmChannel, err := client.SubRTPPM()
	if err != nil {
		return nil, err
	}
	return &NetworkRail{
		client:            client,
		DataChan:          rtppmChannel,
		NationalChan:      make(chan *realtime.NationalPPM),
		TrainOperatorChan: make(chan *realtime.OperatorData),
	}, nil
}

func (n *NetworkRail) ProcessData() {
	for {
		select {
		case data := <-n.DataChan:
			n.processNational(data)
		}
	}
}

func (n *NetworkRail) processNational(data *realtime.RTPPMDataMsg) {
	nationalData := data.RTPPMDataMsgV1.RTPPMData.NationalPage.NationalPPM
	n.NationalChan <- &nationalData
}

func (n *NetworkRail) processTrainOperator(data *realtime.RTPPMDataMsg) {
	operatorPages := data.RTPPMDataMsgV1.RTPPMData.OperatorPage
	for _, val := range operatorPages {
		n.TrainOperatorChan <- &val.Operator
	}
}
