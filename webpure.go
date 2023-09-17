package main

import (
	"context"
	"flag"
	"github.com/gorilla/mux"
	"golang.org/x/sys/unix"
	"log"
	"net"
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
		conf = flag.String("f", "./", "config file path")
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
		for _, conf := range cfs {
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
			lc := net.ListenConfig{
				Control: func(network, address string, c syscall.RawConn) error {
					var socketErr error
					err := c.Control(func(fd uintptr) {
						socketErr = unix.SetsockoptInt(int(fd), unix.SOL_SOCKET, unix.SO_REUSEPORT, 1)
					})
					if err != nil {
						return err
					}
					return socketErr
				},
			}
			ln, err := lc.Listen(context.Background(), "tcp", pvHost)
			if err != nil {
				panic(err)
			}
			if conf.Ssl == "ssl" {
				go func() {
					if err := srv.ServeTLS(ln, conf.Pem, conf.key); err != nil {
						log.Println(err)
					}
				}()
			} else {
				go func() {
					if err := srv.Serve(ln); err != nil {
						log.Println(err)
					}
				}()
			}
			HostSets.Store(conf.ServerName+pvHost, &HostPayload{
				Had:  &spa,
				Svc:  srv,
				Conf: conf,
			})
			log.Println(conf.ServerName + pvHost)
		}
	}
}
