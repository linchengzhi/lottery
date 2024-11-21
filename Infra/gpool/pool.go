package gpool

import (
	"github.com/panjf2000/ants"
	"go.uber.org/zap"
	"sync"
)

// Pool 封装ants协程池
type Pool struct {
	pool *ants.PoolWithFunc // ants 协程池
	wg   sync.WaitGroup     // 用于等待所有任务完成
	log  *zap.Logger        // 日志
}

// NewPool 创建一个协程池，数量为核心数的 15 倍
func NewPool(log *zap.Logger, num int) (*Pool, error) {
	pool, err := ants.NewPoolWithFunc(num, func(task interface{}) {
		// 处理任务的逻辑
		if taskFunc, ok := task.(func()); ok {
			taskFunc()
		} else {
			log.Error("收到的任务不是一个有效的函数")
		}
	})
	if err != nil {
		return nil, err
	}

	return &Pool{pool: pool}, nil
}

// Submit 提交一个任务到协程池
func (p *Pool) Submit(task func()) {
	p.wg.Add(1) // 增加等待组的计数器
	err := p.pool.Invoke(task)
	if err != nil {
		p.log.Warn("提交任务失败", zap.Error(err))
		p.wg.Done() // 提交失败时需要手动减少计数
	}
}

// Wait 等待所有任务完成
func (p *Pool) Wait() {
	p.wg.Wait()
}

// Release 释放协程池
func (p *Pool) Release() {
	p.pool.Release()
}
