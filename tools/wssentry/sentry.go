package wssentry

import (
	"strings"

	"github.com/getsentry/sentry-go"
)

var DefaultSentryInstance = New()

type SentryInstance struct {
	opts Options
}

const (
	FilterKeywordSentryInstance = "wssentry"
	FilterKeywordGf             = "github.com/gogf/gf/"
)

func (s *SentryInstance) Init() error {
	if s.opts.dsn == "" {
		return nil
	}
	err := sentry.Init(sentry.ClientOptions{
		Dsn: s.opts.dsn,
		// Set tracesSampleRate to 1.0 to capture 100%
		// of transactions for performance monitoring.
		// We recommend adjusting this value in production,
		Debug:            s.opts.debug,
		Environment:      string(s.opts.env),
		TracesSampleRate: s.opts.tracesSampleRate,
		BeforeSend: func(event *sentry.Event, hint *sentry.EventHint) *sentry.Event {
			return filterAlertWrapper(event)
		},
	})
	return err
}

func New(opts ...Option) *SentryInstance {
	options := NewOptions(opts...)
	return &SentryInstance{
		opts: options,
	}
}

// nolint
func filterAlertWrapperFrames(frames []sentry.Frame) []sentry.Frame {
	filteredFrames := make([]sentry.Frame, 0, len(frames))
	for _, frame := range frames {
		if strings.Contains(frame.Module, FilterKeywordSentryInstance) || strings.Contains(frame.Module, FilterKeywordGf) {
			continue
		}
		filteredFrames = append(filteredFrames, frame)
	}
	return filteredFrames
}

// nolint
func filterAlertWrapper(event *sentry.Event) *sentry.Event {
	for _, ex := range event.Exception {
		if ex.Stacktrace == nil {
			continue
		}
		ex.Stacktrace.Frames = filterAlertWrapperFrames(ex.Stacktrace.Frames)
	}
	// This interface is used when we extract stacktrace from caught strings, eg. in panics
	for _, th := range event.Threads {
		if th.Stacktrace == nil {
			continue
		}
		th.Stacktrace.Frames = filterAlertWrapperFrames(th.Stacktrace.Frames)
	}
	return event
}
