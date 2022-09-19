package main

import (
	"context"
	"encoding/json"
	nginxparser "github.com/faceair/nginx-parser"
	"github.com/tidwall/gjson"
	"io/ioutil"
	"log"
	"net"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"
)

var (
	config     = make(map[string][]Config)
	configLock = new(sync.RWMutex)
)

type Config struct {
	ServerName string
	Listen     string
	Index      string
	Root       string
	Location   string
	Ssl        string
	Pem        string
	key        string
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
	HostSets.Range(func(_, v any) bool {
		_ = v.(*HostPayload).Svc.Shutdown(ctx)
		return true
	})
	config = make(map[string][]Config)
	HostSets = sync.Map{}
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
		listen := temp.Get(`0.block.#(directive="listen").args.0`)
		if !listen.Exists() {
			log.Println("open config: ", "listen not found")
			if fail {
				os.Exit(1)
			}
		}

		ssl := temp.Get(`0.block.#(directive="listen").args.1`)

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

		pem := temp.Get(`0.block.#(directive="ssl_certificate").args.0`)
		key := temp.Get(`0.block.#(directive="ssl_certificate_key").args.0`)

		addr := net.ParseIP(svName.String())
		if addr == nil {
			hostName, _ := net.LookupHost(svName.String())
			if len(hostName) == 0 {
				log.Println("err: load config", svName.String(), "is not ip or domain")
				continue
			}
		}
		configLock.Lock()
		cfs, ok := config[listen.String()]
		if !ok {
			cfs = make([]Config, 0)
		}
		cf := Config{
			Listen:     listen.String(),
			Root:       rot.String(),
			Index:      idx.String(),
			ServerName: svName.String(),
			Location:   lc.String(),
		}
		if ssl.Exists() {
			cf.Ssl = ssl.String()
		}
		if pem.Exists() {
			cf.Pem = pem.String()
		}
		if key.Exists() {
			cf.key = key.String()
		}
		cfs = append(cfs, cf)
		config[listen.String()] = cfs
		configLock.Unlock()
	}
}

func GetConfig() map[string][]Config {
	configLock.RLock()
	defer configLock.RUnlock()
	return config
}
