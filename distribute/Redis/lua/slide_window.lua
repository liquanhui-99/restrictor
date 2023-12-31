---
--- Generated by EmmyLua(https://github.com/EmmyLua)
--- Created by liquanhui.
--- DateTime: 2023/9/5 13:06
---
--- 缓存中的key
local key = KEYS[1]
--- 窗口的大小
local window = tonumber(ARGV[1])
--- 阈值
local threshold = tonumber(ARGV[2])
--- 当前请求的时间戳
local now = tonumber(ARGV[3])
--- 窗口的最小时间戳
local min = now - window

--- 清空不在窗口内的所有请求数据
redis.call("ZREMRANGEBYSCORE", key, "-inf", min)
--- 获取集合中的请求数量
local cnt = redis.call("ZCOUNT", key, "-inf", "+inf")

if cnt >= threshold then
    --- 达到了滑动窗口内的最大请求数量，执行限流
    return "true"
else
    redis.call("ZADD", key, now, now)
    redis.call("PEXPIRE", key, window)
    return "false"
end