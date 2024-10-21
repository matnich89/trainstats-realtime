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
	NationalChan chan *realtime.NationalPPM
}

func NewNetworkRail(client Client) (*NetworkRail, error) {
	rtppmChannel, err := client.SubRTPPM()
	if err != nil {
		return nil, err
	}
	return &NetworkRail{
		client:       client,
		DataChan:     rtppmChannel,
		NationalChan: make(chan *realtime.NationalPPM, 10),
	}, nil
}

func (n *NetworkRail) ProcessData(shutdown <-chan struct{}) {
	for {
		select {
		case data := <-n.DataChan:
			log.Println("processing...")
			n.processNational(data)
		case <-shutdown:
			log.Println("process data stopped")
			return
		}
	}
}

func (n *NetworkRail) processNational(data *realtime.RTPPMDataMsg) {
	nationalData := data.RTPPMDataMsgV1.RTPPMData.NationalPage.NationalPPM
	n.NationalChan <- &nationalData
}
