package handler

import (
	"net/http"

	"github.com/joeblew999/infra/api/fast/internal/logic"
	"github.com/joeblew999/infra/api/fast/internal/svc"
	"github.com/joeblew999/infra/api/fast/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func ComputeHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.ComputeRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := logic.NewComputeLogic(r.Context(), svcCtx)
		resp, err := l.Compute(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
