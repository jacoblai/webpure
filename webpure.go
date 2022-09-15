package main

import (
	"context"
	"flag"
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"log"
	"os"
	"path/filepath"
	"time"
)

func main() {
	var (
		addr    = flag.String("l", ":8080", "绑定Host地址")
		webPath = flag.String("p", "*", "前端静态文件夹本地绝对路径，默认值表示以当前程序所在路径为服务")
	)
	flag.Parse()
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Fatal(err)
	}
	if *webPath != "*" {
		dir = *webPath
	}
	h := server.Default(
		server.WithHostPorts(*addr),
		server.WithExitWaitTime(time.Second),
	)
	h.StaticFS("/", &app.FS{
		Root:               dir + "/",
		IndexNames:         []string{"index.html"},
		GenerateIndexPages: false,
		Compress:           false,
		PathNotFound:       NotFound,
	})
	h.Spin()
}

func NotFound(c context.Context, ctx *app.RequestContext) {
	ctx.String(consts.StatusOK, "page not found")
}
