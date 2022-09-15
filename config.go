package main

import (
	"context"
	"encoding/json"
	"github.com/cloudwego/hertz/pkg/app/server"
	nginxparser "github.com/faceair/nginx-parser"
	"github.com/tidwall/gjson"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"
)

var (
	config     = make(map[string]Config)
	configLock = new(sync.RWMutex)
)

type Config struct {
	ServerName string
	Addr       string
	Index      string
	Root       string
	Location   string
}

func initConfig(conf string) {
	loadConfig(conf, true)
	s := make(chan os.Signal, 1)
	signal.Notify(s, syscall.SIGUSR2)
	go func() {
		for {
			<-s
			closeAllSvc()
			loadConfig(conf, false)
			startSvc()
			log.Println("WebPure Reloaded")
		}
	}()
}

func closeAllSvc() {
	ctx, _ := context.WithTimeout(context.Background(), time.Second)
	for n, svc := range svcs {
		_ = svc.Shutdown(ctx)
		log.Println(n, "shutdown")
	}
	svcs = make(map[string]*server.Hertz)
}

func loadConfig(conf string, fail bool) {
	if strings.HasPrefix(conf, ".") {
		dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
		if err != nil {
			log.Println("open config: ", err)
			if fail {
				os.Exit(1)
			}
		}
		conf = strings.Replace(conf, ".", dir, 1)
	}

	if ok, _ := IsDirectory(conf); !ok {
		log.Println("open config: ", "path not directory")
		if fail {
			os.Exit(1)
		}
	}

	files, err := ioutil.ReadDir(conf)
	if err != nil {
		log.Println("open config: ", err)
		if fail {
			os.Exit(1)
		}
	}

	for _, file := range files {
		if filepath.Ext(file.Name()) != ".conf" {
			//log.Println(file.Name(), "config file must by ext (.conf)")
			continue
		}
		directives, err := nginxparser.New(nil).ParseFile(conf + file.Name())
		if err != nil {
			log.Println("open config: ", err)
			if fail {
				os.Exit(1)
			}
		}
		body, err := json.MarshalIndent(directives, "", "  ")
		if err != nil {
			log.Println("open config: ", err)
			if fail {
				os.Exit(1)
			}
		}

		temp := gjson.ParseBytes(body)
		addr := temp.Get(`0.block.#(directive="listen").args.0`)
		if !addr.Exists() {
			log.Println("open config: ", "listen not found")
			if fail {
				os.Exit(1)
			}
		}

		rot := temp.Get(`0.block.#(directive="root").args.0`)
		if !rot.Exists() {
			log.Println("open config: ", "root not found")
			if fail {
				os.Exit(1)
			}
		}

		idx := temp.Get(`0.block.#(directive="index").args.0`)
		if !rot.Exists() {
			log.Println("open config: ", "index not found")
			if fail {
				os.Exit(1)
			}
		}

		svName := temp.Get(`0.block.#(directive="server_name").args.0`)
		if !rot.Exists() {
			log.Println("open config: ", "server_name not found")
			if fail {
				os.Exit(1)
			}
		}

		lc := temp.Get(`0.block.#(directive="location").args.0`)
		if !rot.Exists() {
			log.Println("open config: ", "location not found")
			if fail {
				os.Exit(1)
			}
		}

		configLock.Lock()
		config[path.Base(file.Name())] = Config{
			Addr:       addr.String(),
			Root:       rot.String(),
			Index:      idx.String(),
			ServerName: svName.String(),
			Location:   lc.String(),
		}
		configLock.Unlock()
	}
}

func GetConfig() map[string]Config {
	configLock.RLock()
	defer configLock.RUnlock()
	return config
}
