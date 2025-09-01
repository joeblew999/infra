package logic

import (
	"context"

	"github.com/joeblew999/infra/api/deck/internal/svc"
	"github.com/joeblew999/infra/api/deck/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetDeckStatusLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetDeckStatusLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetDeckStatusLogic {
	return &GetDeckStatusLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetDeckStatusLogic) GetDeckStatus(req *types.GetDeckStatusRequest) (resp *types.GetDeckStatusResponse, err error) {
	// todo: add your logic here and delete this line

	return
}
