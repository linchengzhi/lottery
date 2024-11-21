package entity

import (
	"context"
	"github.com/linchengzhi/lottery/domain/dto"
	"time"
)

const (
	TNLotteryDrawRecord = "lottery_draw_record"
)

type LotteryDrawRecord struct {
	ID         int64     `gorm:"primaryKey;autoIncrement;comment:'抽奖记录ID'" json:"id"`
	ActivityID int64     `gorm:"not null;comment:'活动ID'" json:"activity_id"`
	UserID     int64     `gorm:"not null;comment:'用户ID'" json:"user_id"`
	DrawCount  int       `gorm:"not null;default:1;comment:'抽奖次数，例如1次或10次抽奖'" json:"draw_count"`
	CreatedAt  time.Time `gorm:"not null;comment:'记录创建时间'" json:"created_at"`
	RequestID  string    `gorm:"size:36;uniqueIndex:uniq_request_id;comment:'请求ID，用于幂等'" json:"request_id"`
}

type ILotteryDrawRecordRepo interface {
	CreateTable(ctx context.Context, activityId int64) error
	TableName(activityId int64) string
	Create(ctx context.Context, drawRecord *LotteryDrawRecord, prizes []*dto.Item) error
	BatchCreate(ctx context.Context, drawRecords []*LotteryDrawRecord, prizeRecords []*LotteryPrizeRecord) error
}
