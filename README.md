# restrictor
restrictor仓库封装的是请求流量的限流器，分为单体服务和分布式服务两种。

Limiter接口是单体服务，单体服务提供了四种实现：
1. 令牌痛限流
2. 漏桶限流
3. 固定窗口限流
4. 滑动窗口限流

DistributedLimiter接口是分布式服务的限流器，提供了两种实现：
1. 固定窗口限流
2. 滑动窗口限流

IpLimiter在单体Limiter基础上封装的ip限流器
