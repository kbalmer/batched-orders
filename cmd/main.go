package main

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gorilla/mux"

	"batched-orders/batcher"
	"batched-orders/server"
	"batched-orders/sftp"
)

func main() {
	// env vars
	batchSizeEnvVar := os.Getenv("BATCH_SIZE")
	if batchSizeEnvVar == "" {
		batchSizeEnvVar = "100"
	}

	batchSize, err := strconv.Atoi(batchSizeEnvVar)
	if err != nil {
		fmt.Printf("failed to parse env var for batch size")
		return
	}

	ordersDirectory := os.Getenv("ORDERS_DIRECTORY")
	if ordersDirectory == "" {
		fmt.Printf("failed to parse env var for orders directory")
		return
	}

	// init
	sftp := sftp.NewSFTPClient()

	downloadTicker := time.NewTicker(time.Minute * 5)

	b, doneCh := batcher.NewBatcher(batchSize, downloadTicker, sftp)

	go b.BatchListener()
	go b.Send()
	go b.GetLatest()

	h := server.NewHandler(b)

	// http serve and listen
	r := mux.NewRouter()
	r.HandleFunc("/orders", h.ProcessOrder).Methods(http.MethodPost)
	server := &http.Server{
		Handler:      r,
		Addr:         "0.0.0.0:8000",
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  30 * time.Second,
	}

	err = server.ListenAndServe()
	if err != nil {
		if errors.Is(err, http.ErrServerClosed) {
			fmt.Println("Server closed.")
		} else {
			fmt.Printf("Server err: %s\n", err)
		}
	}

	close(doneCh)
}
