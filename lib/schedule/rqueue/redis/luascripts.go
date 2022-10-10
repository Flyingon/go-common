// rqueue.redis lua scripts

package redis

import goredis "github.com/go-redis/redis/v8"

const (
	luaZPopByScoreToNew = `
local srcKey = KEYS[1]
local dstKey = KEYS[2]
local scoreIncrKey = KEYS[3]
local minScore = ARGV[1]
local maxScore = ARGV[2]
local popSize = ARGV[3]
local scoreNow = tonumber(ARGV[4])

local dataList = redis.call('ZRANGEBYSCORE', srcKey, minScore, maxScore, 'WITHSCORES', "LIMIT", 0, popSize)

if dataList == nil or #dataList == 0
then
 	return nil
end

for i=1,#dataList,2 do
	local key = dataList[i]
	--解析score
    local taskType = ""
  	for token in string.gmatch(key, "[^|]+") do
    	taskType = token
    	break
  	end
	--查询score增量
  	local incrScore = redis.call('HGET', scoreIncrKey, taskType)
	--计算新score，默认增加60
  	local score = scoreNow + 60
  	if incrScore ~= nil and incrScore ~= false and incrScore ~= 0
  	then
    	score = scoreNow + incrScore
  	end
  	local num = redis.call('ZREM', srcKey, key)
  	redis.call('ZADD', dstKey, score, key)
end
return dataList
`

	luaZPopMaxToNew = `
local srcKey = KEYS[1]
local dstKey = KEYS[2]
local scoreIncrKey = KEYS[3]
local popSize = tonumber(ARGV[1])
local scoreNow = tonumber(ARGV[2])
if popSize < 1
then
  	return -1
end
local dataList = redis.call('ZREVRANGE', srcKey, 0, popSize-1,'WITHSCORES')

if dataList == nil or #dataList == 0
then
 	return nil
end

for i=1,#dataList,2 do
  	local key = dataList[i]
	--解析score
  	local taskType = ""
  	for token in string.gmatch(key, "[^|]+") do
    	taskType = token
    	break
	end
	--查询score增量
  	local incrScore = redis.call('HGET', scoreIncrKey, taskType)
	--计算新score
  	local score = scoreNow + 60
  	if incrScore ~= nil and incrScore ~= false
  	then
    	score = scoreNow + incrScore
  	end
  	redis.call('ZREM', srcKey, key)
  	redis.call('ZADD', dstKey, score, key)
end
return dataList
`
)

var zPopByScoreToNew = goredis.NewScript(luaZPopByScoreToNew)
var zPopMaxToNew = goredis.NewScript(luaZPopMaxToNew)
