package types

import "go.uber.org/zap"

const (
	SessionMaxAge = 7 * 24 * 3600
)

type MysqlLogger *zap.Logger
