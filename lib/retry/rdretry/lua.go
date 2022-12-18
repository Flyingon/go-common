// lua scripts

package rdretry

const (
	luaZPopByScore = `
local queueKey = KEYS[1]
local minScore = ARGV[1]
local maxScore = ARGV[2]
local popSize = ARGV[3]

local dataList = redis.call('ZRANGEBYSCORE', queueKey, minScore, maxScore, 'WITHSCORES', "LIMIT", 0, popSize)

if dataList == nil or #dataList == 0
then
 	return nil
end

for i=1,#dataList,2 do
	local key = dataList[i]
  	local num = redis.call('ZREM', queueKey, key)
end
return dataList
`
)
