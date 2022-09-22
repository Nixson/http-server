package http_server

import (
	"embed"
	"github.com/Nixson/environment"
	"github.com/Nixson/http-server/server"
)

func InitSever(emb embed.FS, env *environment.Env, funcsInit []func()) {
	server.InitServer(emb, env, funcsInit)
}
func RunServer() {
	server.Run()
}

func InitController(name string, controller *server.ContextInterface) {
	server.InitController(name, controller)
}
