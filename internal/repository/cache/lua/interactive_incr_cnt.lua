local key = KEYS[1]
local cntKey = ARGV[1]
local delta = tonumber(ARGV[2])
local exists = redis.call("EXISTS", key)

if exists == 1 then
    -- Hincrby 命令用于为哈希表(key)中的字段值(cntKey)加上指定增量值(dalta)
    -- HINCRBY KEY_NAME FIELD_NAME INCR_BY_NUMBER
    redis.call("HINCRBY", key, cntKey, delta)
    -- 自增成功
    return 1
else
    -- 自增失败
    return 0
end