package main

import (
	"encoding/gob"
	"flag"
	"github.com/linchengzhi/lottery/Infra/database/mysql_db"
	"github.com/linchengzhi/lottery/cmd/initializer"
	"github.com/linchengzhi/lottery/util"

	"github.com/gin-gonic/gin"
	"github.com/linchengzhi/lottery/api/http/router"
	"github.com/linchengzhi/lottery/domain/entity"
	"go.uber.org/zap"
)

var configFilePath = flag.String("f", "config/config_dev.yaml", "the config file")

func main() {
	flag.Parse()

	app, err := initializer.NewApp(*configFilePath)
	if err != nil {
		panic(err)
	}
	go util.InitDebug([]string{"0.0.0.0:" + app.Conf.DebugPort})
	// AutoMigrate
	mysql_db.AutoMigrate(app.MysqlDb.DB)

	g := gin.Default()
	gob.Register(entity.User{})
	router.SetRoutes(app.UcAll, app.Log, g, app.RedisDb)
	app.Log.Debug("server is running", zap.Any("config", app.Conf))
	g.Run(":" + app.Conf.HTTP.Port)
}
