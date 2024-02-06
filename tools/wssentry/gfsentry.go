package wssentry

import (
	"context"
	"time"

	"github.com/gogf/gf/v2/net/ghttp"

	"github.com/getsentry/sentry-go"
)

var GfSentry *Handler = newGfSentry(true, DefaultSentryInstance, 3*time.Second)

type Handler struct {
	rePanic         bool
	waitForDelivery bool
	timeout         time.Duration
	sentry          *SentryInstance
}

// The identifier of the Iris SDK.
const sdkIdentifier = "sentry.go.gf"

const valuesKey = "sentry"

func newGfSentry(rePanic bool, sentry *SentryInstance, timeout time.Duration) *Handler {
	return &Handler{
		rePanic:         rePanic,
		timeout:         timeout,
		waitForDelivery: false,
		sentry:          sentry,
	}
}

func (h *Handler) MiddleWare(r *ghttp.Request) {

	if h.sentry.opts.dsn != "" {
		hub := sentry.GetHubFromContext(r.Context())
		if hub == nil {
			hub = sentry.CurrentHub().Clone()
		}
		if client := hub.Client(); client != nil {
			client.SetSDKIdentifier(sdkIdentifier)
		}
		hub.Scope().SetRequest(r.Request)
		r.SetCtxVar(valuesKey, hub)

	}
	r.Middleware.Next()
}

func (h *Handler) RecoverHttp(r *ghttp.Request, handle func(r *ghttp.Request)) {
	if h.sentry.opts.dsn != "" {
		defer h.recoverWithSentry(GetHubFromContext(r), r)
	}
	handle(r)
}

func (h *Handler) recoverWithSentry(hub *sentry.Hub, r *ghttp.Request) {
	err := recover()
	if err != nil {
		eventID := hub.RecoverWithContext(
			context.WithValue(r.Context(), sentry.RequestContextKey, r),
			err,
		)
		if eventID != nil && h.waitForDelivery {
			hub.Flush(h.timeout)
		}
		if h.rePanic {
			panic(err)
		}
	}
}

// GetHubFromContext retrieves attached *sentry.Hub instance from goframe.Context.
func GetHubFromContext(r *ghttp.Request) *sentry.Hub {
	if hub, ok := r.Context().Value(valuesKey).(*sentry.Hub); ok {
		return hub
	}
	return nil
}
