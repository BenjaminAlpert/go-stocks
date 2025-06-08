package server

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
)

var (
	plotErr  error
	plotData bytes.Buffer
	mutex    sync.Mutex
)

func New() {
	http.HandleFunc("/avg_rate.svg", func(w http.ResponseWriter, r *http.Request) {
		mutex.Lock()
		if plotErr != nil {
			w.Header().Set("Content-Type", "text/plain")
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "plotting error: %s", plotErr)
		} else {
			w.Header().Set("Content-Type", "image/svg+xml")
			w.WriteHeader(http.StatusOK)
			w.Write(plotData.Bytes())
			fmt.Println("[INFO] Served: /avg_rate.svg")
		}
		mutex.Unlock()
	})

	log.Fatal(http.ListenAndServe(":8080", nil))
}

func UpdateWriter(writerTo io.WriterTo, pErr error) {
	mutex.Lock()

	plotData = bytes.Buffer{}
	writer := bufio.NewWriter(&plotData)
	writerTo.WriteTo(writer)

	plotErr = pErr

	mutex.Unlock()
}
