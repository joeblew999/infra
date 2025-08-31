package handler

import (
	"net/http"

	"github.com/joeblew999/infra/pkg/api/deck/internal/logic"
	"github.com/joeblew999/infra/pkg/api/deck/internal/svc"
	"github.com/joeblew999/infra/pkg/api/deck/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func GenerateDeckHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.GenerateDeckRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := logic.NewGenerateDeckLogic(r.Context(), svcCtx)
		resp, err := l.GenerateDeck(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
