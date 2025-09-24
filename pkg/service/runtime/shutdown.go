package runtime

import (
	"time"

	"github.com/joeblew999/infra/pkg/log"
	svcports "github.com/joeblew999/infra/pkg/service/ports"
)

// Shutdown stops running infrastructure services by signalling processes, ports, and goreman groups.
func Shutdown() {
	log.Info("🛑 Shutting down all infrastructure services...")

	log.Info("🔍 Looking for main service process...")
	mainProcessKilled := false

	if err := svcports.KillInfraGoRunProcess(); err == nil {
		log.Info("✅ Sent shutdown signal to infra go run process")
		mainProcessKilled = true
		time.Sleep(2 * time.Second)
	}

	if err := svcports.KillProcessByName("infra"); err == nil {
		log.Info("✅ Sent graceful shutdown signal to infra binary process")
		mainProcessKilled = true
		time.Sleep(1 * time.Second)
	}

	log.Info("🔌 Shutting down services by port...")
	specs := activeServiceSpecs
	if len(specs) == 0 {
		specs = buildServiceSpecs(activeOptions)
	}
	portSpecs := collectServicePortsForSpecs(specs)
	portSpecs = append(portSpecs, ServicePort{Service: "HTTPS Proxy", Port: "443"})

	portsKilled := 0
	for _, spec := range portSpecs {
		port := svcports.ParsePort(spec.Port)
		if port == 0 {
			continue
		}
		if err := svcports.KillProcessByPort(port); err == nil {
			log.Info("✅ Stopped service port", "service", spec.Service, "port", spec.Port)
			portsKilled++
		}
	}

	log.Info("📝 Shutting down goreman-supervised processes...")
	processNames := collectGoremanProcessesForSpecs(specs)
	processNames = append(processNames, "infra")

	processesKilled := 0
	for _, name := range processNames {
		if err := svcports.KillProcessByName(name); err == nil {
			log.Info("✅ Stopped process", "name", name)
			processesKilled++
		}
	}

	if mainProcessKilled {
		log.Info("✅ Main service process shutdown gracefully")
	}
	if portsKilled > 0 {
		log.Info("✅ Stopped services on ports", "count", portsKilled)
	}
	if processesKilled > 0 {
		log.Info("✅ Stopped processes by name", "count", processesKilled)
	}

	if mainProcessKilled || portsKilled > 0 || processesKilled > 0 {
		log.Info("🎉 All infrastructure services shutdown complete!")
	} else {
		log.Info("ℹ️  No running services found to shutdown")
	}

	for _, spec := range specs {
		port := svcports.ParsePort(spec.Port)
		publishAction(spec, actionShutdown, "Service shutdown requested")
		publishStatus(spec, serviceStateStopped, false, 0, port, "unknown", "Service stopped")
	}
}
