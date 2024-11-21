package asset_uc

import (
	"context"
	"github.com/linchengzhi/lottery/domain/entity"
	"github.com/linchengzhi/lottery/repository/mysql_repo"
	"github.com/linchengzhi/lottery/repository/redis_repo"
	"go.uber.org/zap"
	"time"
)

type IAssetUc interface {
	//创建资产
	CreateAsset(ctx context.Context, userId int64) (*entity.UserAsset, error)
	// 获取资产
	GetAsset(ctx context.Context, userID int64) (*entity.UserAsset, error)
	// 获取资产记录
	GetAssetRecord(ctx context.Context, requestId string) (*entity.UserAssetRecord, error)
	// 获取物品
	ListItem(ctx context.Context, userID int64) (map[int64]int64, error)
	// 更新资产
	UpdateAsset(ctx context.Context, asset *entity.UserAsset, requestId string, requestTime time.Time) error
	// 更新物品
	UpdateItems(ctx context.Context, userId int64, items map[int64]int64, requestId string, requestTime time.Time) error
}

type AssetUc struct {
	log *zap.Logger

	assetCache redis_repo.UserAssetCache
	itemCache  redis_repo.UserItemCache

	assetRepo   mysql_repo.UserAssetRepo
	assetRecord mysql_repo.UserAssetRecordRepo
	itemRepo    mysql_repo.UserItemRepo
}

func NewAssetUc(log *zap.Logger, repoMysql mysql_repo.RepoMysql, repoRedis redis_repo.RepoRedis) AssetUc {
	return AssetUc{
		log:        log,
		assetCache: repoRedis.UserAssetCache,
		itemCache:  repoRedis.UserItemCache,
		assetRepo:  repoMysql.UserAssetRepo,
		itemRepo:   repoMysql.UserItemRepo,
	}
}

func (uc *AssetUc) CreateAsset(ctx context.Context, userId int64) (*entity.UserAsset, error) {
	// 创建资产
	asset := &entity.UserAsset{
		UserID: userId,
	}
	err := uc.assetRepo.Create(ctx, asset)
	if err != nil {
		uc.log.Error("创建资产执行数据库失败", zap.Error(err))
		return nil, err
	}
	return asset, nil
}

func (uc *AssetUc) GetAsset(ctx context.Context, userID int64) (*entity.UserAsset, error) {
	// 尝试从缓存中获取
	asset, err := uc.assetCache.Get(ctx, userID)
	if err != nil {
		uc.log.Error("获取资产读取redis错误", zap.Error(err))
		return nil, err
	}

	if asset != nil {
		return asset, nil
	}

	// 缓存未命中，从数据库获取
	asset, err = uc.assetRepo.Get(ctx, userID)
	if err != nil {
		uc.log.Error("获取资产读取数据库错误", zap.Error(err))
		return nil, err
	}

	// 设置到缓存
	if err = uc.assetCache.Set(ctx, userID, asset); err != nil {
		uc.log.Warn("获取资产设置缓存错误", zap.Error(err))
	}
	return asset, nil
}

func (uc *AssetUc) GetAssetRecord(ctx context.Context, requestId string) (*entity.UserAssetRecord, error) {
	return uc.assetRecord.GetByRequestID(ctx, requestId)
}

func (uc *AssetUc) ListItem(ctx context.Context, userID int64) (map[int64]int64, error) {
	// 尝试从缓存中获取
	items, err := uc.itemCache.Get(ctx, userID)
	if err != nil {
		uc.log.Error("获取物品读取redis错误", zap.Error(err))
		return nil, err
	}

	if items != nil {
		return items, nil
	}

	// 缓存未命中，从数据库获取
	items, err = uc.itemRepo.List(ctx, userID)
	if err != nil {
		uc.log.Error("获取物品读取数据库错误", zap.Error(err))
		return nil, err
	}

	// 设置到缓存
	if err = uc.itemCache.Set(ctx, userID, items); err != nil {
		uc.log.Warn("获取物品设置缓存错误", zap.Error(err))
	}
	return items, nil
}

func (uc *AssetUc) UpdateAsset(ctx context.Context, asset *entity.UserAsset, requestId string, requestTime time.Time) error {
	if asset.Gold == 0 && asset.Stone == 0 && asset.Crystal == 0 {
		return nil
	}
	// 更新数据库
	err := uc.assetRepo.Update(ctx, asset, requestId, requestTime)
	if err != nil {
		uc.log.Error("更新资产执行数据库失败", zap.Error(err))
		return err
	}

	// 删除缓存
	if err = uc.assetCache.Delete(ctx, asset.UserID); err != nil {
		uc.log.Warn("更新资产删除缓存失败", zap.Error(err))
	}
	return nil
}

func (uc *AssetUc) UpdateItems(ctx context.Context, userId int64, items map[int64]int64, requestId string, requestTime time.Time) error {
	// 更新数据库
	err := uc.itemRepo.Update(ctx, userId, items, requestId, requestTime)
	if err != nil {
		uc.log.Error("更新物品执行数据库失败", zap.Error(err))
		return err
	}

	// 删除缓存
	if err = uc.itemCache.Delete(ctx, userId); err != nil {
		uc.log.Warn("更新资产删除缓存失败", zap.Error(err))
	}
	return nil
}
