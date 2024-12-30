package router

import (
	"github.com/gin-gonic/gin"
	"github.com/linchengzhi/lottery/domain/cerror"
	"net/http"
)

type HandlerFunc func(*gin.Context) (interface{}, error)

// Handle 是一个装饰器，它将 HandlerFunc 包装为 gin.HandlerFunc
func Handle(fn HandlerFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 调用实际的处理器函数
		data, err := fn(c)
		if err != nil {
			// 错误处理逻辑
			customErr, ok := err.(*cerror.CustomError)
			if ok {
				c.JSON(http.StatusBadRequest, gin.H{
					"code": customErr.GetCode(),
					"msg":  customErr.GetMsg(),
				})
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{
					"code": cerror.ErrSystem.GetCode(),
					"msg":  cerror.ErrSystem.GetMsg(),
				})
			}
			return
		}
		// 成功响应
		c.JSON(http.StatusOK, gin.H{
			"data": data,
		})
	}
}
