package dcls

import (
	"context"
	"fmt"
	"strconv"

	redis "github.com/redis/go-redis/v9"
)

type RedisClient struct {
	client *redis.Client
	script string
}

func NewBucketClient(server *redis.Client) *RedisClient {
	return &RedisClient{
		client: server,
		script: `
        -- KEY[1]:  桶名
        -- ARGV[1]: 桶容器容量
        -- ARGV[2]: 令牌速率
        
        local bucket = KEYS[1]
        local capacity = tonumber(ARGV[1])
        local rate = tonumber(ARGV[2])
        local now = redis.call("TIME")
        
        local seconds = tonumber(now[1])
        local table = redis.call('HMGET',bucket,'tokens','lastfill') 
        local tokens = tonumber(table[1])
        local lastfill = tonumber(table[2])

        -- 如果nil则初始化桶
        if tokens == nil or lastfill == nil then
            tokens = capacity
            lastfill = seconds
        else
            local durationTk = (seconds - lastfill) *rate

            if  tokens + durationTk > capacity then
                tokens = capacity
            else
                tokens = durationTk + tokens
            end
        end

        if tokens < 1 then
            return 0
        else
            redis.call('HMSET',bucket,'tokens',tokens - 1,'lastfill',seconds)
            return 1
        end
		`,
	}
}

func (c RedisClient) Check(cxt context.Context, bucket string, capacity int64, rate int64) (bool, error) {
	var scr = redis.NewScript(c.script)

	bol, _ := scr.Run(cxt, c.client, []string{bucket}, []string{strconv.FormatInt(capacity, 10), strconv.FormatInt(rate, 10)}).Int64()
	fmt.Println("state :", bol)
	if bol == 1 {
		return true, nil
	} else {
		return false, nil
	}
}
