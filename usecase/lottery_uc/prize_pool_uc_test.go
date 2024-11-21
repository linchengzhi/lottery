package lottery_uc

import (
	"context"
	"github.com/linchengzhi/lottery/Infra/logger"
	"github.com/linchengzhi/lottery/domain/dto"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestPrizePoolUc_RandomAward(t *testing.T) {
	starLevels := []*dto.StarLevel{
		{
			Level:  1,
			Weight: 50,
			Prizes: []*dto.Prize{
				{Id: 1, Num: 1, Weight: 30},
				{Id: 2, Num: 1, Weight: 20},
			},
		},
		{
			Level:  2,
			Weight: 30,
			Prizes: []*dto.Prize{
				{Id: 3, Num: 1, Weight: 15},
				{Id: 4, Num: 1, Weight: 15},
			},
		},
	}

	lotterConf := new(dto.LotteryConf)
	lotterConf.ActivityId = 12345
	lotterConf.Price = 100
	lotterConf.StarLevels = starLevels

	l, _ := logger.New(nil)

	uc, err := NewPrizePoolUc(l, *lotterConf)
	assert.NoError(t, err)

	ctx := context.Background()
	drawNum := int64(100)
	awards, err := uc.RandomPrizes(ctx, 0, drawNum)
	assert.NoError(t, err)
	assert.Len(t, awards.Prizes, int(drawNum))
}

//
//func BenchmarkRandomAward(b *testing.B) {
//	starLevels := createStarLevels()
//	uc, err := NewPrizePoolUc(starLevels)
//	if err != nil {
//		b.Fatalf("Failed to create PrizePoolUc: %v", err)
//	}
//
//	ctx := context.Background()
//	drawNum := int64(100) // 每次抽取10个奖品
//
//	b.ResetTimer()
//
//	b.RunParallel(func(pb *testing.PB) {
//		for pb.Next() {
//			_, err = uc.RandomPrizes(ctx, drawNum)
//			if err != nil {
//				b.Fatalf("RandomPrizes failed: %v", err)
//			}
//		}
//	})
//}

func createStarLevels() []*dto.StarLevel {
	starLevels := make([]*dto.StarLevel, 3)

	// 一星级：60个物品
	starLevels[0] = &dto.StarLevel{
		Level:  1,
		Weight: 60,
		Prizes: createPrizes(1, 60),
	}

	// 二星级：30个物品
	starLevels[1] = &dto.StarLevel{
		Level:  2,
		Weight: 30,
		Prizes: createPrizes(61, 90),
	}

	// 三星级：10个物品
	starLevels[2] = &dto.StarLevel{
		Level:  3,
		Weight: 10,
		Prizes: createPrizes(91, 100),
	}

	return starLevels
}

func createPrizes(start, end int) []*dto.Prize {
	prizes := make([]*dto.Prize, end-start+1)
	for i := start; i <= end; i++ {
		prizes[i-start] = &dto.Prize{
			Id:     int64(i),
			Weight: int64(i), // 权重从1到100
		}
	}
	return prizes
}

//func TestRandomAwardStatistics(t *testing.T) {
//	starLevels := createStarLevels()
//	uc, err := NewPrizePoolUc(starLevels)
//	if err != nil {
//		t.Fatalf("Failed to create PrizePoolUc: %v", err)
//	}
//
//	ctx := context.Background()
//	totalDraws := 1000000
//	drawNum := int64(1) // 每次抽取1个奖品
//
//	results := make(map[int64]int)
//
//	for i := 0; i < totalDraws; i++ {
//		awards, err := uc.RandomPrizes(ctx, drawNum)
//		if err != nil {
//			t.Fatalf("RandomPrizes failed: %v", err)
//		}
//		results[awards[0]]++
//	}
//
//	fmt.Printf("Total draws: %d\n", totalDraws)
//	fmt.Println("Prize ID | Count | Probability")
//	fmt.Println("---------|-------|------------")
//	for i := int64(1); i <= 100; i++ {
//		count := results[i]
//		probability := float64(count) / float64(totalDraws) * 100
//		fmt.Printf("%8d | %5d | %10.6f%%\n", i, count, probability)
//	}
//
//	// 验证总数
//	totalCount := 0
//	for _, count := range results {
//		totalCount += count
//	}
//	if totalCount != totalDraws {
//		t.Errorf("Total count (%d) does not match total draws (%d)", totalCount, totalDraws)
//	}
//}
