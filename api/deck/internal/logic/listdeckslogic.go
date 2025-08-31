package logic

import (
	"context"

	"github.com/joeblew999/infra/pkg/api/deck/internal/svc"
	"github.com/joeblew999/infra/pkg/api/deck/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListDecksLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListDecksLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListDecksLogic {
	return &ListDecksLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListDecksLogic) ListDecks() (resp *types.ListDecksResponse, err error) {
	// todo: add your logic here and delete this line

	return
}
