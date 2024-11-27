package handler

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/linchengzhi/lottery/api/http/middleware"
	"github.com/linchengzhi/lottery/domain/cerror"
	"github.com/linchengzhi/lottery/domain/dto"
	"github.com/linchengzhi/lottery/usecase/lottery_uc"
	"github.com/linchengzhi/lottery/util"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"time"
)

type LotteryHdr struct {
	lotteryUc lottery_uc.LotteryUc
	log       *zap.Logger
}

func NewLotteryHandler(uc lottery_uc.LotteryUc, log *zap.Logger) *LotteryHdr {
	return &LotteryHdr{
		uc,
		log,
	}
}

func (hdr *LotteryHdr) DrawLottery(c *gin.Context) {
	req := new(dto.DrawReq)
	if err := c.ShouldBindJSON(req); err != nil {
		hdr.log.Error("参数错误", zap.Any("req", req), zap.Error(err))
		util.RespondErr(c, cerror.ErrParam)
		return
	}
	if err := hdr.validateDrawRequest(req); err != nil {
		hdr.log.Error("抽奖失败", zap.Any("req", req), zap.Error(err))
		util.RespondErr(c, err)
		return
	}
	req.RequestId = c.GetHeader("request_id")
	req.RequestTime = time.Now()
	hdr.log.Info("抽奖", zap.Any("req", req))

	//设置30s超时
	tracingCtx, exists := c.Get("tracingContext")
	if !exists {
		tracingCtx = c.Request.Context()
	}

	ctx, cancel := middleware.WithTimeoutAndSpan(tracingCtx.(context.Context), 30*time.Second)
	defer cancel()
	resp, err := hdr.lotteryUc.Draw(ctx, req)
	if err != nil {
		hdr.log.Error("抽奖失败", zap.Any("req", req), zap.Any("error", err))
		util.RespondErr(c, err)
		return
	}
	hdr.log.Debug("抽奖成功", zap.Any("req", req))
	util.Respond(c, resp)
}

// List Article
func (hdr *LotteryHdr) ListPrize(c *gin.Context) {
	req := new(dto.ListPrizeReq)
	if err := c.ShouldBindJSON(req); err != nil {
		hdr.log.Error("参数错误", zap.Any("req", req), zap.Error(err))
		util.RespondErr(c, cerror.ErrParam)
		return
	}
	hdr.log.Info("获取奖品列表", zap.Any("req", req))
	resp, err := hdr.lotteryUc.ListPrizes(c, req)
	if err != nil {
		hdr.log.Error("抽奖失败", zap.Any("req", req), zap.Any("error", err))
		util.RespondErr(c, err)
		return
	}
	hdr.log.Debug("获取奖品列表成功", zap.Int("req", len(resp)))
	util.Respond(c, resp)
}

// 参数校验函数
func (hdr *LotteryHdr) validateDrawRequest(req *dto.DrawReq) error {
	if req.UserId == 0 {
		return errors.New("用户ID不能为空")
	}
	if req.ActivityId == 0 {
		return errors.New("活动ID不能为空")
	}
	// 其他校验逻辑
	return nil
}

func (hdr *LotteryHdr) validateListRequest(req *dto.ListPrizeReq) error {
	if req.UserId == 0 {
		return errors.New("用户ID不能为空")
	}
	if req.ActivityId == 0 {
		return errors.New("活动ID不能为空")
	}
	if req.Page == 0 {
		req.Page = 1
	}
	if req.PageSize == 0 {
		req.PageSize = 10
	}
	return nil
}

//func setReqInfo(c *gin.Context) context.Context {
//	reqId := c.GetHeader("request_id")
//	reqTime := c.GetHeader("request_time")
//	if reqId == "" {
//		reqId = util.UUID()
//	}
//	var d = time.Now()
//	if reqTime != "" {
//		parsedTime, err := time.Parse("2006-01-02 15:04:05", reqTime)
//		if err != nil {
//			return nil
//		}
//		d = parsedTime
//	}
//	return util.WithRequestInfo(c, reqId, d)
//}
