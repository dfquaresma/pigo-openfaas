// Copyright (c) Alex Ellis 2017. All rights reserved.
// Copyright (c) OpenFaaS Author(s) 2018. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package main

import (
	"handler/function"
	"net/http"
	"os"
)

func main() {
	http.HandleFunc("/__gci", gci)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		body, err := function.Handle(*r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Write(body)
	})
	if err := http.ListenAndServe(":"+os.Getenv("HANDLER_PORT"), nil); err != nil {
		panic(err)
	}
}
