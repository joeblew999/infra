package logic

import (
	"context"

	"github.com/joeblew999/infra/api/deck/internal/svc"
	"github.com/joeblew999/infra/api/deck/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type DownloadDeckLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewDownloadDeckLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DownloadDeckLogic {
	return &DownloadDeckLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DownloadDeckLogic) DownloadDeck(req *types.DownloadDeckRequest) (resp *types.DownloadDeckResponse, err error) {
	// todo: add your logic here and delete this line

	return
}
