package config

import (
	cryptoRand "crypto/rand"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/spf13/viper"
)

type ServerCfg struct {
	Port int `mapstructure:"port"`
}

type MysqlCfg struct {
	DSN     string `mapstructure:"dsn"`
	MaxOpen int    `mapstructure:"maxOpen"`
	MaxIdle int    `mapstructure:"maxIdle"`
}

type RedisCfg struct {
	Addr     string `mapstructure:"addr"`
	DB       int    `mapstructure:"db"`
	Password string `mapstructure:"password"`
}

type JwtCfg struct {
	Secret        string `mapstructure:"secret"`
	ExpireMinutes int    `mapstructure:"expireMinutes"`
	Header        string `mapstructure:"header"`
	Prefix        string `mapstructure:"prefix"`
	LoginUserKey  string `mapstructure:"loginUserKey"`
}

type CaptchaCfg struct {
	Enabled bool   `mapstructure:"enabled"`
	Type    string `mapstructure:"type"`
}

type UploadCfg struct {
	Path                string `mapstructure:"path"`
	Profile             string `mapstructure:"profile"`
	MbtiTemplates       string `mapstructure:"mbtiTemplates"`
	MbtiTemplatesSimple string `mapstructure:"mbtiTemplatesSimple"`
	ExportTemplates     string `mapstructure:"exportTemplates"`
	// LegacyPdfRoot 旧 Java 系统的 PDF 实际根目录（绝对路径）
	// 用于将 DB 中残留的旧路径（如 c:/wwwroot/home/pdf/...）映射到 Linux 实际位置
	// 例如：客户服务器 /root/deploy6/c:/wwwroot/home  → 配置为 /root/deploy6
	LegacyPdfRoot string `mapstructure:"legacyPdfRoot"`
}

type PdfGenCfg struct {
	Enabled       bool   `mapstructure:"enabled"`
	ChromePath    string `mapstructure:"chromePath"`
	PoolSize      int    `mapstructure:"poolSize"`
	PageTimeoutMs int    `mapstructure:"pageTimeoutMs"`
	InternalToken string `mapstructure:"internalToken"` // 为空时启动随机生成
	ReportBaseURL string `mapstructure:"reportBaseURL"` // 如 http://localhost
	MaxBatch      int    `mapstructure:"maxBatch"`
}

type Config struct {
	Server  ServerCfg  `mapstructure:"server"`
	Mysql   MysqlCfg   `mapstructure:"mysql"`
	Redis   RedisCfg   `mapstructure:"redis"`
	Jwt     JwtCfg     `mapstructure:"jwt"`
	Captcha CaptchaCfg `mapstructure:"captcha"`
	Upload  UploadCfg  `mapstructure:"upload"`
	PdfGen  PdfGenCfg  `mapstructure:"pdfgen"`
}

var Global *Config

func Load() *Config {
	v := viper.New()
	v.SetConfigType("yaml")

	// 基础文件
	v.SetConfigName("application")
	v.AddConfigPath("./configs")
	v.AddConfigPath("../configs")
	v.AddConfigPath(".")
	if err := v.ReadInConfig(); err != nil {
		log.Printf("[config] base not found: %v (continue with env only)", err)
	}

	// 环境覆盖
	env := strings.ToLower(os.Getenv("APP_ENV"))
	if env == "" {
		env = "local"
	}
	v.SetConfigName("application-" + env)
	if err := v.MergeInConfig(); err != nil {
		log.Printf("[config] env overlay 'application-%s' not loaded: %v", env, err)
	}

	// 环境变量覆盖（SERVER_PORT, MYSQL_DSN, REDIS_ADDR, REDIS_DB, REDIS_PASSWORD, JWT_SECRET, JWT_EXPIRE_MINUTES, CAPTCHA_ENABLED, UPLOAD_PATH）
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	bindEnv(v, "server.port", "SERVER_PORT")
	bindEnv(v, "mysql.dsn", "MYSQL_DSN")
	bindEnv(v, "redis.addr", "REDIS_ADDR")
	bindEnv(v, "redis.db", "REDIS_DB")
	bindEnv(v, "redis.password", "REDIS_PASSWORD")
	bindEnv(v, "jwt.secret", "JWT_SECRET")
	bindEnv(v, "jwt.expireMinutes", "JWT_EXPIRE_MINUTES")
	bindEnv(v, "captcha.enabled", "CAPTCHA_ENABLED")
	bindEnv(v, "upload.path", "UPLOAD_PATH")
	bindEnv(v, "pdfgen.enabled", "PDFGEN_ENABLED")
	bindEnv(v, "pdfgen.chromePath", "PDFGEN_CHROME_PATH")
	bindEnv(v, "pdfgen.poolSize", "PDFGEN_POOL_SIZE")
	bindEnv(v, "pdfgen.internalToken", "PDFGEN_INTERNAL_TOKEN")
	bindEnv(v, "pdfgen.reportBaseURL", "PDFGEN_REPORT_BASE_URL")

	var c Config
	if err := v.Unmarshal(&c); err != nil {
		log.Fatalf("[config] unmarshal failed: %v", err)
	}
	// 填默认
	if c.Server.Port == 0 {
		c.Server.Port = 8092
	}
	if c.Jwt.Header == "" {
		c.Jwt.Header = "Authorization"
	}
	if c.Jwt.Prefix == "" {
		c.Jwt.Prefix = "Bearer "
	}
	if c.Jwt.LoginUserKey == "" {
		c.Jwt.LoginUserKey = "login_user_key"
	}
	if c.Mysql.MaxOpen == 0 {
		c.Mysql.MaxOpen = 50
	}
	if c.Mysql.MaxIdle == 0 {
		c.Mysql.MaxIdle = 10
	}
	// pdfgen 默认值
	if c.PdfGen.PoolSize == 0 {
		c.PdfGen.PoolSize = 2
	}
	if c.PdfGen.PageTimeoutMs == 0 {
		c.PdfGen.PageTimeoutMs = 45000
	}
	if c.PdfGen.ChromePath == "" {
		c.PdfGen.ChromePath = "/usr/bin/google-chrome"
	}
	if c.PdfGen.ReportBaseURL == "" {
		c.PdfGen.ReportBaseURL = "http://127.0.0.1"
	}
	if c.PdfGen.MaxBatch == 0 {
		c.PdfGen.MaxBatch = 100
	}
	if c.PdfGen.InternalToken == "" {
		// 启动时随机生成一个 32 字节 hex
		b := make([]byte, 16)
		_, _ = cryptoRand.Read(b)
		c.PdfGen.InternalToken = fmt.Sprintf("%x", b)
		log.Printf("[config] pdfgen.internalToken auto-generated")
	}
	Global = &c
	return &c
}

func bindEnv(v *viper.Viper, key, env string) {
	if val, ok := os.LookupEnv(env); ok && val != "" {
		v.Set(key, val)
	}
}
