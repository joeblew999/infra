package logic

import (
	"context"

	"github.com/joeblew999/infra/api/testservice/internal/svc"
	"github.com/joeblew999/infra/api/testservice/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type McpLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewMcpLogic(ctx context.Context, svcCtx *svc.ServiceContext) *McpLogic {
	return &McpLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *McpLogic) Mcp(req *types.McpRequest) (resp *types.McpResponse, err error) {
	// todo: add your logic here and delete this line

	return
}
