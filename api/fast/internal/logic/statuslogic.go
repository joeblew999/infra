package logic

import (
	"context"

	"github.com/joeblew999/infra/api/fast/internal/svc"
	"github.com/joeblew999/infra/api/fast/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type StatusLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewStatusLogic(ctx context.Context, svcCtx *svc.ServiceContext) *StatusLogic {
	return &StatusLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *StatusLogic) Status() (resp *types.StatusResponse, err error) {
	// todo: add your logic here and delete this line

	return
}
