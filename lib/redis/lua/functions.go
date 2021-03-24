// 自定义方法，lua脚本
package luafunction

import redigo "github.com/gomodule/redigo/redis"

var SPopAndRecord *redigo.Script
var SPopToZSet *redigo.Script
var ZPopByScore *redigo.Script

func init() {
	SPopAndRecord = redigo.NewScript(2, luaScriptKeySPopAndRecord)
	SPopToZSet = redigo.NewScript(2, luaScriptKeySPopToZSet)
	ZPopByScore = redigo.NewScript(1, LuaScriptZPopByScore)
}
