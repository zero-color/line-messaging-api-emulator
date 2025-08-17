package server

import (
	"net/http"

	"github.com/zero-color/line-messaging-api-emulator/api/messagingapi"
)

func (s server) ListCoupon(w http.ResponseWriter, r *http.Request, params messagingapi.ListCouponParams) {
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
