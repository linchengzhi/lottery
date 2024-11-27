package mysql_repo

import (
	"context"
	"github.com/linchengzhi/lottery/domain/cerror"
	"github.com/linchengzhi/lottery/domain/entity"
	"github.com/pkg/errors"
	"gorm.io/gorm"
	"time"
)

// UserAssetRepo 实现 IAssetRepo 接口
type UserAssetRepo struct {
	db *gorm.DB
}

// NewUserAssetRepo 创建 UserAssetRepo 实例
func NewUserAssetRepo(db *gorm.DB) UserAssetRepo {
	return UserAssetRepo{db: db}
}

// Create 创建资产表
func (r *UserAssetRepo) Create(ctx context.Context, at *entity.UserAsset) error {
	return r.db.WithContext(ctx).Table(at.TableName()).Create(&at).Error
}

// Get 根据 user_id 获取资产信息
func (r *UserAssetRepo) Get(ctx context.Context, userId int64) (*entity.UserAsset, error) {
	var asset entity.UserAsset
	// 使用 Gorm 查询资产表
	userAsset := entity.UserAsset{UserID: userId}
	if err := r.db.WithContext(ctx).Table(userAsset.TableName()).Where("user_id = ?", userId).First(&asset).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) { // 用户不存在, 自动创建资产表
			asset = entity.UserAsset{UserID: userId}
			if err := r.db.WithContext(ctx).Create(&asset).Error; err != nil {
				return nil, errors.Wrap(err, "create asset_rd error")
			}
		}
		return nil, err // 其他错误返回
	}
	return &asset, nil
}

// Update 更新资产表和插入资产交易表
func (r *UserAssetRepo) Update(ctx context.Context, at *entity.UserAsset, requestId string, requestTime time.Time) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 更新资产表，确保更新后的值不小于零
		result := tx.Table(at.TableName()).Model(&entity.UserAsset{}).
			Where("user_id = ? AND gold + ? >= 0 AND stone + ? >= 0 AND crystal + ? >= 0",
				at.UserID, at.Gold, at.Stone, at.Crystal).
			Updates(map[string]interface{}{
				"gold":    gorm.Expr("gold + ?", at.Gold),
				"stone":   gorm.Expr("stone + ?", at.Stone),
				"crystal": gorm.Expr("crystal + ?", at.Crystal),
			})

		if result.Error != nil {
			return result.Error
		}

		if result.RowsAffected == 0 {
			return cerror.ErrAssetLess
		}

		// 插入资产变更记录
		assetRecord := entity.UserAssetRecord{
			UserID:      at.UserID,
			Gold:        at.Gold,
			Stone:       at.Stone,
			Crystal:     at.Crystal,
			CreatedAt:   time.Now(),
			RequestID:   requestId,
			RequestTime: requestTime,
		}
		if err := tx.Table(assetRecord.TableName()).Create(&assetRecord).Error; err != nil {
			return err
		}
		return nil
	})
}
