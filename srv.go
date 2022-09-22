package http_server

import (
	"embed"
	"github.com/Nixson/environment"
	"github.com/Nixson/http-server/server"
)

func InitSever(emb embed.FS, env *environment.Env) {
	server.InitServer(emb, env)
}

func InitController(name string, controller *server.ContextInterface) {
	server.InitController(name, controller)
}
