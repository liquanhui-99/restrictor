---
--- Generated by EmmyLua(https://github.com/EmmyLua)
--- Created by liquanhui.
--- DateTime: 2023/9/5 12:31
---
--- 缓存中是否有这个key的限流：key可以是服务，也可以是接口
local val = redis.call("GET", KEYS[1])
--- 最大的限流数
local limit = tonumber(ARGV[1])
--- key的超时时间
local expiration = tonumber(ARGV[2])

if val == false then
    if limit < 1 then
        -- 执行限流
        return "true"
    else
        -- 通过限流器
        redis.call("SET", KEYS[1], 1, "EX", expiration)
        return "false"
    end
elseif tonumber(val) < limit then
    -- 存在限流对象，但是还未到阈值，可以通过限流器
    redis.call("INCR", KEYS[1])
    return "false"
else
    -- 执行限流
    return "true"
end