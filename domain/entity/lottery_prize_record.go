package entity

import (
	"context"
	"time"
)

const (
	TNLotteryPrizeRecord = "lottery_prize_record"
)

type LotteryPrizeRecord struct {
	ID         int64     `gorm:"primaryKey;autoIncrement;comment:'奖品记录ID'" json:"id"`
	ActivityID int64     `gorm:"not null;comment:'活动ID'" json:"activity_id"`
	UserID     int64     `gorm:"not null;comment:'用户ID'" json:"user_id"`
	PrizeID    int64     `gorm:"not null;comment:'奖品ID'" json:"prize_id"`
	PrizeNum   int64     `gorm:"not null;comment:'奖品数量'" json:"prize_num"`
	CreatedAt  time.Time `gorm:"autoCreateTime;comment:'创建时间'" json:"created_at"`
}

type ILotteryPrizeRecordRepo interface {
	CreateTable(ctx context.Context, activityId int64) error
	TableName(activityId int64) string
	ListByUserId(ctx context.Context, activityId, userId int64, page, pageSize int) ([]*LotteryPrizeRecord, error)
}
