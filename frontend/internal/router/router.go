package router

import "mini-workflow/frontend/internal/ports"

type Router struct {
	history ports.HistoryClient
}

func New(history ports.HistoryClient) *Router {
	return &Router{history: history}
}

func (r *Router) GetHistoryClient(_ string) ports.HistoryClient {
	return r.history
}
