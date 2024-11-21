package mysql_repo

import (
	"context"
	"github.com/bytedance/sonic"
	"github.com/linchengzhi/lottery/domain/cerror"
	"github.com/linchengzhi/lottery/domain/entity"
	"gorm.io/gorm"
	"time"
)

type UserItemRepo struct {
	db *gorm.DB
}

func NewUserItemRepo(db *gorm.DB) UserItemRepo {
	return UserItemRepo{db: db}
}

func (r *UserItemRepo) Create(ctx context.Context, userId int64, items map[int64]int64) error {
	for itemID, num := range items {
		userItem := &entity.UserItem{
			UserID: userId,
			ItemID: itemID,
			Num:    num,
		}
		if err := r.db.WithContext(ctx).Create(userItem).Error; err != nil {
			return err
		}
	}
	return nil
}

func (r *UserItemRepo) List(ctx context.Context, userId int64) (map[int64]int64, error) {
	var userItems []*entity.UserItem
	if err := r.db.WithContext(ctx).Where("user_id = ?", userId).Find(&userItems).Error; err != nil {
		return nil, err
	}
	var result = make(map[int64]int64)
	for _, userItem := range userItems {
		result[userItem.ItemID] = userItem.Num
	}
	return result, nil
}

func (r *UserItemRepo) Update(ctx context.Context, userId int64, items map[int64]int64, requestId string, requestTime time.Time) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var userItems []entity.UserItem
		itemIDs := make([]int64, 0, len(items))
		for itemId := range items {
			itemIDs = append(itemIDs, itemId)
		}

		// 查询所有相关物品
		if err := tx.Where("user_id = ? AND item_id IN ?", userId, itemIDs).Find(&userItems).Error; err != nil {
			return err
		}

		// 创建一个map用于快速查找已存在的物品
		existingItemsMap := make(map[int64]*entity.UserItem)
		for i := range userItems {
			existingItemsMap[userItems[i].ItemID] = &userItems[i]
		}

		// 更新内存中的数量或新增物品
		for itemId, change := range items {
			if item, exists := existingItemsMap[itemId]; exists {
				newNum := item.Num + change
				if newNum < 0 {
					return cerror.ErrItemLess
				}
				item.Num = newNum
			} else {
				if change < 0 {
					return cerror.ErrItemLess
				}
				userItems = append(userItems, entity.UserItem{
					UserID: userId,
					ItemID: itemId,
					Num:    change,
				})
			}
		}

		// 批量保存更新和新增的物品
		if err := tx.Save(&userItems).Error; err != nil {
			return err
		}

		// 创建物品变更记录
		itemsJSON, err := sonic.Marshal(items)
		if err != nil {
			return err
		}

		itemRecord := entity.UserItemRecord{
			UserID:      userId,
			Items:       string(itemsJSON),
			CreatedAt:   time.Now(),
			RequestID:   requestId,
			RequestTime: requestTime,
		}
		if err = tx.Create(&itemRecord).Error; err != nil {
			return err
		}
		return nil
	})
}
