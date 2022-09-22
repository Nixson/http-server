package server

import (
	"context"
	"embed"
	"github.com/Nixson/annotation"
	"github.com/Nixson/environment"
	"net/http"
	"os"
)

func Run() {
	done := make(chan os.Signal, 1)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() {
		_ = srv.ListenAndServe()
	}()
	<-done
	_ = srv.Close()
	_ = srv.Shutdown(ctx)

}

var srv *http.Server

func InitServer(emb embed.FS, env *environment.Env, funcsInit []func()) {
	params = &Params{
		Annotation: annotation.InitAnnotation(emb),
		Env:        env,
	}
	mux := http.NewServeMux()
	mux.Handle("/", http.HandlerFunc(handle))
	srv = &http.Server{
		Addr:    params.Env.GetString("server.port"),
		Handler: mux,
	}
	if funcsInit != nil && len(funcsInit) > 0 {
		for _, funcInit := range funcsInit {
			funcInit()
		}
	}
}

func handle(w http.ResponseWriter, r *http.Request) {
	ctx := Context{
		Request:  r,
		Response: w,
		Path:     r.URL.Path,
	}
	if ctx.IsGranted() {
		ctx.Call()
	}
}
