package server

import (
	"net/http"

	"github.com/zero-color/line-messaging-api-emulator/api/messagingapi"
)

func (s server) GetAggregationUnitUsage(w http.ResponseWriter, r *http.Request) {
	//TODO implement me
	panic("implement me")
}

func (s server) GetAggregationUnitNameList(w http.ResponseWriter, r *http.Request, params messagingapi.GetAggregationUnitNameListParams) {
	//TODO implement me
	panic("implement me")
}

func (s server) GetNumberOfSentBroadcastMessages(w http.ResponseWriter, r *http.Request, params messagingapi.GetNumberOfSentBroadcastMessagesParams) {
	//TODO implement me
	panic("implement me")
}

func (s server) GetNumberOfSentMulticastMessages(w http.ResponseWriter, r *http.Request, params messagingapi.GetNumberOfSentMulticastMessagesParams) {
	//TODO implement me
	panic("implement me")
}

func (s server) GetPNPMessageStatistics(w http.ResponseWriter, r *http.Request, params messagingapi.GetPNPMessageStatisticsParams) {
	//TODO implement me
	panic("implement me")
}

func (s server) GetNumberOfSentPushMessages(w http.ResponseWriter, r *http.Request, params messagingapi.GetNumberOfSentPushMessagesParams) {
	//TODO implement me
	panic("implement me")
}

func (s server) GetNumberOfSentReplyMessages(w http.ResponseWriter, r *http.Request, params messagingapi.GetNumberOfSentReplyMessagesParams) {
	//TODO implement me
	panic("implement me")
}
