package logic

import (
	"context"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/joeblew999/infra/api/deck/internal/svc"
	"github.com/joeblew999/infra/api/deck/internal/types"
	"github.com/joeblew999/infra/pkg/config"

	"github.com/zeromicro/go-zero/core/logx"
)

type HealthLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewHealthLogic(ctx context.Context, svcCtx *svc.ServiceContext) *HealthLogic {
	return &HealthLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *HealthLogic) Health() (resp *types.HealthResponse, err error) {
	resp = &types.HealthResponse{
		Status:  "ok",
		Version: "v1.0",
		Tools:   l.checkDeckTools(),
	}
	
	return resp, nil
}

// checkDeckTools verifies availability of deck tools
func (l *HealthLogic) checkDeckTools() []types.ToolStatus {
	tools := []string{"decksh", "svgdeck", "pngdeck", "pdfdeck", "dshfmt", "dshlint"}
	var status []types.ToolStatus

	for _, tool := range tools {
		available, version := l.checkToolAvailability(tool)
		toolStatus := types.ToolStatus{
			Name:      tool,
			Available: available,
			Version:   version,
		}
		status = append(status, toolStatus)
	}

	return status
}

// checkToolAvailability checks if a tool is available and gets its version
func (l *HealthLogic) checkToolAvailability(toolName string) (bool, string) {
	// First check if tool is in .dep directory (managed by dep system)
	depPath := filepath.Join(config.GetDepPath(), toolName)
	if cmd := exec.Command(depPath, "--version"); cmd != nil {
		if output, err := cmd.Output(); err == nil {
			version := strings.TrimSpace(string(output))
			return true, version
		}
	}

	// Then check if tool is in PATH
	if _, err := exec.LookPath(toolName); err == nil {
		if cmd := exec.Command(toolName, "--version"); cmd != nil {
			if output, err := cmd.Output(); err == nil {
				version := strings.TrimSpace(string(output))
				return true, version
			}
		}
		// Tool exists but version check failed - still consider it available
		return true, "unknown"
	}

	return false, ""
}
