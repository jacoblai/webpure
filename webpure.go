package main

import (
	"context"
	"flag"
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"log"
	"sync"
	"syscall"
	"time"
)

var wg sync.WaitGroup

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
	confs := GetConfig()
	if len(confs) <= 0 {
		log.Fatal("webpure not has website config start...")
	}
	for _, cf := range confs {
		log.Println(cf.ServerName, "site online...")
		wg.Add(1)
		go func(conf Config) {
			defer wg.Done()
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
		}(cf)
	}
	wg.Wait()
}

func NotFound(c context.Context, ctx *app.RequestContext) {
	ctx.String(consts.StatusOK, "page not found")
}
