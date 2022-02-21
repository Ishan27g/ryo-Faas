package rateLimit

import (
	"context"
	"crypto/md5"
	"fmt"
	"strconv"
	"time"

	goredislib "github.com/go-redis/redis/v8"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	redis "github.com/go-redis/redis/v8"
)

const timeout = 5 * time.Second
const Interval = 60 * time.Second
const RequestLimit = 10 // 10 request every minute

var redisHost = func() string {
	// os.getenv
	return "redis:6379"
}
var cache *redisClient

type Cache interface {
	Allow(user string) (bool, int)
}

func init() {
	options := redis.Options{
		Addr:        redisHost(), // "localhost:6379"
		Password:    "",
		DB:          0,
		MaxConnAge:  1 * time.Minute,
		IdleTimeout: 30 * time.Second,
	}
	cache = new(redisClient)
	cache.client = goredislib.NewClient(&options)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	rsp := cache.client.Ping(ctx)
	if rsp.Err() != nil {
		fmt.Println(rsp.Err())
	} else {
		fmt.Println(rsp.String())
	}
}

type redisClient struct {
	client *redis.Client
}

func (r *redisClient) Allow(ctx context.Context, user string) (bool, int, string, trace.Span) {
	span := trace.SpanFromContext(ctx)

	now := time.Now()
	redisKey := fmt.Sprintf("%s%x", "rate-session-id-", md5.Sum([]byte(user)))

	if exists, _ := r.check(redisKey); !exists {
		r.add(redisKey, "0", Interval)
		span.AddEvent("Redis-add-id" + time.Since(now).String())
		span.SetAttributes(attribute.String("redis-add-id", time.Since(now).String()))
		now = time.Now()
	}
	requestCount := r.incAndGet(redisKey)
	span.SetAttributes(attribute.String("redis-inc-id", time.Since(now).String()))
	return requestCount < RequestLimit, requestCount, redisKey, span
}

func (r *redisClient) check(key string) (bool, string) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	res := r.client.Get(ctx, key).Val()
	if res != "" {
		return true, res
	}
	return false, ""
}
func (r *redisClient) add(key, value string, expireIn time.Duration) {
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()
		query := r.client.Set(ctx, key, value, expireIn)
		if query.Err() != nil {
			fmt.Println(query.Err().Error())
			return
		}
	}()
}
func (r *redisClient) incAndGet(key string) int {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	if err := r.client.Incr(ctx, key).Err(); err == nil {
		result, err := r.client.Get(ctx, key).Result()
		if err != nil {
			fmt.Println(err.Error())
			return 0
		}
		count, err := strconv.Atoi(result)
		if err != nil {
			fmt.Println(err.Error())
			return 0
		}
		return count
	} else {
		fmt.Println(err.Error())
		return 0
	}
}
