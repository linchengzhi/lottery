package entity

import (
	"context"
	"time"
)

type UserItem struct {
	ID     int64 `gorm:"primaryKey;autoIncrement;comment:'用户物品id'" json:"id"`
	UserID int64 `gorm:"not null;index:idx_user_id;comment:'用户ID'" json:"user_id"`
	ItemID int64 `gorm:"not null;comment:'物品id'" json:"item_id"`
	Num    int64 `gorm:"not null;comment:'数量'" json:"num"`
}

func (u *UserItem) TableName() string {
	return "user_item"
}

type IUserItemRepo interface {
	Create(ctx context.Context, userId int64, items map[int64]int64) error
	List(ctx context.Context, userId int64) ([]*UserItem, error, error)
	Update(ctx context.Context, userId int64, items map[int64]int64, requestId string, requestTime time.Time) error //同时更新物品表和插入记录表
}
