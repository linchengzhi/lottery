package mysql_repo

import (
	"context"
	"fmt"
	"github.com/linchengzhi/lottery/domain/entity"
	"gorm.io/gorm"
	"strings"
)

type LotteryPrizeRecordRepo struct {
	db *gorm.DB
}

// NewLotteryPrizeRecordRepo 创建 LotteryPrizeRecordRepo 实例
func NewLotteryPrizeRecordRepo(db *gorm.DB) LotteryPrizeRecordRepo {
	return LotteryPrizeRecordRepo{db: db}
}

// CreateTable 创建奖品记录表 table_name = "lottery_prize_record_" + activityId
func (r *LotteryPrizeRecordRepo) CreateTable(ctx context.Context, activityId int64) error {
	tableName := r.TableName(activityId)
	if err := r.db.WithContext(ctx).Table(tableName).Migrator().CreateTable(&entity.LotteryPrizeRecord{}); err != nil {
		//如果表已存在，则忽略错误
		if strings.Contains(err.Error(), "already exists") {
			return nil
		}
		return fmt.Errorf("failed to create table %s: %w", tableName, err)
	}
	return nil
}

func (r *LotteryPrizeRecordRepo) TableName(activityId int64) string {
	tableName := fmt.Sprintf("%s_%d", entity.TNLotteryPrizeRecord, activityId)
	return tableName
}

// ListByUserId 根据 user_id 查询奖品记录，支持分页
func (r *LotteryPrizeRecordRepo) ListByUserId(ctx context.Context, activityId, userId int64, page, pageSize int) ([]*entity.LotteryPrizeRecord, error) {
	tableName := r.TableName(activityId)
	var prizeRecords []*entity.LotteryPrizeRecord
	// 查询奖品记录
	if err := r.db.WithContext(ctx).Table(tableName).
		Where("user_id = ?", userId).
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&prizeRecords).Error; err != nil {
		return nil, err
	}
	return prizeRecords, nil
}
