package waMsgTransport

import (
	"github.com/jrevanaldi-ai/gowa/proto/armadilloutil"
	"github.com/jrevanaldi-ai/gowa/proto/instamadilloTransportPayload"
	"github.com/jrevanaldi-ai/gowa/proto/waMsgApplication"
)

const (
	FBMessageApplicationVersion = 2
	IGMessageApplicationVersion = 3
)

func (msg *MessageTransport_Payload) DecodeFB() (*waMsgApplication.MessageApplication, error) {
	return armadilloutil.Unmarshal(&waMsgApplication.MessageApplication{}, msg.GetApplicationPayload(), FBMessageApplicationVersion)
}

func (msg *MessageTransport_Payload) DecodeIG() (*instamadilloTransportPayload.TransportPayload, error) {
	return armadilloutil.Unmarshal(&instamadilloTransportPayload.TransportPayload{}, msg.GetApplicationPayload(), IGMessageApplicationVersion)
}
