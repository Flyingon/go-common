// 自定义方法，lua脚本
package luafunction

// luaScriptKeySPopAndRecord 从zset中pop key并存储到另一个zset
// 成功返回key(string), 失败返回nil
var luaScriptKeySPopAndRecord = `
redis.replicate_commands()
local popQueue = KEYS[1]
local recordQueue = KEYS[2]

local elems = redis.call("SPOP", popQueue, 1)
if #elems == 1 then
  redis.call('SADD', recordQueue, elems[1])
  return elems[1]
end
`

// luaScriptKeySPopAndRecord 从zset中pop key并存按时间记录到zset中
// 成功返回key(string), 失败返回nil
var luaScriptKeySPopToZSet = `
redis.replicate_commands()
local popQueue = KEYS[1]
local recordQueue = KEYS[2]
local ts=redis.call('TIME')[1]
 
local elems = redis.call("SPOP", popQueue, 1)
if #elems == 1 then
  redis.call('ZADD', recordQueue, ts, elems[1])
  return elems[1]
end
`

// LuaScriptZPopByScore 从zset中pop key，根据score限制
var LuaScriptZPopByScore = `
local setname = KEYS[1]
local minscore = ARGV[1]
local maxscore = ARGV[2]
local order = ARGV[3]

local redisTable = nil
if order == "desc" then
    redisTable = redis.call('ZREVRANGEBYSCORE', setname, maxscore, minscore,'WITHSCORES')
else
    redisTable = redis.call('ZRANGEBYSCORE', setname, minscore, maxscore,'WITHSCORES')
end

if redisTable == nil
then
  return redisTable
end

for i=1,#redisTable,2 do
  local key = redisTable[i]
  local num = redis.call('ZREM', setname, key)
end
return redisTable
`
