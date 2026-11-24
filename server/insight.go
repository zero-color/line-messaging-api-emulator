package server

import (
	"context"
	"fmt"

	"github.com/samber/lo"
	"github.com/zero-color/line-messaging-api-emulator/api/insightapi"
	"github.com/zero-color/line-messaging-api-emulator/db"
	"github.com/zero-color/line-messaging-api-emulator/internal/auth"
)

// GetFriendsDemographics reports that demographics aren't available,
// as the LINE Platform does for accounts with too few friends to aggregate.
func (s *server) GetFriendsDemographics(ctx context.Context, request insightapi.GetFriendsDemographicsRequestObject) (insightapi.GetFriendsDemographicsResponseObject, error) {
	return insightapi.GetFriendsDemographics200JSONResponse{Available: lo.ToPtr(false)}, nil
}

// GetNumberOfFollowers gets the number of followers.
// It reports the current number for any date, because the emulator keeps no follower history.
func (s *server) GetNumberOfFollowers(ctx context.Context, request insightapi.GetNumberOfFollowersRequestObject) (insightapi.GetNumberOfFollowersResponseObject, error) {
	followers, err := db.FromContext(ctx).Queries().GetBotFollowerCount(ctx, auth.GetBotID(ctx))
	if err != nil {
		return nil, fmt.Errorf("failed to count followers: %w", err)
	}
	return insightapi.GetNumberOfFollowers200JSONResponse{
		Status:          lo.ToPtr(insightapi.GetNumberOfFollowersResponseStatusReady),
		Followers:       &followers,
		TargetedReaches: &followers,
		Blocks:          lo.ToPtr(int64(0)),
	}, nil
}

// GetNumberOfMessageDeliveries reports that delivery statistics aren't aggregated yet
func (s *server) GetNumberOfMessageDeliveries(ctx context.Context, request insightapi.GetNumberOfMessageDeliveriesRequestObject) (insightapi.GetNumberOfMessageDeliveriesResponseObject, error) {
	return insightapi.GetNumberOfMessageDeliveries200JSONResponse{
		Status: lo.ToPtr(insightapi.GetNumberOfMessageDeliveriesResponseStatusUnready),
	}, nil
}

// GetMessageEvent reports no interaction statistics, since the emulator doesn't track impressions or clicks
func (s *server) GetMessageEvent(ctx context.Context, request insightapi.GetMessageEventRequestObject) (insightapi.GetMessageEventResponseObject, error) {
	return insightapi.GetMessageEvent200JSONResponse{
		Overview: &insightapi.GetMessageEventResponseOverview{},
		Messages: &[]insightapi.GetMessageEventResponseMessage{},
		Clicks:   &[]insightapi.GetMessageEventResponseClick{},
	}, nil
}

// GetStatisticsPerUnit reports no statistics for the aggregation unit.
// The values are null, as on the LINE Platform when too few users are involved to aggregate.
func (s *server) GetStatisticsPerUnit(ctx context.Context, request insightapi.GetStatisticsPerUnitRequestObject) (insightapi.GetStatisticsPerUnitResponseObject, error) {
	return insightapi.GetStatisticsPerUnit200JSONResponse{
		Overview: insightapi.GetStatisticsPerUnitResponseOverview{},
		Messages: []insightapi.GetStatisticsPerUnitResponseMessage{},
		Clicks:   []insightapi.GetStatisticsPerUnitResponseClick{},
	}, nil
}
