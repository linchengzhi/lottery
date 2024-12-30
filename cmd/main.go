package main

import (
	"context"
	"encoding/gob"
	"flag"
	"github.com/gin-gonic/gin"
	"github.com/linchengzhi/lottery/Infra/database/mysql_db"
	"github.com/linchengzhi/lottery/api/http/router"
	"github.com/linchengzhi/lottery/cmd/initializer"
	"github.com/linchengzhi/lottery/domain/entity"
	"github.com/linchengzhi/lottery/util"
	"go.uber.org/zap"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var configFilePath = flag.String("f", "config/config_dev.yaml", "the config file")

func init() {
	//debug.SetGCPercent(100)                 // 调整 GC 触发阈值
	//debug.SetMemoryLimit(100 * 1024 * 1024) // 设置内存限制
}

func main() {
	flag.Parse()

	app, err := initializer.NewApp(*configFilePath)
	if err != nil {
		panic(err)
	}

	// 初始化调试服务
	go util.InitDebug([]string{"0.0.0.0:" + app.Conf.DebugPort})

	// 数据库迁移
	mysql_db.AutoMigrate(app.MysqlDb.DB)

	// 初始化 gin
	gin.SetMode(gin.ReleaseMode) // 生产环境建议使用 ReleaseMode
	g := gin.New()
	g.Use(gin.Recovery())
	g.Use(gin.Logger())

	// 注册 gob
	gob.Register(entity.User{})

	// 设置路由
	router.SetRoutes(app.UcAll, app.Log, g, app.RedisDb)

	// 创建 HTTP 服务器
	srv := &http.Server{
		Addr:           ":" + app.Conf.HTTP.Port,
		Handler:        g,
		ReadTimeout:    30 * time.Second,
		WriteTimeout:   30 * time.Second,
		MaxHeaderBytes: 1 << 20,
		IdleTimeout:    120 * time.Second,
	}

	// 启动服务器
	go func() {
		app.Log.Info("server is running",
			zap.String("port", app.Conf.HTTP.Port),
			zap.Any("config", app.Conf))

		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			app.Log.Fatal("listen failed", zap.Error(err))
		}
	}()

	// 优雅关闭
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	app.Log.Info("shutting down server...")

	// 创建关闭超时上下文
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 关闭 HTTP 服务器
	if err := srv.Shutdown(ctx); err != nil {
		app.Log.Fatal("server forced to shutdown", zap.Error(err))
	}

	app.Log.Info("server exited properly")
}
