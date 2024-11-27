package lottery_uc

import (
	"context"
	"github.com/bytedance/sonic"
	"github.com/go-redis/redis/v8"
	"github.com/linchengzhi/lottery/Infra/database/redis_db"
	"github.com/linchengzhi/lottery/Infra/gpool"
	"github.com/linchengzhi/lottery/domain/cerror"
	"github.com/linchengzhi/lottery/domain/dto"
	"github.com/linchengzhi/lottery/domain/entity"
	"github.com/linchengzhi/lottery/repository/mysql_repo"
	"github.com/linchengzhi/lottery/repository/redis_repo"
	"github.com/linchengzhi/lottery/usecase/asset_uc"
	"github.com/linchengzhi/lottery/util"
	"github.com/opentracing/opentracing-go"
	"go.uber.org/zap"
	"strings"
	"sync"
	"time"
)

type ILotteryUc interface {
	// 设置活动奖池
	SetPrizePool(ctx context.Context, conf dto.LotteryConf) error
	// 抽奖
	Draw(ctx context.Context, req *dto.DrawReq) (*dto.DrawResp, error)
	// 奖品列表
	ListPrizes(ctx context.Context, req *dto.ListPrizeReq) ([]*entity.LotteryPrizeRecord, error)
}

// 私有接口，仅在包内使用
type lotteryInternal interface {
	// 获取活动奖池
	getPrizePool(ctx context.Context, activityId int64) (IPrizePoolUc, error)
	// 抽奖处理逻辑
	lotteryHandle(ctx context.Context, req *dto.DrawReq) (*dto.PrizeData, error)
	// 奖品发放
	award(ctx context.Context, prize *dto.AwardStream) error
	// 批量插入抽奖记录
	processAwardData(ctx context.Context)
}

type LotteryUc struct {
	log  *zap.Logger
	pool *gpool.Pool

	prizeMu   *sync.RWMutex
	prizePool map[int64]IPrizePoolUc //活动id->奖池

	reqCh    chan *DrawData  //抽奖请求 先入channel等待处理
	recordCh chan *AwardData //抽奖记录，用于批量插入

	drawRepo     mysql_repo.LotteryDrawRecordRepo
	prizeRepo    mysql_repo.LotteryPrizeRecordRepo
	lotteryCache redis_repo.LotteryRecordCache
	awardRs      redis_db.IStream

	assetUc asset_uc.AssetUc
}

type DrawData struct {
	req    *dto.DrawReq
	result chan *dto.DrawResp
	ctx    context.Context // 添加 context 字段
}

type AwardData struct {
	drawRecords  []*entity.LotteryDrawRecord
	prizeRecords []*entity.LotteryPrizeRecord
	ch           chan error
}

func NewLotteryUc(log *zap.Logger, g *gpool.Pool, repoMysql mysql_repo.RepoMysql, repoRedis redis_repo.RepoRedis, repoStream redis_repo.RepoStream, assetUc asset_uc.AssetUc) LotteryUc {
	uc := LotteryUc{
		log:  log,
		pool: g,

		prizeMu:   &sync.RWMutex{},
		prizePool: make(map[int64]IPrizePoolUc),

		reqCh:    make(chan *DrawData, 1000),
		recordCh: make(chan *AwardData, 10000),

		drawRepo:     repoMysql.LotteryDrawRecordRepo,
		prizeRepo:    repoMysql.LotteryPrizeRecordRepo,
		lotteryCache: repoRedis.LotteryRecordCache,
		awardRs:      repoStream.AwardRs,

		assetUc: assetUc,
	}
	go uc.lotteryCache.GetTimeout(context.Background(), uc.RollbackCallBack)
	go uc.processDrawData(context.Background())
	go uc.awardRs.Get(uc.AwardCallBack)
	go uc.processAwardData(context.Background())
	return uc
}

func (uc *LotteryUc) AwardCallBack(message redis.XMessage) error {
	var awardReq *dto.AwardStream
	defer util.CheckGoPanicWithParam(uc.log, awardReq)
	// 1. 解析 Redis 消息中的数据

	err := sonic.Unmarshal([]byte(message.Values["data"].(string)), &awardReq)
	if err != nil {
		uc.log.Error("抽奖 发奖参数错误 数据异常", zap.Any("message", message), zap.Any("err", err))
		return nil
	}

	err = uc.award(context.Background(), awardReq)
	if err != nil {
		return err
	}
	return nil
}

// 读取req channel中的请求进行处理
func (uc *LotteryUc) processDrawData(ctx context.Context) {
	for {
		select {
		case data := <-uc.reqCh:
			uc.pool.Submit(func() {
				resp, err := uc.lotteryHandle(data.ctx, data.req)
				data.result <- &dto.DrawResp{
					RequestId: data.req.RequestId,
					PrizeData: resp,
					Err:       err,
				}
			})
		}
	}
}

func (uc *LotteryUc) SetPrizePool(ctx context.Context, conf dto.LotteryConf) error {
	puc, err := NewPrizePoolUc(uc.log, conf)
	if err != nil {
		return err
	}
	uc.prizeMu.Lock()
	uc.prizePool[conf.ActivityId] = puc
	uc.prizeMu.Unlock()
	err = uc.drawRepo.CreateTable(ctx, conf.ActivityId)
	if err != nil {
		return err
	}
	err = uc.prizeRepo.CreateTable(ctx, conf.ActivityId)
	if err != nil {
		return err
	}
	uc.log.Info("设置奖池成功", zap.Int64("activityId", conf.ActivityId))
	return nil
}

func (uc *LotteryUc) getPrizePool(ctx context.Context, activityId int64) (IPrizePoolUc, error) {
	uc.prizeMu.RLock()
	defer uc.prizeMu.RUnlock()
	if p, ok := uc.prizePool[activityId]; ok {
		return p, nil
	}
	return nil, cerror.ErrLotteryNoAct
}

var drawDataPool = sync.Pool{
	New: func() interface{} {
		return &DrawData{
			result: make(chan *dto.DrawResp, 1),
		}
	},
}

// Draw 抽奖
func (uc *LotteryUc) Draw(ctx context.Context, req *dto.DrawReq) (*dto.DrawResp, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "Draw")
	if span != nil {
		defer span.Finish()
		span.SetTag("request_id", req.RequestId)
	}
	//设置参数到redis stream中
	data := drawDataPool.Get().(*DrawData)
	defer drawDataPool.Put(data)
	data.req = req
	data.ctx = ctx

	uc.reqCh <- data
	for {
		select {
		case resp := <-data.result:
			return resp, resp.Err
		case <-ctx.Done():
			span.SetTag("error", true)
			span.SetTag("error.message", "timeout")
			return nil, cerror.ErrTimeout
		}
	}
}

func (uc *LotteryUc) lotteryHandle(ctx context.Context, req *dto.DrawReq) (*dto.PrizeData, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "lotteryHandle")
	if span != nil {
		defer span.Finish()
		span.SetTag("request_id", req.RequestId)
	}

	// 检查请求时间
	if time.Now().Unix()-req.RequestTime.Unix() > 28 {
		uc.log.Warn("抽奖请求超时", zap.Any("req", req), zap.Any("now", time.Now()))
		return nil, cerror.ErrTimeout
	}

	// 1. 设置抽奖请求缓存
	cacheSpan, cacheCtx := opentracing.StartSpanFromContext(ctx, "set_lottery_cache")
	defer cacheSpan.Finish()
	err := uc.lotteryCache.Set(cacheCtx, req)
	if err != nil {
		uc.log.Warn("抽奖失败 设置缓存记录失败", zap.Any("req", req), zap.Error(err))
		return nil, cerror.ErrBusy
	}

	// 读取奖池
	poolSpan, poolCtx := opentracing.StartSpanFromContext(ctx, "get_prize_pool")
	defer poolSpan.Finish()
	puc, err := uc.getPrizePool(poolCtx, req.ActivityId)
	if err != nil {
		return nil, err
	}

	// 2. 随机抽取奖品
	randomSpan, randomCtx := opentracing.StartSpanFromContext(ctx, "random_prizes")
	defer randomSpan.Finish()
	prizesData, err := puc.RandomPrizes(randomCtx, req.UserId, req.DrawNum)
	if err != nil {
		uc.log.Warn("抽奖失败 随机奖品失败", zap.Any("req", req), zap.Error(err))
		return nil, err
	}

	// 3. 扣除用户资产
	assetSpan, assetCtx := opentracing.StartSpanFromContext(ctx, "update_asset")
	defer assetSpan.Finish()
	at := new(entity.UserAsset)
	at.UserID = req.UserId
	at.Stone = -puc.getPrice(ctx) * req.DrawNum
	err = uc.assetUc.UpdateAsset(assetCtx, at, req.RequestId, req.RequestTime)
	if err != nil {
		uc.log.Warn("抽奖失败 更新资产失败", zap.Any("req", req), zap.Error(err))
		return nil, err
	}

	// 4. 保存奖品列表到 redis stream
	streamSpan, _ := opentracing.StartSpanFromContext(ctx, "save_to_stream")
	defer streamSpan.Finish()
	awardStream := new(dto.AwardStream)
	awardStream.RequestId = req.RequestId
	awardStream.RequestTime = req.RequestTime
	awardStream.PrizeData = prizesData

	byteAs, _ := sonic.Marshal(awardStream)
	_, err = uc.awardRs.Add(string(byteAs))
	if err != nil {
		return nil, err
	}

	// 5. 删除缓存
	delSpan, delCtx := opentracing.StartSpanFromContext(ctx, "delete_cache")
	defer delSpan.Finish()
	uc.lotteryCache.Del(delCtx, req.RequestId)

	// 设置总体执行结果标签
	span.SetTag("success", true)
	span.SetTag("prizes_count", len(prizesData.Prizes))

	return prizesData, nil
}

var awardDataPool = sync.Pool{
	New: func() interface{} {
		return &AwardData{
			ch: make(chan error, 1),
		}
	},
}

func (uc *LotteryUc) award(ctx context.Context, aStream *dto.AwardStream) error {
	var err error
	var items = make(map[int64]int64)
	for _, v := range aStream.PrizeData.Prizes {
		items[v.Id] += v.Num
	}
	currentTime := time.Now()
	// 1. 更新用户物品数据
	err = uc.assetUc.UpdateItems(ctx, aStream.PrizeData.UserId, items, aStream.RequestId, aStream.RequestTime)
	if err != nil {
		if strings.Contains(err.Error(), "Duplicate entry") {
			// 重复发奖，跳过发奖，继续
		} else {
			return cerror.ErrBusy
		}

	}
	// 2. 插入抽奖记录
	var prizeRecords = make([]*entity.LotteryPrizeRecord, 0)
	for _, v := range aStream.PrizeData.Prizes {
		prizeRecords = append(prizeRecords, &entity.LotteryPrizeRecord{
			ActivityID: aStream.PrizeData.ActivityId,
			UserID:     aStream.PrizeData.UserId,
			PrizeID:    v.Id,
			PrizeNum:   v.Num,
			CreatedAt:  currentTime,
		})
	}
	record := new(entity.LotteryDrawRecord)
	record.ActivityID = aStream.PrizeData.ActivityId
	record.UserID = aStream.PrizeData.UserId
	record.DrawCount = len(aStream.PrizeData.Prizes)
	record.RequestID = aStream.RequestId
	record.CreatedAt = currentTime

	ad := awardDataPool.Get().(*AwardData)
	defer awardDataPool.Put(ad)
	ad.drawRecords = []*entity.LotteryDrawRecord{record}
	ad.prizeRecords = prizeRecords
	uc.recordCh <- ad
	return <-ad.ch
}

// 定时任务处理函数
func (uc *LotteryUc) processAwardData(ctx context.Context) {
	ticker := time.NewTicker(time.Duration(1) * time.Second)
	defer ticker.Stop()

	var buffer []*AwardData // 用于批量处理的缓冲区

	var writeRecord = func() {
		// 合并所有 AwardData 的 drawRecords 和 prizeRecords
		var allDrawRecords []*entity.LotteryDrawRecord
		var allPrizeRecords []*entity.LotteryPrizeRecord

		for _, awardData := range buffer {
			allDrawRecords = append(allDrawRecords, awardData.drawRecords...)
			allPrizeRecords = append(allPrizeRecords, awardData.prizeRecords...)
		}

		// 模拟插入数据库
		err := uc.drawRepo.BatchCreate(ctx, allDrawRecords, allPrizeRecords)
		for _, awardData := range buffer {
			awardData.ch <- err // 通知插入数据库的结果
		}
		// 清空缓冲区
		buffer = []*AwardData{}
	}

	for {
		select {
		case awardData := <-uc.recordCh:
			// 收到一个 AwardData，放入缓冲区
			buffer = append(buffer, awardData)
			if len(buffer) >= 200 {
				// 缓冲区满了，批量处理
				writeRecord()
			}
		case <-ticker.C:
			// 定时批量处理
			if len(buffer) == 0 {
				continue
			}
			writeRecord()
		}
	}
}

func (uc *LotteryUc) ListPrizes(ctx context.Context, req *dto.ListPrizeReq) ([]*entity.LotteryPrizeRecord, error) {
	list, err := uc.prizeRepo.ListByUserId(ctx, req.ActivityId, req.UserId, req.Page, req.PageSize)
	if err != nil {
		return nil, err
	}
	return list, nil
}

func (uc *LotteryUc) RollbackCallBack(req *dto.DrawReq) error {
	var err error
	defer func() {
		if err == nil {
			uc.lotteryCache.Del(context.Background(), req.RequestId)
		}
	}()

	if req == nil {
		return nil
	}
	if req.PrizesData == nil { //无需回滚
		return nil
	}

	if req.RequestTime.Before(time.Now().Add(-10 * time.Minute)) {
		return nil // 超过10分钟，无需回滚
	}
	//读取资产记录
	var record *entity.UserAssetRecord
	record, err = uc.assetUc.GetAssetRecord(context.Background(), req.RequestId)
	if err != nil {
		uc.log.Warn("抽奖失败 回滚资产失败", zap.Any("req", req), zap.Error(err))
		return err
	}
	if record == nil {
		return nil //无需回滚
	}
	at := new(entity.UserAsset)
	at.UserID = req.UserId
	at.Stone = req.PrizesData.Amount
	req.RequestId = req.RequestId + "0"
	// 2. 更新用户资产数据
	err = uc.assetUc.UpdateAsset(context.Background(), at, req.RequestId+"0", req.RequestTime) // 假设每次抽奖扣除用户资产
	if err != nil {
		if strings.Contains(err.Error(), "Duplicate entry ") {
			return nil
		}
		// 失败则等待下次重试
		uc.log.Warn("抽奖失败 回滚资产失败", zap.Any("req", req), zap.Error(err))
		return err
	}
	// todo 发个邮件，通知用户抽奖失败，已返回资产
	return nil
}
