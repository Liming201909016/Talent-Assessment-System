package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sort"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/talent-assessment/refactored/internal/config"
	"github.com/talent-assessment/refactored/internal/router"
	"github.com/talent-assessment/refactored/internal/service"
	"github.com/talent-assessment/refactored/pkg/db"
	"github.com/talent-assessment/refactored/pkg/redisx"
)

type serverRuntimeControls struct {
	ListenAddress            string
	BackgroundWorkersEnabled bool
}

type serverIPResolver func(context.Context, string) ([]net.IPAddr, error)

const serverHostResolveTimeout = 2 * time.Second

func resolveServerRuntimeControls(appEnv string, server config.ServerCfg, resolver serverIPResolver) (serverRuntimeControls, error) {
	host := strings.TrimSpace(server.Host)
	listenHost := host
	loopback := false

	switch {
	case host == "" || host == "0.0.0.0":
		listenHost = ""
	case net.ParseIP(host) != nil:
		ip := net.ParseIP(host)
		listenHost = ip.String()
		loopback = ip.IsLoopback()
	case server.Host == "localhost":
		if resolver == nil {
			return serverRuntimeControls{}, fmt.Errorf("invalid server host %q: resolver is unavailable", host)
		}
		ctx, cancel := context.WithTimeout(context.Background(), serverHostResolveTimeout)
		addresses, err := resolver(ctx, host)
		cancel()
		if err != nil {
			return serverRuntimeControls{}, fmt.Errorf("invalid server host %q: resolution failed", host)
		}
		canonical := make([]string, 0, len(addresses))
		seen := make(map[string]struct{}, len(addresses))
		for _, address := range addresses {
			if address.IP == nil || !address.IP.IsLoopback() {
				return serverRuntimeControls{}, fmt.Errorf("invalid server host %q: every resolved address must be loopback", host)
			}
			numeric := address.IP.String()
			if _, exists := seen[numeric]; !exists {
				seen[numeric] = struct{}{}
				canonical = append(canonical, numeric)
			}
		}
		if len(canonical) == 0 {
			return serverRuntimeControls{}, fmt.Errorf("invalid server host %q: resolution returned no addresses", host)
		}
		sort.Strings(canonical)
		listenHost = canonical[0]
		if _, exists := seen["127.0.0.1"]; exists {
			listenHost = "127.0.0.1"
		} else {
			ipv6 := make([]string, 0, len(addresses))
			for _, address := range addresses {
				if address.IP.To4() == nil {
					ipv6 = append(ipv6, address.IP.String())
				}
			}
			if len(ipv6) > 0 {
				sort.Strings(ipv6)
				listenHost = ipv6[0]
			}
		}
		loopback = true
	default:
		return serverRuntimeControls{}, fmt.Errorf("invalid server host %q: expected an IP address or exact localhost", host)
	}

	workersEnabled := !server.DisableBackgroundWorkers
	if !workersEnabled && (appEnv != "local" || !loopback) {
		return serverRuntimeControls{}, fmt.Errorf("LOCAL_DISABLE_BACKGROUND_WORKERS=true requires APP_ENV=local and a loopback server host")
	}

	port := strconv.Itoa(server.Port)
	address := ":" + port
	if listenHost != "" {
		address = net.JoinHostPort(listenHost, port)
	}
	return serverRuntimeControls{ListenAddress: address, BackgroundWorkersEnabled: workersEnabled}, nil
}

func managementTraitsRuntimeForServer(enabled bool, factory func() *service.ManagementTraitsRuntimeService) *service.ManagementTraitsRuntimeService {
	if !enabled {
		return nil
	}
	return factory()
}

func startManagementTraitsExpiryWorker(ctx context.Context, enabled bool, worker func(context.Context)) bool {
	if !enabled {
		return false
	}
	go worker(ctx)
	return true
}

func main() {
	cfg := config.Load()
	controls, err := resolveServerRuntimeControls(os.Getenv("APP_ENV"), cfg.Server, net.DefaultResolver.LookupIPAddr)
	if err != nil {
		log.Fatalf("[server] unsafe runtime configuration: %v", err)
	}
	database := db.Init(cfg.Mysql.DSN, cfg.Mysql.MaxOpen, cfg.Mysql.MaxIdle)
	redisx.Init(cfg.Redis.Addr, cfg.Redis.DB, cfg.Redis.Password)

	managementTraitsEnabled := config.ManagementTraitsTestRuntimeEnabled()
	managementTraitsRuntime := managementTraitsRuntimeForServer(managementTraitsEnabled, func() *service.ManagementTraitsRuntimeService {
		return service.NewManagementTraitsRuntimeService(database, cfg.Jwt.Secret, 1<<20)
	})
	var runtimeServices []*service.ManagementTraitsRuntimeService
	if managementTraitsRuntime != nil {
		runtimeServices = append(runtimeServices, managementTraitsRuntime)
	}
	r, shutdown := router.SetupWithOptions(cfg, database, router.SetupOptions{
		BackgroundWorkersEnabled: controls.BackgroundWorkersEnabled,
	}, runtimeServices...)
	managementTraitsContext, stopManagementTraits := context.WithCancel(context.Background())
	startManagementTraitsExpiryWorker(managementTraitsContext, managementTraitsEnabled && controls.BackgroundWorkersEnabled, func(ctx context.Context) {
		managementTraitsRuntime.RunExpiry(ctx, 30*time.Second, 100, func(err error) {
			log.Println("[management-traits] expiry scan rejected; will retry")
		})
	})
	srv := &http.Server{
		Addr:         controls.ListenAddress,
		Handler:      r,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 120 * time.Second, // 给 chromedp 报告生成留足时间（page timeout 默认 60s）
	}

	go func() {
		log.Printf("[server] listening on %s", controls.ListenAddress)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("[server] listen: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop
	log.Println("[server] shutting down")
	stopManagementTraits()
	shutdown()
}
