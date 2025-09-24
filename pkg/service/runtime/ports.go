package runtime

import (
	"strconv"
	"strings"

	"github.com/joeblew999/infra/pkg/config"
	"github.com/joeblew999/infra/pkg/goreman"
	"github.com/joeblew999/infra/pkg/log"
	"github.com/joeblew999/infra/pkg/service/ports"
)

const (
	portOwnershipFree        = ports.OwnershipFree
	portOwnershipThisService = ports.OwnershipThisService
	portOwnershipOtherInfra  = ports.OwnershipOtherInfra
	portOwnershipExternal    = ports.OwnershipExternal
)

func inspectPort(serviceID ServiceID, port int) (*ports.Probe, error) {
	var expectedPID string
	if pid, ok := goreman.GetProcessPID(string(serviceID)); ok && pid != 0 {
		expectedPID = strconv.Itoa(pid)
	}
	return ports.Inspect(port, expectedPID)
}

func reclaimManagedProcess(serviceID ServiceID, probe *ports.Probe) bool {
	if probe == nil || probe.Ownership != ports.OwnershipThisService {
		return false
	}
	if err := goreman.Stop(string(serviceID)); err == nil {
		log.Info("Reclaimed existing process", "service", serviceID, "pid", probe.PID)
		return true
	}
	if probe.PID != "" {
		if err := ports.KillProcess(probe.PID); err == nil {
			log.Info("Killed lingering process by PID", "service", serviceID, "pid", probe.PID)
			return true
		}
	}
	return false
}

func ownershipForService(spec ServiceSpec, probe *ports.Probe) ports.Ownership {
	if probe == nil {
		return portOwnershipFree
	}
	ownership := probe.Ownership
	if ownership == ports.OwnershipThisService {
		return ownership
	}
	if matchesServiceCommand(spec, probe.Command) {
		probe.Ownership = ports.OwnershipThisService
		return portOwnershipThisService
	}
	return ownership
}

func matchesServiceCommand(spec ServiceSpec, command string) bool {
	if command == "" {
		return false
	}
	needle := string(spec.ID)
	if needle != "" && strings.Contains(command, needle) {
		return true
	}
	for _, name := range spec.GoremanProcesses {
		if name != "" && strings.Contains(command, name) {
			return true
		}
	}
	return false
}

func formatConflictMessage(service string, probe *ports.Probe) string {
	return ports.FormatConflictMessage(service, probe)
}

func shouldAutoReclaim(serviceID ServiceID) bool {
	return config.IsDevelopment()
}
