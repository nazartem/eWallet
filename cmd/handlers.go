package main

import (
	"context"
	"encoding/json"
	"ewallet/internal/model"
	"ewallet/internal/storage/sqlite"
	"fmt"
	"log"
	"mime"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
)

func NewWalletServer() *walletServer {
	storage, err := sqlite.New(sqliteStoragePath)
	if err != nil {
		log.Fatal("can't connect to storage: ", err)
	}

	if err := storage.Init(context.TODO()); err != nil {
		log.Fatal("can't init storage: ", err)
	}

	return &walletServer{storage: storage}
}

// renderJSON renders 'v' as JSON and writes it as a response into w.
func renderJSON(w http.ResponseWriter, v interface{}) {
	js, err := json.Marshal(v)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}

// Send sends funds from one of the wallets to the specified wallet.
// The method accepts in the request body a JSON object containing the fields: from, then, amount.
func (ws *walletServer) Send(w http.ResponseWriter, req *http.Request) {
	ws.Lock()
	defer ws.Unlock()

	log.Printf("handling Send at %s\n", req.URL.Path)

	type Request struct {
		From   string  `json:"from"`
		To     string  `json:"to"`
		Amount float64 `json:"amount"`
	}

	// Enforce a JSON Content-Type.
	contentType := req.Header.Get("Content-Type")
	mediatype, _, err := mime.ParseMediaType(contentType)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if mediatype != "application/json" {
		http.Error(w, "expect application/json Content-Type", http.StatusUnsupportedMediaType)
		return
	}

	dec := json.NewDecoder(req.Body)
	dec.DisallowUnknownFields()
	var re Request
	if err := dec.Decode(&re); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if re.From == re.To {
		http.Error(w, "'from' and 'to' adresses must be different", http.StatusBadRequest)
		return
	}

	if re.Amount <= 0 {
		http.Error(w, "amount must be greater than 0", http.StatusBadRequest)
		return
	}

	ok, err := ws.storage.IsExistsWallet(context.TODO(), re.From)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if !ok {
		http.Error(w, "sender's wallet does not found", http.StatusNotFound)
		return
	}

	ok, err = ws.storage.IsExistsWallet(context.TODO(), re.To)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if !ok {
		http.Error(w, "receiver's wallet does not found", http.StatusNotFound)
		return
	}

	balance, err := ws.storage.GetBalance(context.TODO(), re.From)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	if balance < re.Amount {
		http.Error(w, "insufficient funds", http.StatusNotAcceptable)
		return
	}

	err = ws.storage.SendMoney(context.TODO(), re.From, re.To, re.Amount)
	if err != nil {
		http.Error(w, "can't make send", http.StatusInternalServerError)
	}

	transaction := &model.Transaction{
		Amount:             re.Amount,
		CreatedAt:          fmt.Sprint(time.Now()),
		SenderAddress:      re.From,
		DestinationAddress: re.To,
	}

	err = ws.storage.SaveTransaction(context.TODO(), transaction)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// GetLast returns information about the N most recent funds transfers. The method accepts numeric
// JSON objects as query parameters, returned in an array.
func (ws *walletServer) GetLast(w http.ResponseWriter, req *http.Request) {
	ws.Lock()
	defer ws.Unlock()

	log.Printf("handling get last transactions at %s\n", req.URL.Path)

	var lastTransaction []model.Transaction

	n, err := strconv.Atoi(req.URL.Query().Get("count"))
	if err != nil {
		http.Error(w, "expected count=N in reqests GET parametres (/api/transactions?count=N)", http.StatusBadRequest)
		return
	}

	lastTransaction, err = ws.storage.GetLastTransactions(context.TODO(), n)

	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	renderJSON(w, lastTransaction)
}

// GetBalance returns information about the balance of the wallet in a JSON object.
// The wallet address is specified in the request path.
func (ws *walletServer) GetBalance(w http.ResponseWriter, req *http.Request) {
	ws.Lock()
	defer ws.Unlock()

	log.Printf("handling get balance at %s\n", req.URL.Path)

	address, _ := mux.Vars(req)["address"]
	balance, err := ws.storage.GetBalance(context.TODO(), address)

	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	renderJSON(w, balance)
}
