package server

import (
	"net/http"

	"github.com/zero-color/line-messaging-api-emulator/api"
)

type server struct {
}

func New() api.ServerInterface {
	return &server{}
}

func (s server) PushMessagesByPhone(w http.ResponseWriter, r *http.Request, params api.PushMessagesByPhoneParams) {
	//TODO implement me
	panic("implement me")
}

func (s server) GetWebhookEndpoint(w http.ResponseWriter, r *http.Request) {
	//TODO implement me
	panic("implement me")
}

func (s server) SetWebhookEndpoint(w http.ResponseWriter, r *http.Request) {
	//TODO implement me
	panic("implement me")
}

func (s server) TestWebhookEndpoint(w http.ResponseWriter, r *http.Request) {
	//TODO implement me
	panic("implement me")
}

func (s server) ShowLoadingAnimation(w http.ResponseWriter, r *http.Request) {
	//TODO implement me
	panic("implement me")
}

func (s server) ListCoupon(w http.ResponseWriter, r *http.Request, params api.ListCouponParams) {
	//TODO implement me
	panic("implement me")
}

func (s server) CreateCoupon(w http.ResponseWriter, r *http.Request) {
	//TODO implement me
	panic("implement me")
}

func (s server) GetCouponDetail(w http.ResponseWriter, r *http.Request, couponId string) {
	//TODO implement me
	panic("implement me")
}

func (s server) CloseCoupon(w http.ResponseWriter, r *http.Request, couponId string) {
	//TODO implement me
	panic("implement me")
}

func (s server) GetFollowers(w http.ResponseWriter, r *http.Request, params api.GetFollowersParams) {
	//TODO implement me
	panic("implement me")
}

func (s server) LeaveGroup(w http.ResponseWriter, r *http.Request, groupId string) {
	//TODO implement me
	panic("implement me")
}

func (s server) GetGroupMemberProfile(w http.ResponseWriter, r *http.Request, groupId string, userId string) {
	//TODO implement me
	panic("implement me")
}

func (s server) GetGroupMemberCount(w http.ResponseWriter, r *http.Request, groupId string) {
	//TODO implement me
	panic("implement me")
}

func (s server) GetGroupMembersIds(w http.ResponseWriter, r *http.Request, groupId string, params api.GetGroupMembersIdsParams) {
	//TODO implement me
	panic("implement me")
}

func (s server) GetGroupSummary(w http.ResponseWriter, r *http.Request, groupId string) {
	//TODO implement me
	panic("implement me")
}

func (s server) GetBotInfo(w http.ResponseWriter, r *http.Request) {
	//TODO implement me
	panic("implement me")
}

func (s server) GetMembershipList(w http.ResponseWriter, r *http.Request) {
	//TODO implement me
	panic("implement me")
}

func (s server) GetMembershipSubscription(w http.ResponseWriter, r *http.Request, userId string) {
	//TODO implement me
	panic("implement me")
}

func (s server) GetJoinedMembershipUsers(w http.ResponseWriter, r *http.Request, membershipId int, params api.GetJoinedMembershipUsersParams) {
	//TODO implement me
	panic("implement me")
}

func (s server) GetAggregationUnitUsage(w http.ResponseWriter, r *http.Request) {
	//TODO implement me
	panic("implement me")
}

func (s server) GetAggregationUnitNameList(w http.ResponseWriter, r *http.Request, params api.GetAggregationUnitNameListParams) {
	//TODO implement me
	panic("implement me")
}

func (s server) Broadcast(w http.ResponseWriter, r *http.Request, params api.BroadcastParams) {
	//TODO implement me
	panic("implement me")
}

func (s server) GetNumberOfSentBroadcastMessages(w http.ResponseWriter, r *http.Request, params api.GetNumberOfSentBroadcastMessagesParams) {
	//TODO implement me
	panic("implement me")
}

func (s server) GetNumberOfSentMulticastMessages(w http.ResponseWriter, r *http.Request, params api.GetNumberOfSentMulticastMessagesParams) {
	//TODO implement me
	panic("implement me")
}

func (s server) GetPNPMessageStatistics(w http.ResponseWriter, r *http.Request, params api.GetPNPMessageStatisticsParams) {
	//TODO implement me
	panic("implement me")
}

func (s server) GetNumberOfSentPushMessages(w http.ResponseWriter, r *http.Request, params api.GetNumberOfSentPushMessagesParams) {
	//TODO implement me
	panic("implement me")
}

func (s server) GetNumberOfSentReplyMessages(w http.ResponseWriter, r *http.Request, params api.GetNumberOfSentReplyMessagesParams) {
	//TODO implement me
	panic("implement me")
}

func (s server) MarkMessagesAsRead(w http.ResponseWriter, r *http.Request) {
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

func (s server) GetMessageQuota(w http.ResponseWriter, r *http.Request) {
	//TODO implement me
	panic("implement me")
}

func (s server) GetMessageQuotaConsumption(w http.ResponseWriter, r *http.Request) {
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

func (s server) GetProfile(w http.ResponseWriter, r *http.Request, userId string) {
	//TODO implement me
	panic("implement me")
}

func (s server) CreateRichMenu(w http.ResponseWriter, r *http.Request) {
	//TODO implement me
	panic("implement me")
}

func (s server) CreateRichMenuAlias(w http.ResponseWriter, r *http.Request) {
	//TODO implement me
	panic("implement me")
}

func (s server) GetRichMenuAliasList(w http.ResponseWriter, r *http.Request) {
	//TODO implement me
	panic("implement me")
}

func (s server) DeleteRichMenuAlias(w http.ResponseWriter, r *http.Request, richMenuAliasId string) {
	//TODO implement me
	panic("implement me")
}

func (s server) GetRichMenuAlias(w http.ResponseWriter, r *http.Request, richMenuAliasId string) {
	//TODO implement me
	panic("implement me")
}

func (s server) UpdateRichMenuAlias(w http.ResponseWriter, r *http.Request, richMenuAliasId string) {
	//TODO implement me
	panic("implement me")
}

func (s server) RichMenuBatch(w http.ResponseWriter, r *http.Request) {
	//TODO implement me
	panic("implement me")
}

func (s server) LinkRichMenuIdToUsers(w http.ResponseWriter, r *http.Request) {
	//TODO implement me
	panic("implement me")
}

func (s server) UnlinkRichMenuIdFromUsers(w http.ResponseWriter, r *http.Request) {
	//TODO implement me
	panic("implement me")
}

func (s server) GetRichMenuList(w http.ResponseWriter, r *http.Request) {
	//TODO implement me
	panic("implement me")
}

func (s server) GetRichMenuBatchProgress(w http.ResponseWriter, r *http.Request, params api.GetRichMenuBatchProgressParams) {
	//TODO implement me
	panic("implement me")
}

func (s server) ValidateRichMenuObject(w http.ResponseWriter, r *http.Request) {
	//TODO implement me
	panic("implement me")
}

func (s server) ValidateRichMenuBatchRequest(w http.ResponseWriter, r *http.Request) {
	//TODO implement me
	panic("implement me")
}

func (s server) DeleteRichMenu(w http.ResponseWriter, r *http.Request, richMenuId string) {
	//TODO implement me
	panic("implement me")
}

func (s server) GetRichMenu(w http.ResponseWriter, r *http.Request, richMenuId string) {
	//TODO implement me
	panic("implement me")
}

func (s server) GetRichMenuImage(w http.ResponseWriter, r *http.Request, richMenuId string) {
	//TODO implement me
	panic("implement me")
}

func (s server) SetRichMenuImage(w http.ResponseWriter, r *http.Request, richMenuId string) {
	//TODO implement me
	panic("implement me")
}

func (s server) LeaveRoom(w http.ResponseWriter, r *http.Request, roomId string) {
	//TODO implement me
	panic("implement me")
}

func (s server) GetRoomMemberProfile(w http.ResponseWriter, r *http.Request, roomId string, userId string) {
	//TODO implement me
	panic("implement me")
}

func (s server) GetRoomMemberCount(w http.ResponseWriter, r *http.Request, roomId string) {
	//TODO implement me
	panic("implement me")
}

func (s server) GetRoomMembersIds(w http.ResponseWriter, r *http.Request, roomId string, params api.GetRoomMembersIdsParams) {
	//TODO implement me
	panic("implement me")
}

func (s server) CancelDefaultRichMenu(w http.ResponseWriter, r *http.Request) {
	//TODO implement me
	panic("implement me")
}

func (s server) GetDefaultRichMenuId(w http.ResponseWriter, r *http.Request) {
	//TODO implement me
	panic("implement me")
}

func (s server) SetDefaultRichMenu(w http.ResponseWriter, r *http.Request, richMenuId string) {
	//TODO implement me
	panic("implement me")
}

func (s server) IssueLinkToken(w http.ResponseWriter, r *http.Request, userId string) {
	//TODO implement me
	panic("implement me")
}

func (s server) UnlinkRichMenuIdFromUser(w http.ResponseWriter, r *http.Request, userId string) {
	//TODO implement me
	panic("implement me")
}

func (s server) GetRichMenuIdOfUser(w http.ResponseWriter, r *http.Request, userId string) {
	//TODO implement me
	panic("implement me")
}

func (s server) LinkRichMenuIdToUser(w http.ResponseWriter, r *http.Request, userId string, richMenuId string) {
	//TODO implement me
	panic("implement me")
}

var _ api.ServerInterface = (*server)(nil)
