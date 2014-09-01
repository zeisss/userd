package v2

import (
	servicePkg "../../service"

	middlewarePkg "github.com/catalyst-zero/middleware-server"
	logPkg "github.com/op/go-logging"
)

type Ctx struct{}

type V2 struct {
	Logger      *logPkg.Logger
	UserService *servicePkg.UserService
}

func (v2 *V2) SetupRoutes(srv *middlewarePkg.Server) {
	srv.Serve("POST", "/v2/user/", v2.CreateUser)
	srv.Serve("POST", "/v2/user/{userOrMail}/login", v2.LoginUser)
}
