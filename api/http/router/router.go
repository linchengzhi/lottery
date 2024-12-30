package router

import (
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/linchengzhi/lottery/api/http/handler"
	"github.com/linchengzhi/lottery/api/http/middleware"
	"github.com/linchengzhi/lottery/usecase"
	"go.uber.org/zap"
)

func SetRoutes(uc usecase.UcAll, log *zap.Logger, gin *gin.Engine, rdb *redis.Client) {
	// All Public APIs
	publicRouter := gin.Group("")

	// All Private APIs
	//protectedRouter := gin.Group("")

	publicRouter.Use(

		middleware.RateLimitMiddleware(600, 12000),
		middleware.RequestIdMiddleware(rdb),

		//middleware.TracingMiddleware(), // 添加追踪中间件
		//middleware.RepeatedLimitMiddleware(rdb),
	)

	// Middleware to verify AccessToken
	//protectedRouter.Use(
	//	middleware.CheckLogin(),
	//	middleware.RateLimitMiddleware(100, 1000),
	//	middleware.RequestIdMiddleware(rdb),
	//)

	NewLotteryRouter(uc, log, publicRouter)
	NewAssetRouter(uc, log, publicRouter)
}

func NewLotteryRouter(uc usecase.UcAll, log *zap.Logger, public *gin.RouterGroup) {
	ud := handler.NewLotteryHandler(uc.LotteryUc, log)

	pu := public.Group("lottery")
	pu.POST("draw", Handle(ud.DrawLottery))
	pu.GET("prize/list", Handle(ud.ListPrize))
}

func NewAssetRouter(uc usecase.UcAll, log *zap.Logger, public *gin.RouterGroup) {
	ud := handler.NewAssetHandler(uc.AssetUc, log)

	pu := public.Group("asset")
	pu.GET("get", Handle(ud.GetAsset))
	pu.GET("item/list", Handle(ud.ListItem))
}
