package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/linchengzhi/goany"
	"github.com/linchengzhi/lottery/api/http/middleware"
	"github.com/linchengzhi/lottery/domain/cerror"
	"github.com/linchengzhi/lottery/usecase/asset_uc"
	"go.uber.org/zap"
)

type AssetHdr struct {
	assetUc asset_uc.AssetUc
	log     *zap.Logger
}

func NewAssetHandler(uc asset_uc.AssetUc, log *zap.Logger) AssetHdr {
	return AssetHdr{
		uc,
		log,
	}
}

func (hdr *AssetHdr) GetAsset(c *gin.Context) (interface{}, error) {
	span, err := middleware.StartSpanFromContext(c, "GetAsset")
	if err != nil {
		return nil, err
	}
	defer span.Finish()
	// 从body中读取用户id
	uid, _ := c.GetPostForm("user_id")
	if uid == "" {
		return nil, cerror.ErrLogout
	}
	userId := goany.ToInt64(uid)
	hdr.log.Info("获取用户资产", zap.Int64("userId", userId))
	resp, err := hdr.assetUc.GetAsset(c, userId)
	if err != nil {
		hdr.log.Error("获取用户资产失败", zap.Int64("userId", userId), zap.Any("error", err))
		return nil, err
	}
	hdr.log.Debug("获取奖品列表成功", zap.Any("resp", resp))
	return resp, nil
}

func (hdr *AssetHdr) ListItem(c *gin.Context) (interface{}, error) {
	uid, _ := c.GetPostForm("user_id")
	if uid == "" {
		return nil, cerror.ErrLogout
	}
	userId := goany.ToInt64(uid)
	hdr.log.Info("获取用户物品", zap.Int64("userId", userId))
	resp, err := hdr.assetUc.ListItem(c, userId)
	if err != nil {
		hdr.log.Error("获取用户物品失败", zap.Int64("userId", userId), zap.Any("error", err))
		return nil, err
	}
	hdr.log.Debug("获取用户物品成功", zap.Any("resp", resp))
	return resp, nil
}
