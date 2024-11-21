package mysql_repo

import (
	"context"
	"github.com/linchengzhi/lottery/domain/entity"
	"gorm.io/gorm"
)

type UserAssetRecordRepo struct {
	db *gorm.DB
}

func NewUserAssetRecordRepo(db *gorm.DB) UserAssetRecordRepo {
	return UserAssetRecordRepo{
		db: db,
	}
}

func (u *UserAssetRecordRepo) GetByRequestID(ctx context.Context, requestId string) (*entity.UserAssetRecord, error) {
	var record entity.UserAssetRecord
	err := u.db.WithContext(ctx).Where("request_id = ?", requestId).First(&record).Error
	if err != nil {
		return nil, err
	}
	return &record, nil
}
