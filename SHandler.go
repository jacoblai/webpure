package main

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

type SHandler struct {
	StaticPath string
	IndexPage  string
}

func (h SHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	k := r.Host
	ur := strings.Split(r.Host, ":")
	if len(ur) == 1 {
		if r.TLS != nil {
			k += ":443"
		} else {
			k += ":80"
		}
	}

	if playLoad, ok := HostSets.Load(k); ok && playLoad != nil {
		handler := playLoad.(*HostPayload).Had
		path, err := filepath.Abs(r.URL.Path)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		path = filepath.Join(handler.StaticPath, path)
		_, err = os.Stat(path)
		if os.IsNotExist(err) {
			http.ServeFile(w, r, filepath.Join(handler.StaticPath, handler.IndexPage))
			return
		} else if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		http.FileServer(http.Dir(handler.StaticPath)).ServeHTTP(w, r)
	} else {
		http.Error(w, "Forbidden", http.StatusForbidden)
	}
}
