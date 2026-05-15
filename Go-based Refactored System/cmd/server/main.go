package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/talent-assessment/refactored/internal/config"
	"github.com/talent-assessment/refactored/internal/router"
	"github.com/talent-assessment/refactored/pkg/db"
	"github.com/talent-assessment/refactored/pkg/redisx"
)

func main() {
	cfg := config.Load()
	database := db.Init(cfg.Mysql.DSN, cfg.Mysql.MaxOpen, cfg.Mysql.MaxIdle)
	redisx.Init(cfg.Redis.Addr, cfg.Redis.DB, cfg.Redis.Password)

	r, shutdown := router.Setup(cfg, database)
	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	srv := &http.Server{
		Addr:         addr,
		Handler:      r,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 120 * time.Second, // 给 chromedp 报告生成留足时间（page timeout 默认 60s）
	}

	go func() {
		log.Printf("[server] listening on %s", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("[server] listen: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop
	log.Println("[server] shutting down")
	shutdown()
}
