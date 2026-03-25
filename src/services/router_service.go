package services

import (
	"fmt"

	"github.com/spf13/viper"
)

type RouterInterface interface {
	Next() string
	Pick() string
}

type routerStruct struct {
	routers []string
	index   int
}

func NewRouter(cfg *viper.Viper) (RouterInterface, error) {
	routers := cfg.GetStringSlice("routers")
	fmt.Println("Routers:", routers)
	return &routerStruct{
		routers: routers,
		index:   0,
	}, nil
}

func (r *routerStruct) Pick() string {
	if len(r.routers) == 0 {
		return ""
	}
	return r.routers[r.index]
}

func (r *routerStruct) update() {
	r.index = (r.index + 1) % len(r.routers)
}

func (r *routerStruct) Next() string {
	router := r.Pick()
	r.update()
	return router
}
