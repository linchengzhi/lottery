package shutdown

import (
	"os"
	"os/signal"
	"sync"
	"syscall"
)

var (
	hooks   []func()
	hooksMu sync.Mutex
)

// Register 注册一个在程序关闭时需要执行的函数
func Register(fn func()) {
	hooksMu.Lock()
	defer hooksMu.Unlock()
	hooks = append(hooks, fn)
}

// Listen 监听系统信号并执行注册的关闭函数
func Listen() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		// 执行所有注册的关闭函数
		hooksMu.Lock()
		defer hooksMu.Unlock()

		for i := len(hooks) - 1; i >= 0; i-- {
			hooks[i]()
		}

		os.Exit(0)
	}()
}
