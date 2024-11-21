package initializer

import (
	"context"
	"github.com/go-redis/redis/v8"
	"github.com/linchengzhi/lottery/Infra/config"
	"github.com/linchengzhi/lottery/Infra/database/mysql_db"
	redis2 "github.com/linchengzhi/lottery/Infra/database/redis_db"
	"github.com/linchengzhi/lottery/Infra/gpool"
	"github.com/linchengzhi/lottery/Infra/logger"
	"github.com/linchengzhi/lottery/domain/dto"
	"github.com/linchengzhi/lottery/repository/mysql_repo"
	"github.com/linchengzhi/lottery/repository/redis_repo"
	"github.com/linchengzhi/lottery/usecase"
	"go.uber.org/zap"
	"runtime"
)

type App struct {
	Conf        *dto.Config
	Log         *zap.Logger
	MysqlLog    *zap.Logger
	GPool       *gpool.Pool
	MysqlDb     *mysql_db.Gorm
	RedisDb     *redis.Client
	RedisStream redis_repo.RepoStream
	RepoMysql   mysql_repo.RepoMysql
	RepoRedis   redis_repo.RepoRedis
	UcAll       usecase.UcAll
}

func NewApp(configPath string) (*App, error) {
	app := &App{}
	if err := app.initialize(configPath); err != nil {
		return nil, err
	}
	return app, nil
}

func (app *App) initialize(configPath string) error {
	if err := app.initConfigAndLogger(configPath); err != nil {
		return err
	}
	app.initGoPool()

	if err := app.initMysql(); err != nil {
		return err
	}

	if err := app.initRedis(); err != nil {
		return err
	}

	app.initRepositories()
	app.initUsecases()

	app.InitActivity()
	return nil
}

func (app *App) initConfigAndLogger(configPath string) error {
	conf, err := config.NewConfig(configPath)
	if err != nil {
		return err
	}
	app.Conf = conf

	log, err := logger.New(app.Conf.Log)
	if err != nil {
		return err
	}
	app.Log = log

	logConf := app.Conf.Log
	logConf.Level = "info"
	logConf.Filename = "logs/mysql_repo.log"
	mysqlLog, err := logger.New(logConf)
	if err != nil {
		return err
	}
	app.MysqlLog = mysqlLog

	return nil
}

func (app *App) initMysql() error {
	db, err := mysql_db.NewGorm(app.Conf.Mysql, app.MysqlLog)
	if err != nil {
		return err
	}
	app.MysqlDb = db
	return nil
}

func (app *App) initRedis() error {
	db, err := redis2.NewRedis(app.Conf.Redis.Addr, app.Conf.Redis.Password, app.Conf.Redis.Db)
	if err != nil {
		return err
	}
	app.RedisDb = db

	for i := 0; i < len(app.Conf.Stream); i++ {
		app.Conf.Stream[i].Consumer = "test"
	}
	stream, err := redis_repo.NewRepoStream(app.RedisDb, app.GPool, app.Conf.Stream)
	if err != nil {
		return err
	}
	app.RedisStream = stream
	return nil
}

func (app *App) initGoPool() {
	pool, err := gpool.NewPool(app.Log, runtime.NumCPU()*30)
	if err != nil {
		app.Log.Fatal("Failed to initialize go pool", zap.Error(err))
	}
	app.GPool = pool
}

func (app *App) initRepositories() {
	app.RepoMysql = mysql_repo.NewRepoMysql(app.MysqlDb.DB)
	app.RepoRedis = redis_repo.NewRepoRedis(app.RedisDb)
}

func (app *App) initUsecases() {
	app.UcAll = usecase.NewUcAll(app.Log, app.GPool, app.RepoMysql, app.RepoRedis, app.RedisStream)
}

// 初始化活动
func (app *App) InitActivity() {
	err := app.UcAll.LotteryUc.SetPrizePool(context.Background(), app.Conf.Lottery)
	if err != nil {
		app.Log.Error("初始化活动失败", zap.Any("err", err))
	}
}
