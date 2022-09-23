package server

import (
	"context"
	"embed"
	"github.com/Nixson/annotation"
	"github.com/Nixson/environment"
	"github.com/Nixson/logNx"
	"net/http"
	"os"
)

func RunWithSignal() {
	done := make(chan os.Signal, 1)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	logNx.GetLogger().Info("listen " + srv.Addr)
	go func() {
		_ = srv.ListenAndServe()
	}()
	<-done
	logNx.GetLogger().Info("server done")
	_ = srv.Close()
	_ = srv.Shutdown(ctx)

}

func Run() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	logNx.GetLogger().Info("listen " + srv.Addr)
	_ = srv.ListenAndServe()
	logNx.GetLogger().Info("server done")
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
		Addr:           params.Env.GetString("server.port"),
		Handler:        mux,
		MaxHeaderBytes: params.Env.GetInt("server.maxSize"),
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
