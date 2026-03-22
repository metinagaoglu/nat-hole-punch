package router

import (
	"udp-hole-punch/pkg/handlers"
)

func InitializeRoutes(ctx *handlers.HandlerContext) *Router {
	router := NewRouter()
	router.AddRoute("register", ctx.RegisterHandler())
	router.AddRoute("logout", ctx.LogoutHandler())
	return router
}
