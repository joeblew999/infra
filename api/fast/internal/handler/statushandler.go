package handler

import (
	"net/http"

	"github.com/joeblew999/infra/api/fast/internal/logic"
	"github.com/joeblew999/infra/api/fast/internal/svc"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func StatusHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := logic.NewStatusLogic(r.Context(), svcCtx)
		resp, err := l.Status()
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
