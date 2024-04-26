package scratch

import (
	"log/slog"
	"net/http"
)

func (a *App) WithCustomRestHandler(handler http.Handler) *App {
	if a == nil {
		slog.Error("Info", "a", a)
		return a
	}
	a.opts.EnableCustomHandler = true
	a.opts.CustomHandler = handler

	return a
}

func (a *App) WithCustomRestMiddleware(mwFuncs ...func(handler http.Handler) http.Handler) *App {
	for _, mwFunc := range mwFuncs {
		a.opts.EnableCustomMuxMiddleware = true
		a.opts.CustomMuxMiddleware = append(a.opts.CustomMuxMiddleware, mwFunc)
	}
	return a
}

func (a *App) WithGatewayMiddleware(mwFuncs ...func(handler http.Handler) http.Handler) *App {
	for _, mwFunc := range mwFuncs {
		a.opts.EnableGatewayMiddleware = true
		a.opts.GatewayMiddleware = append(a.opts.GatewayMiddleware, mwFunc)
	}
	return a
}

func (a *App) WithPublicMuxMiddleware(mwFuncs ...func(handler http.Handler) http.Handler) *App {
	for _, mwFunc := range mwFuncs {
		a.opts.EnablePublicMuxMiddleware = true
		a.opts.PublicMuxMiddleware = append(a.opts.PublicMuxMiddleware, mwFunc)
	}
	return a
}
