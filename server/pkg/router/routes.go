package router

import (
	h "udp-hole-punch/pkg/handlers"
)

func InitializeRoutes() *Router {
	router := NewRouter()
	router.AddRoute("register", h.Register)
	router.AddRoute("logout", h.Logout)
	return router
}
