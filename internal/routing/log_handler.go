package routing

import (
	"log"

	"github.com/krylphi/helloworld-data-handler/internal/domain"
	"github.com/krylphi/helloworld-data-handler/internal/errs"
	"github.com/krylphi/helloworld-data-handler/internal/utils"

	"github.com/valyala/fasthttp"
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
	err = r.streamHandler.Send(entry)
	if err != nil {
		r := utils.Concat("err: ", err.Error(), "data: ", string(data))
		log.Print(r)
		if err == errs.ErrServerShuttingDown {
			ctx.Error(r, fasthttp.StatusInternalServerError)
			return
		}
		ctx.Error(r, fasthttp.StatusBadRequest)
		return
	}
	ctx.Success("text/plain", []byte("OK"))
	//log.Print(utils.Concat("end: ", time.Now().Format(time.StampMilli)))
}

func (r *Router) handleGetLog(ctx *fasthttp.RequestCtx) {
	ctx.Success("text/plain", []byte("OK"))
}
