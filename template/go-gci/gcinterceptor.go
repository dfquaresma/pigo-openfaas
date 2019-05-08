package main

import (
	"net/http"
	"runtime"
	"runtime/debug"
	"strconv"
)

func init() {
	debug.SetGCPercent(-1) // Disabling automatic garbage collection.
}

func gci(w http.ResponseWriter, r *http.Request) {
	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)
	st, err := strconv.ParseUint(r.Header.Get("gci"), 10, 64)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if mem.HeapAlloc > st {
		runtime.GC()
		w.Write([]byte{'1'})
		return
	}
	w.Write([]byte{'0'})
}
