package sentry_instance

import (
	"github.com/getsentry/sentry-go"
)

var DefaultSentryInstance = New()

type SentryInstance struct {
	opts Options
}

func (s *SentryInstance) Init() error {
	err := sentry.Init(sentry.ClientOptions{
		Dsn: "https://51da076279386f0174c2d3237aeb657e@o4506597786517504.ingest.sentry.io/4506597788614656",
		// Set TracesSampleRate to 1.0 to capture 100%
		// of transactions for performance monitoring.
		// We recommend adjusting this value in production,
		TracesSampleRate: 1.0,
	})
	return err
}

func New(opts ...Option) *SentryInstance {
	options := NewOptions(opts...)
	return &SentryInstance{
		opts: options,
	}
}
