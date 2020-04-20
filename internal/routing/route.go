package routing

import (
	"log"

	"github.com/valyala/fasthttp"

	"github.com/krylphi/helloworld-data-handler/internal/stream"
)

// Router http router
type Router struct {
	streamHandler stream.Handler
}

// NewRouter router constructor
func NewRouter(handler stream.Handler) Router {
	return Router{
		streamHandler: handler,
	}
}

// HTTPRouter contains routing paths. Methods are being checked in respective handlers
func (r *Router) HTTPRouter(ctx *fasthttp.RequestCtx) {
	switch string(ctx.Path()) {
	case "/log":
		r.handleLog(ctx)
	default:
		ctx.Error("not found", fasthttp.StatusNotFound)
	}
}

// Shutdown shutdown router and clear stream handler
func (r *Router) Shutdown() {
	log.Print("Flushing stream handler")
	r.streamHandler.Flush()
	log.Print("Stream handler flushed")
}
