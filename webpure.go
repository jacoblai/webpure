package main

import (
	"context"
	"flag"
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
)

var svcs = make(map[string]*server.Hertz)

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
	dir, _ := filepath.Abs(filepath.Dir(os.Args[0]))
	f, err := os.OpenFile(dir+"/log.txt", os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	hlog.SetOutput(f)
	log.Println("WebPure V1.0.0")
	startSvc()
	signalChan := make(chan os.Signal, 1)
	cleanupDone := make(chan bool)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		for range signalChan {
			closeAllSvc()
			log.Println("safe exit")
			cleanupDone <- true
		}
	}()
	<-cleanupDone
}

func startSvc() {
	confs := GetConfig()
	if len(confs) <= 0 {
		log.Fatal("webpure not has website config start...")
	}
	for _, cf := range confs {
		log.Println(cf.ServerName, cf.Addr, "online")
		go func(conf Config) {
			h := server.Default(
				server.WithHostPorts(":" + conf.Addr),
			)
			if !strings.HasSuffix(conf.Root, "/") {
				conf.Root += "/"
			}
			h.StaticFS(conf.Location, &app.FS{
				Root:               conf.Root,
				IndexNames:         []string{conf.Index},
				GenerateIndexPages: false,
				Compress:           false,
				PathNotFound:       NotFound,
			})
			svcs[conf.ServerName] = h
			err := h.Run()
			if err != nil {
				log.Println(conf.ServerName, err)
			}
		}(cf)
	}
}

func NotFound(c context.Context, ctx *app.RequestContext) {
	ctx.String(consts.StatusOK, "page not found")
}
