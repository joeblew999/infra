package handler

import (
	"net/http"

	"github.com/joeblew999/infra/pkg/api/deck/internal/logic"
	"github.com/joeblew999/infra/pkg/api/deck/internal/svc"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func ListDecksHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := logic.NewListDecksLogic(r.Context(), svcCtx)
		resp, err := l.ListDecks()
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
