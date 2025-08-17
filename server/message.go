package server

import (
	"net/http"

	"github.com/zero-color/line-messaging-api-emulator/api"
)

func (s server) Broadcast(w http.ResponseWriter, r *http.Request, params api.BroadcastParams) {
	//TODO implement me
	panic("implement me")
}

func (s server) Multicast(w http.ResponseWriter, r *http.Request, params api.MulticastParams) {
	//TODO implement me
	panic("implement me")
}

func (s server) Narrowcast(w http.ResponseWriter, r *http.Request, params api.NarrowcastParams) {
	//TODO implement me
	panic("implement me")
}

func (s server) GetNarrowcastProgress(w http.ResponseWriter, r *http.Request, params api.GetNarrowcastProgressParams) {
	//TODO implement me
	panic("implement me")
}

func (s server) PushMessage(w http.ResponseWriter, r *http.Request, params api.PushMessageParams) {
	//TODO implement me
	panic("implement me")
}

func (s server) PushMessagesByPhone(w http.ResponseWriter, r *http.Request, params api.PushMessagesByPhoneParams) {
	//TODO implement me
	panic("implement me")
}

func (s server) ReplyMessage(w http.ResponseWriter, r *http.Request) {
	//TODO implement me
	panic("implement me")
}

func (s server) ValidateBroadcast(w http.ResponseWriter, r *http.Request) {
	//TODO implement me
	panic("implement me")
}

func (s server) ValidateMulticast(w http.ResponseWriter, r *http.Request) {
	//TODO implement me
	panic("implement me")
}

func (s server) ValidateNarrowcast(w http.ResponseWriter, r *http.Request) {
	//TODO implement me
	panic("implement me")
}

func (s server) ValidatePush(w http.ResponseWriter, r *http.Request) {
	//TODO implement me
	panic("implement me")
}

func (s server) ValidateReply(w http.ResponseWriter, r *http.Request) {
	//TODO implement me
	panic("implement me")
}

func (s server) GetMessageContent(w http.ResponseWriter, r *http.Request, messageId string) {
	//TODO implement me
	panic("implement me")
}

func (s server) GetMessageContentPreview(w http.ResponseWriter, r *http.Request, messageId string) {
	//TODO implement me
	panic("implement me")
}

func (s server) GetMessageContentTranscodingByMessageId(w http.ResponseWriter, r *http.Request, messageId string) {
	//TODO implement me
	panic("implement me")
}

func (s server) MarkMessagesAsRead(w http.ResponseWriter, r *http.Request) {
	//TODO implement me
	panic("implement me")
}

func (s server) ShowLoadingAnimation(w http.ResponseWriter, r *http.Request) {
	//TODO implement me
	panic("implement me")
}
