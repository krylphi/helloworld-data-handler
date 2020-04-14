package routing

import (
	"github.com/krylphi/helloworld-data-handler/internal/stream/aws"
	"github.com/valyala/fasthttp"
	"log"
)

type Router struct {
	streamHandler *aws.StreamHandler
}

func NewRouter() Router {
	return Router{
		streamHandler: aws.NewStreamHandler(),
	}
}

// HttpRouter contains routing paths. Methods are being checked in respective handlers
func (r *Router) HttpRouter(ctx *fasthttp.RequestCtx) {
	switch string(ctx.Path()) {
	case "/log":
		r.handleLog(ctx)
	default:
		ctx.Error("not found", fasthttp.StatusNotFound)
	}
}

func (r *Router) Shutdown() {
	log.Print("Flushing stream handler")
	r.streamHandler.Flush()
	log.Print("Stream handler flushed")
}
