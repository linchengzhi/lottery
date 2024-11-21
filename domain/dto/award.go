package dto

import "time"

type AwardStream struct {
	RequestId   string     `json:"request_id"`
	RequestTime time.Time  `json:"request_time"`
	PrizeData   *PrizeData `json:"prize_data"`
}

type PrizeData struct {
	UserId     int64   `json:"user_id"`
	ActivityId int64   `json:"activity_id"`
	Prizes     []*Item `json:"prize_ids"`
	Amount     int64   `json:"amount"`
}

type Item struct {
	Id  int64 `json:"id" yaml:"id"`   // 奖品ID  固定为一个
	Num int64 `json:"num" yaml:"num"` // 奖品数量
}

// 奖池奖品
type Prize struct {
	Id     int64 `json:"id" yaml:"id"`         // 奖品ID  固定为一个
	Num    int64 `json:"num" yaml:"num"`       // 奖品数量
	Weight int64 `json:"weight" yaml:"weight"` // 奖品的权重，用于随机
}

// 星级奖品
type StarLevel struct {
	Level  int      `json:"level" yaml:"level"`   // 星级，例如1，2，3
	Weight int64    `json:"weight" yaml:"weight"` // 星级的权重 weight/sum(weight)
	Prizes []*Prize `json:"prizes" yaml:"prizes"` // 该星级下的奖品列表
}

// 奖池配置
type PrizePool struct {
	Prizes []*StarLevel `json:"prizes" yaml:"prizes"` // 奖品配置
}
