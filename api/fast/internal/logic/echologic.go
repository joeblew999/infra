package logic

import (
	"context"

	"github.com/joeblew999/infra/api/fast/internal/svc"
	"github.com/joeblew999/infra/api/fast/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type EchoLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewEchoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *EchoLogic {
	return &EchoLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *EchoLogic) Echo(req *types.EchoRequest) (resp *types.EchoResponse, err error) {
	// todo: add your logic here and delete this line

	return
}
