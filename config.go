package main

import (
	"encoding/json"
	nginxparser "github.com/faceair/nginx-parser"
	"github.com/tidwall/gjson"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

var (
	config     Config
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
			loadConfig(conf, false)
			log.Println("Reloaded")
		}
	}()
}

func loadConfig(conf string, fail bool) {
	directives, err := nginxparser.New(nil).ParseFile(conf)
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
	config = Config{
		Addr:       addr.String(),
		Root:       rot.String(),
		Index:      idx.String(),
		ServerName: svName.String(),
		Location:   lc.String(),
	}
	configLock.Unlock()
}

func GetConfig() Config {
	configLock.RLock()
	defer configLock.RUnlock()
	return config
}
