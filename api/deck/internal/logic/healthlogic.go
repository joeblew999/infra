package logic

import (
	"context"

	"github.com/joeblew999/infra/pkg/api/deck/internal/svc"
	"github.com/joeblew999/infra/pkg/api/deck/internal/types"

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
		toolStatus := types.ToolStatus{
			Name:      tool,
			Available: false, // TODO: Actually check if tools are available
			Version:   "",
		}
		status = append(status, toolStatus)
	}
	
	return status
}
