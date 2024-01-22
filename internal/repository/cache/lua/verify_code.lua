local key = KEYS[1]
local expectedCode = ARGV[1]
local code = redis.call("get", key)
local cntKey = key..":cnt"
local cnt = tonumber(redis.call("get", cntKey))
if cnt <= 0 then
    -- 用户一直输错验证码或者用完
    return -1
elseif expectedCode == code then
    -- 输对验证码
    -- 把次数标记为 -1，表示验证码不可用
    redis.call("set", cntKey, -1)
    return 0
else
    -- 用户输入错误
    redis.call("decr", cntKey)
    return -2
end