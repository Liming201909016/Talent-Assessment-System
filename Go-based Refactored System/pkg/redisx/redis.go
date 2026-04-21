package redisx

import (
	"context"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

var Client *redis.Client

func Init(addr string, db int, password string) *redis.Client {
	Client = redis.NewClient(&redis.Options{
		Addr:     addr,
		DB:       db,
		Password: password,
	})
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := Client.Ping(ctx).Err(); err != nil {
		log.Fatalf("[redis] ping failed: %v", err)
	}
	log.Printf("[redis] connected %s db=%d", addr, db)
	return Client
}

// Legacy Java Redis 键前缀
const (
	LoginTokenKey = "login_tokens:"
	CaptchaKey    = "captcha_codes:"
	SysConfigKey  = "sys_config:"
	SysDictKey    = "sys_dict:"
)
