package handler

import (
	"net/http"

	"github.com/joeblew999/infra/api/deck/internal/logic"
	"github.com/joeblew999/infra/api/deck/internal/svc"
	"github.com/joeblew999/infra/api/deck/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func GetDeckStatusHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.GetDeckStatusRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := logic.NewGetDeckStatusLogic(r.Context(), svcCtx)
		resp, err := l.GetDeckStatus(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
