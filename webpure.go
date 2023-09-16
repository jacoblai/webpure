package main

import (
	"flag"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"
)

var HostSets sync.Map

type HostPayload struct {
	Had  *SHandler
	Svc  *http.Server
	Conf Config
}

func init() {
	var rLimit syscall.Rlimit
	if err := syscall.Getrlimit(syscall.RLIMIT_NOFILE, &rLimit); err != nil {
		log.Fatal(err)
	}
	rLimit.Cur = rLimit.Max
	if err := syscall.Setrlimit(syscall.RLIMIT_NOFILE, &rLimit); err != nil {
		log.Fatal(err)
	}
}

func main() {
	var (
		conf = flag.String("f", "", "config file path")
	)
	flag.Parse()
	log.Println("WebPure V1.0.0")
	if *conf == "" {
		panic("webpure config file not found...")
	}
	initConfig(*conf)
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
	configs := GetConfig()
	if len(configs) <= 0 {
		log.Fatal("webpure not has website config start...")
	}
	for _, cfs := range configs {
		go func(conFs []Config) {
			for _, conf := range conFs {
				if !strings.HasSuffix(conf.Root, "/") {
					conf.Root += "/"
				}
				router := mux.NewRouter()
				spa := SHandler{StaticPath: conf.Root, IndexPage: conf.Index}
				router.PathPrefix(conf.Location).Handler(spa)
				pvHost := ":" + conf.Listen
				srv := &http.Server{
					Handler:      router,
					Addr:         pvHost,
					WriteTimeout: 5 * time.Second,
					ReadTimeout:  5 * time.Second,
				}
				has := false //share port more than one site
				HostSets.Range(func(_, v any) bool {
					if v.(*HostPayload).Conf.Listen == conf.Listen {
						has = true
						return false
					}
					return true
				})
				if has {
					continue
				}
				HostSets.Store(conf.ServerName+pvHost, &HostPayload{
					Had:  &spa,
					Svc:  srv,
					Conf: conf,
				})
				log.Println(pvHost)
				if conf.Ssl == "ssl" {
					go func() {
						if err := srv.ListenAndServeTLS(conf.Pem, conf.key); err != nil {
							log.Println(err)
						}
					}()
				} else {
					go func() {
						if err := srv.ListenAndServe(); err != nil {
							log.Println(err)
						}
					}()
				}
			}
		}(cfs)
	}
}
