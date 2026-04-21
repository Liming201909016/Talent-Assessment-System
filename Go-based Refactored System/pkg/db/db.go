package db

import (
	"log"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

var DB *gorm.DB

func Init(dsn string, maxOpen, maxIdle int) *gorm.DB {
	g, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Warn),
	})
	if err != nil {
		log.Fatalf("[db] open failed: %v", err)
	}
	sqlDB, err := g.DB()
	if err != nil {
		log.Fatalf("[db] underlying sql.DB unavailable: %v", err)
	}
	sqlDB.SetMaxOpenConns(maxOpen)
	sqlDB.SetMaxIdleConns(maxIdle)
	sqlDB.SetConnMaxLifetime(time.Hour)
	if err := sqlDB.Ping(); err != nil {
		log.Fatalf("[db] ping failed: %v", err)
	}
	log.Printf("[db] connected")
	DB = g
	return g
}
