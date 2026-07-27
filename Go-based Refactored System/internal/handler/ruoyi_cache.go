package handler

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/talent-assessment/refactored/pkg/redisx"
	"github.com/talent-assessment/refactored/pkg/response"
)

const cacheScanBatchSize int64 = 100

func configCacheKey(configKey string) string {
	return redisx.SysConfigKey + configKey
}

func dictCacheKey(dictType string) string {
	return redisx.SysDictKey + dictType
}

func deleteCacheKeys(ctx context.Context, client *redis.Client, keys ...string) error {
	if client == nil {
		return errors.New("Redis不可用")
	}
	unique := make([]string, 0, len(keys))
	seen := make(map[string]struct{}, len(keys))
	for _, key := range keys {
		key = strings.TrimSpace(key)
		if key == "" {
			continue
		}
		if _, exists := seen[key]; exists {
			continue
		}
		seen[key] = struct{}{}
		unique = append(unique, key)
	}
	if len(unique) == 0 {
		return nil
	}
	return client.Del(ctx, unique...).Err()
}

func scanDeletePrefix(ctx context.Context, client *redis.Client, prefix string, batchSize int64) error {
	if client == nil {
		return errors.New("Redis不可用")
	}
	if prefix == "" {
		return errors.New("缓存前缀不能为空")
	}
	if batchSize <= 0 {
		batchSize = cacheScanBatchSize
	}
	for {
		var cursor uint64
		deletedAny := false
		for {
			keys, next, err := client.Scan(ctx, cursor, prefix+"*", batchSize).Result()
			if err != nil {
				return err
			}
			if len(keys) > 0 {
				if err := client.Del(ctx, keys...).Err(); err != nil {
					return err
				}
				deletedAny = true
			}
			cursor = next
			if cursor == 0 {
				break
			}
		}
		if !deletedAny {
			return nil
		}
	}
}

func invalidateConfigKeys(c *gin.Context, keys ...string) bool {
	cacheKeys := make([]string, 0, len(keys))
	for _, key := range keys {
		cacheKeys = append(cacheKeys, configCacheKey(key))
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
	defer cancel()
	if err := deleteCacheKeys(ctx, redisx.Client, cacheKeys...); err != nil {
		response.AjaxErr(c, "配置已保存，但缓存失效失败，请刷新配置缓存")
		return false
	}
	return true
}

func invalidateDictTypes(c *gin.Context, dictTypes ...string) bool {
	cacheKeys := make([]string, 0, len(dictTypes))
	for _, dictType := range dictTypes {
		cacheKeys = append(cacheKeys, dictCacheKey(dictType))
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
	defer cancel()
	if err := deleteCacheKeys(ctx, redisx.Client, cacheKeys...); err != nil {
		response.AjaxErr(c, "字典已保存，但缓存失效失败，请刷新字典缓存")
		return false
	}
	return true
}

func (h *RuoYiSystemHandler) ConfigRefreshCache(c *gin.Context) {
	if _, ok := requireSystemPermission(c, "system:config:edit"); !ok {
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()
	if err := scanDeletePrefix(ctx, redisx.Client, redisx.SysConfigKey, cacheScanBatchSize); err != nil {
		response.AjaxErr(c, "配置缓存刷新失败")
		return
	}
	response.AjaxOK(c, nil)
}

func (h *RuoYiSystemHandler) DictRefreshCache(c *gin.Context) {
	if _, ok := requireSystemPermission(c, "system:dict:edit"); !ok {
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()
	if err := scanDeletePrefix(ctx, redisx.Client, redisx.SysDictKey, cacheScanBatchSize); err != nil {
		response.AjaxErr(c, "字典缓存刷新失败")
		return
	}
	response.AjaxOK(c, nil)
}
