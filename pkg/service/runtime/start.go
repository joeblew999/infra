package runtime

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/joeblew999/infra/pkg/config"
	"github.com/joeblew999/infra/pkg/goreman"
	"github.com/joeblew999/infra/pkg/log"
	runtimeevents "github.com/joeblew999/infra/pkg/runtime/events"
	svcports "github.com/joeblew999/infra/pkg/service/ports"
)

const (
	serviceStatePending   = "pending"
	serviceStateReady     = "ready"
	serviceStateReclaimed = "reclaimed"
	serviceStateRunning   = "running"
	serviceStateBlocked   = "blocked"
	serviceStateError     = "error"
	serviceStateStopped   = "stopped"
)

const (
	actionEnsureFailed      = "ensure_failed"
	actionPortInspectFailed = "port_inspection_failed"
	actionPortReclaimed     = "port_reclaimed"
	actionStartupBlocked    = "startup_blocked"
	actionStartFailed       = "start_failed"
	actionStarted           = "started"
	actionShutdown          = "shutdown"
)

// Start launches all infrastructure services under goreman supervision.
// It blocks until a shutdown signal is received or a startup error occurs.
func Start(opts Options) error {
	activeOptions = opts

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := config.EnsureAppDirectories(); err != nil {
		return fmt.Errorf("failed to prepare runtime directories: %w", err)
	}

	if opts.Preflight != nil {
		opts.Preflight(ctx)
	}

	log.Info("Running in Service mode with goreman supervision...")

	specs := buildServiceSpecs(opts)
	activeServiceSpecs = specs

	errCh := make(chan error, 1)

	var cleanupStack []func()

	defer func() {
		for i := len(cleanupStack) - 1; i >= 0; i-- {
			if cleanupStack[i] != nil {
				cleanupStack[i]()
			}
		}
		goreman.StopAll()
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(sigCh)

	go func() {
		select {
		case <-ctx.Done():
			return
		case <-sigCh:
			log.Info("🛑 Received shutdown signal, stopping all supervised processes...")
			cancel()
		}
	}()

	recordErr := func(err error) {
		if err == nil {
			return
		}
		select {
		case errCh <- err:
		default:
		}
		cancel()
	}

	log.Info("🚀 Starting all infrastructure services...")

	for idx, svc := range specs {
		if svc.Enabled != nil && !svc.Enabled(opts) {
			continue
		}

		step := idx + 1
		log.Info("🚀 Starting service", "step", step, "service", svc.DisplayName)
		startNotes := make([]string, 0, 2)
		port := svcports.ParsePort(svc.Port)

		if svc.Ensure != nil {
			if err := svc.Ensure(ctx, opts); err != nil {
				msg := fmt.Sprintf("Ensure failed: %v", err)
				log.Error("Failed to prepare service", "service", svc.DisplayName, "error", err)
				recordErr(fmt.Errorf("%s ensure failed: %w", svc.DisplayName, err))
				publishAction(svc, actionEnsureFailed, msg)
				publishStatus(svc, serviceStateError, false, 0, port, "unknown", msg)
				continue
			}
		}

		if svc.Port != "" && svc.Port != "0" {
			probe, err := inspectPort(svc.ID, port)
			if err != nil {
				msg := fmt.Sprintf("Port inspection failed: %v", err)
				log.Warn("Failed to inspect port", "service", svc.DisplayName, "port", svc.Port, "error", err)
				publishAction(svc, actionPortInspectFailed, msg)
				publishStatus(svc, serviceStateError, false, 0, port, "unknown", msg)
			} else {
				switch ownershipForService(svc, probe) {
				case portOwnershipFree:
					publishStatus(svc, serviceStatePending, false, 0, port, ownershipString(portOwnershipFree), "")
				case portOwnershipThisService:
					if shouldAutoReclaim(svc.ID) && reclaimManagedProcess(svc.ID, probe) {
						if !svcports.WaitAvailable(port, 3*time.Second) {
							msg := formatConflictMessage(svc.DisplayName, probe)
							log.Error("❌ Port still busy after reclaim attempt", "service", svc.DisplayName, "port", svc.Port, "detail", msg)
							publishAction(svc, actionStartupBlocked, fmt.Sprintf("Auto-reclaim failed: %s", msg))
							publishStatus(svc, serviceStateBlocked, false, 0, port, ownershipString(portOwnershipThisService), msg)
							return fmt.Errorf("service %s port %s is already in use", svc.DisplayName, svc.Port)
						}
						note := fmt.Sprintf("Reclaimed stale process (PID %s)", probe.PID)
						log.Info("Reclaimed port for service startup", "service", svc.DisplayName, "port", svc.Port, "note", note)
						startNotes = append(startNotes, note)
						publishAction(svc, actionPortReclaimed, note)
						publishStatus(svc, serviceStateReclaimed, false, 0, port, ownershipString(portOwnershipFree), "")
					} else {
						msg := formatConflictMessage(svc.DisplayName, probe)
						log.Warn("Port in use by existing service", "service", svc.DisplayName, "port", svc.Port, "detail", msg)
						publishAction(svc, actionStartupBlocked, fmt.Sprintf("Startup blocked: %s", msg))
						publishStatus(svc, serviceStateBlocked, false, 0, port, ownershipString(portOwnershipThisService), msg)
						return fmt.Errorf("service %s port %s is already in use", svc.DisplayName, svc.Port)
					}
				case portOwnershipOtherInfra:
					msg := formatConflictMessage(svc.DisplayName, probe)
					log.Warn("Port in use by another infra session", "service", svc.DisplayName, "port", svc.Port, "detail", msg)
					publishAction(svc, actionStartupBlocked, fmt.Sprintf("Startup blocked: %s", msg))
					publishStatus(svc, serviceStateBlocked, false, 0, port, ownershipString(portOwnershipOtherInfra), msg)
					return fmt.Errorf("service %s port %s is already in use", svc.DisplayName, svc.Port)
				case portOwnershipExternal:
					msg := formatConflictMessage(svc.DisplayName, probe)
					log.Error("❌ Port in use by external process", "service", svc.DisplayName, "port", svc.Port, "detail", msg)
					publishAction(svc, actionStartupBlocked, fmt.Sprintf("Startup blocked: %s", msg))
					publishStatus(svc, serviceStateBlocked, false, 0, port, ownershipString(portOwnershipExternal), msg)
					return fmt.Errorf("service %s port %s is already in use", svc.DisplayName, svc.Port)
				}
			}
		} else {
			publishStatus(svc, serviceStatePending, false, 0, port, "unknown", "")
		}

		cleanup, err := svc.Start(ctx, opts, recordErr)
		if cleanup != nil {
			cleanupStack = append(cleanupStack, cleanup)
		}
		if err != nil {
			msg := fmt.Sprintf("Start failed: %v", err)
			log.Warn("Service failed to start", "service", svc.DisplayName, "error", err)
			recordErr(fmt.Errorf("%s failed to start: %w", svc.DisplayName, err))
			publishAction(svc, actionStartFailed, msg)
			publishStatus(svc, serviceStateError, false, 0, port, "unknown", msg)
			continue
		}

		pid := 0
		if candidatePID, ok := goreman.GetProcessPID(string(svc.ID)); ok {
			pid = candidatePID
		}

		successMsg := "Service started"
		if svc.Port != "" {
			log.Info("✅ Service started", "service", svc.DisplayName, "port", svc.Port)
			successMsg = fmt.Sprintf("Service running on port %s", svc.Port)
		} else {
			log.Info("✅ Service started", "service", svc.DisplayName)
		}
		if len(startNotes) > 0 {
			successMsg = strings.Join(append(startNotes, successMsg), "; ")
		}

		publishAction(svc, actionStarted, successMsg)
		publishStatus(svc, serviceStateRunning, true, pid, port, ownershipString(portOwnershipThisService), "")
	}

	NotifyCaddyRoutesChanged()

	log.Info("📊 External services started with goreman supervision")
	for name, stat := range goreman.GetAllStatus() {
		log.Info("External process status", "name", name, "status", stat)
	}

	log.Info("🎉 All infrastructure services started successfully!")
	log.Info("💡 Web server accessible at http://0.0.0.0:" + config.GetWebServerPort())

	<-ctx.Done()

	select {
	case err := <-errCh:
		return err
	default:
		return nil
	}
}

func publishAction(spec ServiceSpec, kind, message string) {
	if message == "" && kind == "" {
		return
	}
	runtimeevents.Publish(runtimeevents.ServiceAction{
		TS:        time.Now(),
		ServiceID: string(spec.ID),
		Message:   message,
		Kind:      kind,
	})
}

func publishStatus(spec ServiceSpec, state string, running bool, pid int, port int, ownership string, message string) {
	if state == "" {
		if running {
			state = serviceStateRunning
		} else {
			state = serviceStatePending
		}
	}
	runtimeevents.Publish(runtimeevents.ServiceStatus{
		TS:        time.Now(),
		ServiceID: string(spec.ID),
		Running:   running,
		PID:       pid,
		Port:      port,
		Ownership: ownership,
		State:     state,
		Message:   message,
	})
}

func ownershipString(ownership svcports.Ownership) string {
	switch ownership {
	case svcports.OwnershipFree:
		return "free"
	case svcports.OwnershipThisService:
		return "this"
	case svcports.OwnershipOtherInfra:
		return "infra"
	case svcports.OwnershipExternal:
		return "external"
	default:
		return "unknown"
	}
}
