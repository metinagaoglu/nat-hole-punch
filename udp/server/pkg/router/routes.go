package router

import (
	h "udp-hole-punch/pkg/handlers"
)

func InıtializeRoutes() *Router {
	router := NewRouter()
	router.AddRoute("register", h.Register)
	return router
}
