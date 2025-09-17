package handler

import (
	"net/http"

	"github.com/joeblew999/infra/api/testservice/internal/logic"
	"github.com/joeblew999/infra/api/testservice/internal/svc"
	"github.com/joeblew999/infra/api/testservice/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func McpHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.McpRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := logic.NewMcpLogic(r.Context(), svcCtx)
		resp, err := l.Mcp(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
