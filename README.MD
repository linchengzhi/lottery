# 抽奖服务
## 流程
1、用户发送请求，在中间件进行验证，限流，过滤重复请求，必要时熔断和降级
2、进入抽奖请求处理，先校验参数，然后设置到redis stream中，等待处理