//go:build pprof

package main

import (
    _ "net/http/pprof"
    "net/http"
)

func init() {
    go func() {
        http.ListenAndServe("localhost:8080", nil)
    }()
}
