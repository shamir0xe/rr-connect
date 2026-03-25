package services

import "github.com/spf13/viper"

type RouterInterface interface {
	Next() string
}

type routerStruct struct {
	routers []string
	index   int
}

func NewRouter(cfg *viper.Viper) (RouterInterface, error) {
	return &routerStruct{
		routers: cfg.GetStringSlice("routers"),
		index:   0,
	}, nil
}

func (r *routerStruct) Next() string {
	if len(r.routers) == 0 {
		return ""
	}
	router := r.routers[r.index]
	r.index++
	r.index = (r.index + 1) % len(r.routers)
	return router
}
