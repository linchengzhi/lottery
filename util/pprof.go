package util

import (
	"fmt"
	"net/http"
	"net/http/pprof"
)

func InitDebug(pprofBind []string) {
	pprofServeMux := http.NewServeMux()
	pprofServeMux.HandleFunc("/debug/pprof/", pprof.Index)
	pprofServeMux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	pprofServeMux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	pprofServeMux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	for _, addr := range pprofBind {
		go func() {
			if err := http.ListenAndServe(addr, pprofServeMux); err != nil {
				fmt.Errorf("http.ListenAndServe(\"%s\", pprofServeMux) error(%v)", addr, err)
				panic(err)
			}
		}()
	}
}
