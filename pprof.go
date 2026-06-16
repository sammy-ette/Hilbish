//go:build pprof

package main

import (
	"net/http"
	_ "net/http/pprof"
)

func init() {
	go func() {
		http.ListenAndServe("localhost:8080", nil)
	}()
}
