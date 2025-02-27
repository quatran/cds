package api

import (
	"context"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/rockbears/log"

	"github.com/ovh/cds/engine/service"
	"github.com/ovh/cds/sdk"
	cdslog "github.com/ovh/cds/sdk/log"
)

func (api *API) rbacMiddleware(ctx context.Context, w http.ResponseWriter, req *http.Request, rc *service.HandlerConfig) (context.Context, error) {
	for _, checker := range rc.RbacCheckers {
		ctx := context.WithValue(ctx, cdslog.RbackCheckerName, sdk.GetFuncName(checker))
		if err := checker(ctx, getAPIConsumer(ctx), api.Cache, api.mustDB(), mux.Vars(req)); err != nil {
			if isAdmin(ctx) {
				trackSudo(ctx, w)
				return ctx, nil
			}
			log.ErrorWithStackTrace(ctx, err)
			return ctx, sdk.WithStack(sdk.ErrForbidden)
		}
	}
	return ctx, nil
}
