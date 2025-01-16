package gredis

import (
	"context"
	"encoding/json"

	"github.com/pkg/errors"
	"github.com/redis/go-redis/v9"
)

// hsetCASLuaScript 对 hash 某个 field 的 CAS
// KEYS[1] hash key
// ARGV[1] hash field name
// ARGV[2] 预期值
// ARGV[3] 目标值，即需要替换成的值
// 返回 success 代表是否设置成功，latestVal 代表设置后（无论成功与否）的最新值
const hsetCASLuaScript = `
	local val = redis.call('hget', KEYS[1], ARGV[1])
	if val == false then
		return val
	end
	local resp = {success=false}
	if val == ARGV[2] then
		val = ARGV[3]
		resp['success'] = true
		redis.call('hset', KEYS[1], ARGV[1], ARGV[3])
	end
	resp['latestVal'] = val
	return cjson.encode(resp)
`

type HashCASResult struct {
	Success   bool    `json:"success,omitempty"`
	LatestVal *string `json:"latestVal,omitempty"`
}

func HSetCAS(ctx context.Context, client redis.Cmdable, hashKey, fieldName, expectVal, updateVal string) (*HashCASResult, error) {
	result, err := client.Eval(ctx, hsetCASLuaScript, []string{hashKey}, fieldName, expectVal, updateVal).Text()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return &HashCASResult{}, nil
		}
		return nil, errors.WithStack(err)
	}
	var casResult HashCASResult
	err = json.Unmarshal([]byte(result), &casResult)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return &casResult, nil
}

const hsetXXLuaScript = `
	local val = redis.call('hget', KEYS[1], ARGV[1])
	if val then
		redis.call('hset', KEYS[1], ARGV[1], ARGV[2])
		return 1
	end
	return 0

`

func HSetXX(ctx context.Context, client redis.Cmdable, hashKey, fieldName, fieldVal string) (bool, error) {
	success, err := client.Eval(ctx, hsetXXLuaScript, []string{hashKey}, fieldName, fieldVal).Bool()
	if err != nil {
		return false, errors.WithStack(err)
	}
	return success, nil
}
