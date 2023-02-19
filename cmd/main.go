package main

import (
	"ewallet/internal/storage/sqlite"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"os"
	"sync"
	"time"
)

const sqliteStoragePath = "data/sqlite/storage.db"

type walletServer struct {
	sync.Mutex
	storage *sqlite.Storage
}

func main() {
	router := mux.NewRouter()
	router.StrictSlash(true)
	server := NewWalletServer()

	log.Print("service started")

	router.HandleFunc("/api/send", server.Send).Methods("POST")
	router.HandleFunc("/api/transactions", server.GetLast).Methods("GET")
	router.HandleFunc("/api/wallet/{address}/balance", server.GetBalance).Methods("GET")

	// Set up logging and panic recovery middleware.
	file, err := os.OpenFile("logs/all.logs", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0640)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	router.Use(func(h http.Handler) http.Handler {
		return handlers.LoggingHandler(file, h)
	})
	router.Use(handlers.RecoveryHandler(handlers.PrintRecoveryStack(true)))

	// Timeout middleware.
	muxWithTimeout := http.TimeoutHandler(router, time.Second*10, "Timeout!")

	log.Fatal(http.ListenAndServe("localhost:"+os.Getenv("SERVERPORT"), muxWithTimeout))
}
