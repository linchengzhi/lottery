package entity

import (
	"context"
	"fmt"
	"time"
)

type UserAsset struct {
	ID      int64 `gorm:"primaryKey;autoIncrement;comment:'资产表id'" json:"id"`
	UserID  int64 `gorm:"not null;uniqueIndex:uniq_user_id;comment:'用户ID'" json:"user_id"`
	Gold    int64 `gorm:"not null;comment:'金币'" json:"gold"`
	Stone   int64 `gorm:"not null;comment:'原石'" json:"stone"`
	Crystal int64 `gorm:"not null;comment:'创世结晶'" json:"crystal"`
}

// TableName 实现动态表名
func (u *UserAsset) TableName() string {
	return fmt.Sprintf("user_asset_%d", u.UserID%10)
}

type IAssetRepo interface {
	Create(ctx context.Context, at *UserAsset) error
	Get(ctx context.Context, userId int64) (*UserAsset, error)
	Update(ctx context.Context, at *UserAsset, requestId string, requestTime time.Time) error //同时插入资产交易表和更新资产表
}
