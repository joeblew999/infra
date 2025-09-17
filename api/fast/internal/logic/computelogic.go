package logic

import (
	"context"

	"github.com/joeblew999/infra/api/fast/internal/svc"
	"github.com/joeblew999/infra/api/fast/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ComputeLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewComputeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ComputeLogic {
	return &ComputeLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ComputeLogic) Compute(req *types.ComputeRequest) (resp *types.ComputeResponse, err error) {
	// todo: add your logic here and delete this line

	return
}
