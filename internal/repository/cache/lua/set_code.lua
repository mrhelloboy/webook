--- 验证码在 Redis 上的 key
--- eg: phone_code:login:15200000000
local key = KEYS[1]
--- 验证次数，一个验证码最多重复使用 3 次，记录了还可以验证几次
--- eg: phone_code:login:15200000000:cnt
local cntKey = key..":cnt"
--- 验证码 123456
local val = ARGV[1]
--- 过期时间
local ttl = tonumber(redis.call("ttl", key))
if ttl == -1 then
    --- key 存在，但没有过期时间
    return -2
elseif ttl == -2 or ttl < 540 then
    --- key 不存在，或者过期时间小于 9 分钟( 540: 600 - 60 )
    redis.call("set", key, val)
    redis.call("expire", key, 600)
    redis.call("set", cntKey, 3)
    redis.call("expire", cntKey, 600)
    return 0
else
    --- 发送太频繁
    return -1
end
