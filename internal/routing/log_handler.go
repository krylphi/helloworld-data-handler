package routing

import (
	"github.com/krylphi/helloworld-data-handler/internal/domain"
	"github.com/krylphi/helloworld-data-handler/internal/utils"
	"github.com/valyala/fasthttp"
	"log"
)

func (r *Router) handleLog(ctx *fasthttp.RequestCtx) {
	switch {
	case ctx.IsPost():
		r.handlePostLog(ctx)
	case ctx.IsGet():
		r.handleGetLog(ctx)
	default:
		ctx.Error("method is not allowed", fasthttp.StatusMethodNotAllowed)
	}
}

func (r *Router) handlePostLog(ctx *fasthttp.RequestCtx) {
	//log.Print(utils.Concat("start: ", time.Now().Format(time.StampMilli)))
	data := ctx.Request.Body()
	entry, err := domain.ParseEntry(data)
	if err != nil {
		r := utils.Concat("err: ", err.Error(), "data: ", string(data))
		log.Print(r)
		ctx.Error(r, fasthttp.StatusBadRequest)
		return
	}
	r.streamHandler.Send(entry)
	ctx.Success("text/plain", []byte("OK"))
	//log.Print(utils.Concat("end: ", time.Now().Format(time.StampMilli)))
}

func (r *Router) handleGetLog(ctx *fasthttp.RequestCtx) {
	ctx.Success("text/plain", []byte("OK"))
}
