package dto

import "time"

type DrawReq struct {
	RequestId   string     `json:"request_id"`
	RequestTime time.Time  `json:"request_time"`
	UserId      int64      `json:"user_id"`
	ActivityId  int64      `json:"activity_id"`
	DrawNum     int64      `json:"draw_num"`
	PrizesData  *PrizeData `json:"prizes_data"`
}

type DrawResp struct {
	RequestId string     `json:"request_id"`
	PrizeData *PrizeData `json:"prize_data"`
	Err       error      `json:"err"`
}

type ListPrizeReq struct {
	ActivityId int64 `json:"activity_id"`
	UserId     int64 `json:"user_id"`
	Page       int   `json:"page"`
	PageSize   int   `json:"page_size"`
}
