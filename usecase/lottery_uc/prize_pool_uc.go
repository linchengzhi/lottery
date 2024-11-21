package lottery_uc

import (
	"context"
	"github.com/linchengzhi/lottery/domain/cerror"
	"github.com/linchengzhi/lottery/domain/dto"
	"go.uber.org/zap"
	"math/rand"
	"time"
)

type IPrizePoolUc interface {
	// 随机获取奖池中drawNum个奖品
	RandomPrizes(ctx context.Context, userId, drawNum int64) (*dto.PrizeData, error)
	//获取单抽加个
	getPrice(ctx context.Context) int64
}

type PrizePoolUc struct {
	activityId int64
	price      int64
	pool       *dto.PrizePool // 奖池
	log        *zap.Logger
}

// NewPrizePoolUc 创建一个新的 PrizePoolUc 实例
func NewPrizePoolUc(log *zap.Logger, conf dto.LotteryConf) (IPrizePoolUc, error) {
	p := new(PrizePoolUc)
	p.activityId = conf.ActivityId
	p.price = conf.Price
	p.log = log
	err := p.createPool(conf.StarLevels)
	if err != nil {
		return nil, err
	}
	return p, nil
}

func (p *PrizePoolUc) getPrice(ctx context.Context) int64 {
	return p.price
}

// createPool 创建奖池
func (p *PrizePoolUc) createPool(starLevels []*dto.StarLevel) error {
	if len(starLevels) == 0 {
		return cerror.ErrLotteryConfig
	}
	p.pool = &dto.PrizePool{}
	levelsCumWeight := int64(0)
	for _, level := range starLevels {
		levelsCumWeight += level.Weight
		level.Weight = levelsCumWeight

		prizesCumWeight := int64(0)
		for _, prize := range level.Prizes {
			prizesCumWeight += prize.Weight
			prize.Weight = prizesCumWeight
		}
	}
	p.pool.Prizes = starLevels
	return nil
}

// RandomPrizes 随机获取奖池中 drawNum 个奖品
func (p *PrizePoolUc) RandomPrizes(ctx context.Context, userId, drawNum int64) (*dto.PrizeData, error) {
	items := make([]*dto.Item, drawNum)
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := int64(0); i < drawNum; i++ {
		// Step 1: 随机选择一个星级
		starLevel := randomStarLevel(r, p.pool.Prizes)

		// Step 2: 在选中的星级中随机选择一个奖品
		prize := randomPrizeFromLevel(r, starLevel)

		item := new(dto.Item)
		item.Id = prize.Id
		item.Num = prize.Num
		items[i] = item
	}
	data := &dto.PrizeData{
		ActivityId: p.activityId,
		UserId:     userId,
		Prizes:     items,
	}
	return data, nil
}

// randomStarLevel 根据权重随机选择一个星级
func randomStarLevel(r *rand.Rand, levels []*dto.StarLevel) *dto.StarLevel {
	randVal := r.Int63n(levels[len(levels)-1].Weight)
	for _, level := range levels {
		if randVal < level.Weight {
			return level
		}
	}
	return nil
}

// randomPrizeFromLevel 根据权重随机选择一个星级中的奖品
func randomPrizeFromLevel(r *rand.Rand, level *dto.StarLevel) *dto.Prize {
	randVal := r.Int63n(level.Prizes[len(level.Prizes)-1].Weight)
	for _, prize := range level.Prizes {
		if randVal < prize.Weight {
			return prize
		}
	}
	return nil
}
