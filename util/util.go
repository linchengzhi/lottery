package util

import (
	"fmt"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"os"
	"strings"
	"time"
)

func UUID() string {
	id := uuid.New()
	return strings.Replace(id.String(), "-", "", -1)
}

type IELog interface {
	Error(template string, fields ...zap.Field)
}

func CheckGoPanicWithParam(log IELog, param ...interface{}) {
	if err := recover(); err != nil {
		// 获取堆栈信息
		//buf := make([]byte, 4096)
		//n := runtime.Stack(buf, false)
		//stackInfo := string(buf[:n])

		// 格式化参数信息
		var paramInfo string
		if len(param) > 0 {
			paramValues := make([]string, len(param))
			for i, p := range param {
				paramValues[i] = fmt.Sprintf("%+v", p)
			}
			paramInfo = strings.Join(paramValues, ", ")
		} else {
			paramInfo = "no parameters"
		}

		// 构建错误信息  日志自带堆栈
		errMsg := fmt.Sprintf(`
=== PANIC RECOVERED ===
Time: %s
Error: %v
Parameters: %s
Stack Trace:
=====================`,
			time.Now().Format("2006-01-02 15:04:05.000"),
			err,
			paramInfo,
			//stackInfo,
		)

		// 使用日志接口记录错误
		if log != nil {
			log.Error(errMsg)
		} else {
			// 如果没有提供日志接口，则打印到标准错误输出
			fmt.Fprintln(os.Stderr, errMsg)
		}
	}
}
