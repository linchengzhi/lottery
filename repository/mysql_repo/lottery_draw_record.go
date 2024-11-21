package mysql_repo

import (
	"context"
	"fmt"
	"github.com/linchengzhi/lottery/domain/dto"
	"github.com/linchengzhi/lottery/domain/entity"
	"gorm.io/gorm"
	"strings"
	"time"
)

type LotteryDrawRecordRepo struct {
	db *gorm.DB
}

func NewLotteryDrawRecordRepo(db *gorm.DB) LotteryDrawRecordRepo {
	return LotteryDrawRecordRepo{db: db}
}

// CreateTable 创建奖品记录表 table_name = "lottery_prize_record_" + activityId
func (r *LotteryDrawRecordRepo) CreateTable(ctx context.Context, activityId int64) error {
	tableName := r.TableName(activityId)
	if err := r.db.WithContext(ctx).Table(tableName).Migrator().CreateTable(&entity.LotteryDrawRecord{}); err != nil {
		//如果表已存在，则忽略错误
		if strings.Contains(err.Error(), "already exists") {
			return nil
		}
		return fmt.Errorf("failed to create table %s: %w", tableName, err)
	}
	return nil
}

func (r *LotteryDrawRecordRepo) TableName(activityId int64) string {
	tableName := fmt.Sprintf("%s_%d", entity.TNLotteryDrawRecord, activityId)
	return tableName
}

// Create 插入抽奖记录和奖品记录
func (r *LotteryDrawRecordRepo) Create(ctx context.Context, drawRecord *entity.LotteryDrawRecord, prizes []*dto.Item) error {
	// 使用事务保证同时插入抽奖记录和奖品记录的原子性
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 1. 插入抽奖记录
		now := time.Now()
		drawRecord.CreatedAt = now
		if err := tx.Table(r.TableName(drawRecord.ActivityID)).Create(drawRecord).Error; err != nil {
			return err
		}

		// 2. 插入奖品记录
		prizeRecords := make([]*entity.LotteryPrizeRecord, 0)
		for _, prize := range prizes {
			prizeRecords = append(prizeRecords, &entity.LotteryPrizeRecord{
				ActivityID: drawRecord.ActivityID,
				UserID:     drawRecord.UserID,
				PrizeID:    prize.Id,
				PrizeNum:   prize.Num,
				CreatedAt:  now,
			})
		}
		lp := NewLotteryPrizeRecordRepo(tx)
		return tx.Table(lp.TableName(drawRecord.ActivityID)).Create(prizeRecords).Error
	})
	return err
}

func (r *LotteryDrawRecordRepo) BatchCreate(ctx context.Context, drawRecords []*entity.LotteryDrawRecord, prizeRecords []*entity.LotteryPrizeRecord) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Table(r.TableName(drawRecords[0].ActivityID)).Create(drawRecords).Error; err != nil {
			return err
		}
		lp := NewLotteryPrizeRecordRepo(tx)
		return tx.Table(lp.TableName(drawRecords[0].ActivityID)).Create(prizeRecords).Error
	})
}
