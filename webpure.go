package main

import (
	"context"
	"flag"
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"log"
	"syscall"
	"time"
)

func init() {
	var rLimit syscall.Rlimit
	if err := syscall.Getrlimit(syscall.RLIMIT_NOFILE, &rLimit); err != nil {
		log.Fatal(err)
	}
	rLimit.Cur = rLimit.Max
	if err := syscall.Setrlimit(syscall.RLIMIT_NOFILE, &rLimit); err != nil {
		log.Fatal(err)
	}
	var (
		conf = flag.String("f", "", "config file path")
	)
	flag.Parse()
	if *conf == "" {
		panic("webpure config file not found...")
	}
	initConfig(*conf)
}

func main() {
	conf := GetConfig()
	h := server.Default(
		server.WithHostPorts(":"+conf.Addr),
		server.WithExitWaitTime(time.Second),
	)
	h.StaticFS(conf.Location, &app.FS{
		Root:               conf.Root + "/",
		IndexNames:         []string{conf.Index},
		GenerateIndexPages: false,
		Compress:           false,
		PathNotFound:       NotFound,
	})
	h.Spin()
}

func NotFound(c context.Context, ctx *app.RequestContext) {
	ctx.String(consts.StatusOK, "page not found")
}
