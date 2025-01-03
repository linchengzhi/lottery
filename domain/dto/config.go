package dto

import (
	"github.com/linchengzhi/lottery/Infra/logger"
)

type Config struct {
	AppName    string         `yaml:"app_name"`
	Env        string         `yaml:"env"`
	DebugPort  string         `yaml:"debug_port"`
	HTTP       HTTP           `yaml:"http"`
	Log        *logger.Config `yaml:"log"`
	Mysql      Mysql          `yaml:"mysql"`
	Redis      Redis          `yaml:"redis"`
	Stream     []RedisStream  `yaml:"redis_stream"`
	Lottery    LotteryConf    `yaml:"lottery"`
	JaegerConf JaegerConf     `json:"jaeger" yaml:"jaeger"`
}

type HTTP struct {
	Port string `yaml:"port"`
}

type Mysql struct {
	Host         string `json:"host" yaml:"host"`                     // 服务器地址
	Port         string `json:"port" yaml:"port"`                     // 端口
	Dbname       string `json:"dbname" yaml:"dbname"`                 // 数据库名
	Username     string `json:"username" yaml:"username"`             // 数据库用户名
	Password     string `json:"password" yaml:"password"`             // 数据库密码
	Config       string `json:"config" yaml:"config"`                 // 高级配置
	MaxIdleConns int    `json:"max_idle_conns" yaml:"max_idle_conns"` // 空闲中的最大连接数
	MaxOpenConns int    `json:"max_open_conns" yaml:"max_open_conns"` // 打开到数据库的最大连接数
	MaxLifeTime  int    `json:"max_life_time" yaml:"max_life_time"`
	LogLevel     string `json:"logLevel" yaml:"log_level"`
}

type Redis struct {
	Addr     string `yaml:"addr"`
	Password string `yaml:"password"`
	Db       int    `yaml:"db"`
}

type RedisStream struct {
	Name     string `yaml:"name"`
	Group    string `yaml:"group"`
	Consumer string `yaml:"consumer"`
}

type LotteryConf struct {
	ActivityId int64        `json:"activity_id" yaml:"activity_id"`
	Price      int64        `json:"price" yaml:"price"`
	StarLevels []*StarLevel `json:"star_levels" yaml:"star_levels"`
}

type JaegerConf struct {
	Host         string  `json:"host" yaml:"host"`
	Port         string  `json:"port" yaml:"port"`
	SamplingRate float64 `json:"sampling_rate" yaml:"sampling_rate"`
}
