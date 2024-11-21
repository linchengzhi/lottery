package entity

import "time"

type UserItemRecord struct {
	ID          int64     `gorm:"primaryKey;autoIncrement;comment:'用户物品变更记录ID'" json:"id"`
	UserID      int64     `gorm:"not null;comment:'用户ID'" json:"user_id"`
	Items       string    `gorm:"type:json;not null;comment:'变更物品'" json:"items"`
	CreatedAt   time.Time `gorm:"not null;comment:'创建时间'" json:"created_at"`
	RequestID   string    `gorm:"size:36;uniqueIndex:uniq_request_id;comment:'请求ID，用于幂等'" json:"request_id"`
	RequestTime time.Time `gorm:"not null;comment:'请求时间'" json:"request_time"`
}

func (u *UserItemRecord) TableName() string {
	return "user_item_record"
}
