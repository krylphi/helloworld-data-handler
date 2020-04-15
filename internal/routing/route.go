package routing

import (
	"log"

	"github.com/krylphi/helloworld-data-handler/internal/stream/aws"

	"github.com/valyala/fasthttp"
)

// Router http router
type Router struct {
	streamHandler *aws.StreamHandler
}

// NewRouter router constructor
func NewRouter() Router {
	return Router{
		streamHandler: aws.NewStreamHandler(),
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
