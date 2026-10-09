package main

import (
	"context"
	"errors"
	"net"
	"testing"
	"time"

	"github.com/talent-assessment/refactored/internal/config"
	"github.com/talent-assessment/refactored/internal/service"
)

func TestResolveServerRuntimeControls(t *testing.T) {
	tests := []struct {
		name               string
		appEnv             string
		server             config.ServerCfg
		resolved           []net.IPAddr
		resolveErr         error
		wantAddress        string
		wantWorkersEnabled bool
		wantError          bool
		wantResolverCalls  int
	}{
		{name: "empty host preserves wildcard bind", appEnv: "production", server: config.ServerCfg{Port: 8092}, wantAddress: ":8092", wantWorkersEnabled: true},
		{name: "ipv4 wildcard preserves wildcard bind", appEnv: "staging", server: config.ServerCfg{Host: "0.0.0.0", Port: 8092}, wantAddress: ":8092", wantWorkersEnabled: true},
		{name: "loopback ipv4 without DNS", appEnv: "local", server: config.ServerCfg{Host: "127.0.0.1", Port: 18093}, wantAddress: "127.0.0.1:18093", wantWorkersEnabled: true},
		{name: "loopback ipv6 without DNS", appEnv: "local", server: config.ServerCfg{Host: "::1", Port: 18093}, wantAddress: "[::1]:18093", wantWorkersEnabled: true},
		{name: "localhost prefers canonical ipv4", appEnv: "local", server: config.ServerCfg{Host: "localhost", Port: 18093}, resolved: []net.IPAddr{{IP: net.ParseIP("::1")}, {IP: net.ParseIP("127.0.0.1")}}, wantAddress: "127.0.0.1:18093", wantWorkersEnabled: true, wantResolverCalls: 1},
		{name: "localhost canonical ipv6 fallback", appEnv: "local", server: config.ServerCfg{Host: "localhost", Port: 18093}, resolved: []net.IPAddr{{IP: net.ParseIP("::1")}}, wantAddress: "[::1]:18093", wantWorkersEnabled: true, wantResolverCalls: 1},
		{name: "localhost spoof rejected", appEnv: "local", server: config.ServerCfg{Host: "localhost", Port: 18093}, resolved: []net.IPAddr{{IP: net.ParseIP("192.0.2.10")}}, wantError: true, wantResolverCalls: 1},
		{name: "localhost mixed loopback and nonloopback rejected", appEnv: "local", server: config.ServerCfg{Host: "localhost", Port: 18093}, resolved: []net.IPAddr{{IP: net.ParseIP("127.0.0.1")}, {IP: net.ParseIP("192.0.2.10")}}, wantError: true, wantResolverCalls: 1},
		{name: "localhost empty resolution rejected", appEnv: "local", server: config.ServerCfg{Host: "localhost", Port: 18093}, wantError: true, wantResolverCalls: 1},
		{name: "localhost resolver error rejected", appEnv: "local", server: config.ServerCfg{Host: "localhost", Port: 18093}, resolveErr: errors.New("synthetic resolver failure"), wantError: true, wantResolverCalls: 1},
		{name: "uppercase localhost rejected without DNS", appEnv: "local", server: config.ServerCfg{Host: "LOCALHOST", Port: 18093}, wantError: true},
		{name: "localhost with whitespace rejected without DNS", appEnv: "local", server: config.ServerCfg{Host: " localhost ", Port: 18093}, wantError: true},
		{name: "invalid hostname rejected without DNS", appEnv: "local", server: config.ServerCfg{Host: "example.com", Port: 8092}, wantError: true},
		{name: "host with port rejected without DNS", appEnv: "local", server: config.ServerCfg{Host: "127.0.0.1:8092", Port: 8092}, wantError: true},
		{name: "local loopback may disable workers", appEnv: "local", server: config.ServerCfg{Host: "127.0.0.1", Port: 18093, DisableBackgroundWorkers: true}, wantAddress: "127.0.0.1:18093", wantWorkersEnabled: false},
		{name: "local resolved localhost may disable workers", appEnv: "local", server: config.ServerCfg{Host: "localhost", Port: 18093, DisableBackgroundWorkers: true}, resolved: []net.IPAddr{{IP: net.ParseIP("127.0.0.1")}}, wantAddress: "127.0.0.1:18093", wantWorkersEnabled: false, wantResolverCalls: 1},
		{name: "production rejects disabled workers", appEnv: "production", server: config.ServerCfg{Host: "127.0.0.1", Port: 8092, DisableBackgroundWorkers: true}, wantError: true},
		{name: "local wildcard rejects disabled workers", appEnv: "local", server: config.ServerCfg{Port: 18093, DisableBackgroundWorkers: true}, wantError: true},
		{name: "local nonloopback rejects disabled workers", appEnv: "local", server: config.ServerCfg{Host: "192.0.2.10", Port: 18093, DisableBackgroundWorkers: true}, wantError: true},
		{name: "app environment must be exact local", appEnv: "LOCAL", server: config.ServerCfg{Host: "127.0.0.1", Port: 18093, DisableBackgroundWorkers: true}, wantError: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resolverCalls := 0
			resolver := func(ctx context.Context, host string) ([]net.IPAddr, error) {
				resolverCalls++
				if host != "localhost" {
					t.Fatalf("resolver host=%q, want localhost", host)
				}
				deadline, ok := ctx.Deadline()
				if !ok || time.Until(deadline) <= 0 || time.Until(deadline) > 2*time.Second {
					t.Fatalf("resolver context has invalid deadline: %v, %t", deadline, ok)
				}
				return tt.resolved, tt.resolveErr
			}
			got, err := resolveServerRuntimeControls(tt.appEnv, tt.server, resolver)
			if (err != nil) != tt.wantError {
				t.Fatalf("resolveServerRuntimeControls() error=%v, wantError=%t", err, tt.wantError)
			}
			if resolverCalls != tt.wantResolverCalls {
				t.Fatalf("resolver calls=%d, want=%d", resolverCalls, tt.wantResolverCalls)
			}
			if err == nil && (got.ListenAddress != tt.wantAddress || got.BackgroundWorkersEnabled != tt.wantWorkersEnabled) {
				t.Fatalf("resolveServerRuntimeControls()=%+v, want address=%q workers=%t", got, tt.wantAddress, tt.wantWorkersEnabled)
			}
		})
	}
}

// TestBugFB216_ProductionServerSkipsManagementTraitsRuntimeFactoryAndWorker
// 对应：docs/regression-tests.md #FB-216
func TestBugFB216_ProductionServerSkipsManagementTraitsRuntimeFactoryAndWorker(t *testing.T) {
	factoryCalls := 0
	runtime := managementTraitsRuntimeForServer(false, func() *service.ManagementTraitsRuntimeService {
		factoryCalls++
		return service.NewManagementTraitsRuntimeService(nil, "", 1<<20)
	})
	workerCalls := 0
	started := startManagementTraitsExpiryWorker(context.Background(), false, func(context.Context) { workerCalls++ })
	if runtime != nil || factoryCalls != 0 || started || workerCalls != 0 {
		t.Fatalf("disabled runtime constructed or started: runtime=%v factory=%d started=%t worker=%d", runtime, factoryCalls, started, workerCalls)
	}
}

func TestBugFB216_LocalServerBuildsManagementTraitsRuntimeAndStartsWorker(t *testing.T) {
	factoryCalls := 0
	runtime := managementTraitsRuntimeForServer(true, func() *service.ManagementTraitsRuntimeService {
		factoryCalls++
		return service.NewManagementTraitsRuntimeService(nil, "test-only", 1<<20)
	})
	workerCalled := make(chan struct{}, 1)
	started := startManagementTraitsExpiryWorker(context.Background(), true, func(context.Context) { workerCalled <- struct{}{} })
	select {
	case <-workerCalled:
	case <-time.After(time.Second):
		t.Fatal("enabled expiry worker was not invoked")
	}
	if runtime == nil || factoryCalls != 1 || !started {
		t.Fatalf("enabled runtime assembly mismatch: runtime=%v factory=%d started=%t", runtime, factoryCalls, started)
	}
}
