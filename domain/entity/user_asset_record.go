package entity

import (
	"context"
	"time"
)

type UserAssetRecord struct {
	ID          int64     `gorm:"primaryKey;autoIncrement;comment:'资产变更记录ID'" json:"id"`
	UserID      int64     `gorm:"not null;comment:'用户ID'" json:"user_id"`
	Gold        int64     `gorm:"not null;comment:'金币'" json:"gold"`
	Stone       int64     `gorm:"not null;comment:'原石'" json:"stone"`
	Crystal     int64     `gorm:"not null;comment:'创世结晶'" json:"crystal"`
	CreatedAt   time.Time `gorm:"not null;comment:'创建时间'" json:"created_at"`
	RequestID   string    `gorm:"size:36;uniqueIndex:uniq_request_id;comment:'请求ID，用于幂等'" json:"request_id"`
	RequestTime time.Time `gorm:"not null;comment:'请求时间'" json:"request_time"`
}

func (u *UserAssetRecord) TableName() string {
	return "user_asset_record"
}

type IAssetTransactionRepo interface {
	//通过requestId查询
	GetByRequestID(ctx context.Context, requestId string) (*UserAssetRecord, error)
	//Insert(ctx context.Context, at *AssetTransaction) error //与asset一并插入
}
