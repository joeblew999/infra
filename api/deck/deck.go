package main

import (
	"flag"
	"fmt"

	"github.com/joeblew999/infra/pkg/api/deck/internal/config"
	"github.com/joeblew999/infra/pkg/api/deck/internal/handler"
	"github.com/joeblew999/infra/pkg/api/deck/internal/svc"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/rest"
)

var configFile = flag.String("f", "etc/deck-api.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)

	server := rest.MustNewServer(c.RestConf)
	defer server.Stop()

	ctx := svc.NewServiceContext(c)
	handler.RegisterHandlers(server, ctx)

	fmt.Printf("Starting server at %s:%d...\n", c.Host, c.Port)
	server.Start()
}
