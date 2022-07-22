package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"moviecoin/blockchain"
	"moviecoin/utils"
	"net/http"
	"path"
	"regexp"
	"strconv"
	"strings"

	"moviecoin/wallet"
)

const tempDir = "./templates"

type WalletServer struct {
	port    uint16
	gateway string
}

func NewWalletServer(port uint16, gateway string, gateway_port uint16) *WalletServer {
	return &WalletServer{port, gateway + fmt.Sprintf(":%d", gateway_port)}
}

func (ws *WalletServer) Port() uint16 {
	return ws.port
}

func (ws *WalletServer) Gateway() string {
	return ws.gateway
}

func (ws *WalletServer) Index(w http.ResponseWriter, req *http.Request) {
	url := strings.Split(req.RequestURI, " ")
	log.Printf("[%s]%s", req.Method, url[0])
	switch req.Method {
	case http.MethodGet:
		t, err := template.ParseFiles(path.Join(tempDir, "index.html"))
		if err != nil {
			log.Printf("ERROR processing template: %s", err)
		}
		t.Execute(w, "")
	default:
		log.Printf("ERROR: Invalid HTTP Method")
	}
}

func (ws *WalletServer) Wallet(w http.ResponseWriter, req *http.Request) {
	url := strings.Split(req.RequestURI, " ")
	log.Printf("[%s]%s", req.Method, url[0])
	switch req.Method {
	case http.MethodPost:
		w.Header().Add("Content-Type", "application/json")
		myWallet := wallet.NewWallet()
		m, _ := myWallet.MarshalJSON()
		io.WriteString(w, string(m[:]))
	default:
		w.WriteHeader(http.StatusBadRequest)
		log.Println("ERROR: Invalid HTTP Method")
	}
}

func (ws *WalletServer) CreateTransaction(w http.ResponseWriter, req *http.Request) {
	url := strings.Split(req.RequestURI, " ")
	log.Printf("[%s]%s", req.Method, url[0])
	switch req.Method {
	case http.MethodPost:
		decoder := json.NewDecoder(req.Body)
		var t wallet.TransactionRequest
		err := decoder.Decode(&t)
		if err != nil {
			log.Printf("ERROR: %v", err)
			io.WriteString(w, string(utils.JsonStatus("fail")))
			return
		}
		if !t.Validate() {
			log.Println("ERROR: missing field(s)")
			io.WriteString(w, string(utils.JsonStatus("fail")))
			return
		}

		publicKey := utils.PublicKeyFromString(*t.SenderPublicKey)
		privateKey := utils.PrivateKeyFromString(*t.SenderPrivateKey, publicKey)
		value, err := strconv.ParseFloat(*t.Amount, 32)
		if err != nil {
			log.Println("ERROR: parse error")
			io.WriteString(w, string(utils.JsonStatus("fail")))
			return
		}
		value32 := float32(value)

		w.Header().Add("Content-Type", "application/json")

		transaction := wallet.NewTransaction(
			privateKey,
			publicKey,
			*t.SenderAddress,
			*t.ReceiverAddress,
			value32)
		signature := transaction.GenerateSignature()
		signatureStr := signature.String()

		bt := &blockchain.TransactionRequest{
			SenderAddress:   t.SenderAddress,
			ReceiverAddress: t.ReceiverAddress,
			SenderPublicKey: t.SenderPublicKey,
			Amount:          &value32,
			Signature:       &signatureStr,
		}
		m, _ := json.Marshal(bt)
		buf := bytes.NewBuffer(m)

		resp, _ := http.Post(ws.Gateway()+"/transactions", "application/json", buf)
		if resp.StatusCode == 201 {
			io.WriteString(w, string(utils.JsonStatus("success")))
			return
		}
		io.WriteString(w, string(utils.JsonStatus("fail")))
	default:
		w.WriteHeader(http.StatusBadRequest)
		log.Println("ERROR: Invalid HTTP Method")
	}
}

func (ws *WalletServer) WalletAmount(w http.ResponseWriter, req *http.Request) {
	url := strings.Split(req.RequestURI, " ")
	log.Printf("[%s]%s", req.Method, url[0])
	switch req.Method {
	case http.MethodGet:
		blockchainAddress := req.URL.Query().Get("wallet_address")
		endpoint := fmt.Sprintf("%s/amount", ws.Gateway())

		client := &http.Client{}
		bcsReq, _ := http.NewRequest("GET", endpoint, nil)
		q := bcsReq.URL.Query()
		q.Add("blockchain_address", blockchainAddress)
		bcsReq.URL.RawQuery = q.Encode()

		bcsResp, err := client.Do(bcsReq)
		if err != nil {
			log.Printf("ERROR: %v", err)
			io.WriteString(w, string(utils.JsonStatus("fail")))
			return
		}

		w.Header().Add("Content-Type", "application/json")
		if bcsResp.StatusCode == 200 {
			decoder := json.NewDecoder(bcsResp.Body)
			var bar blockchain.AmountResponse
			err := decoder.Decode(&bar)
			if err != nil {
				log.Printf("ERROR: %v", err)
				io.WriteString(w, string(utils.JsonStatus("fail")))
				return
			}

			m, _ := json.Marshal(struct {
				Message string  `json:"message"`
				Amount  float32 `json:"amount"`
			}{
				Message: "success",
				Amount:  bar.Amount,
			})
			io.WriteString(w, string(m[:]))
		} else {
			io.WriteString(w, string(utils.JsonStatus("fail")))
		}
	default:
		log.Printf("ERROR: Invalid HTTP Method")
		w.WriteHeader(http.StatusBadRequest)
	}
}

func (ws *WalletServer) Auth(handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		url := strings.Split(req.RequestURI, " ")
		log.Printf("Auth: [%s]%s Authenticated successfully", req.Method, url[0])
		handler(w, req)
	}
}

func (ws *WalletServer) AssetServe(w http.ResponseWriter, req *http.Request) {
	url := strings.Split(req.RequestURI, " ")
	log.Printf("AssetServe: [%s]%s", req.Method, url[0])

	hd := http.StripPrefix("/templates/", http.FileServer(http.Dir("./templates")))

	re_ico, _ := regexp.Compile(`\.ico`)
	found := re_ico.Find([]byte(url[0]))
	if found != nil {
		w.Header().Set("Content-Type", "image/x-icon")
	}

	re_css, _ := regexp.Compile(`\.css`)
	found = re_css.Find([]byte(url[0]))
	if found != nil {
		w.Header().Set("Content-Type", "text/css")
	}
	hd.ServeHTTP(w, req)
}

func (ws *WalletServer) Run() {
	http.HandleFunc("/", ws.Index)
	http.HandleFunc("/wallet", ws.Wallet)
	http.HandleFunc("/wallet/amount", ws.WalletAmount)
	http.HandleFunc("/transaction", ws.CreateTransaction)
	http.HandleFunc("/templates/", ws.AssetServe)
	log.Fatal(http.ListenAndServe("0.0.0.0:"+strconv.Itoa(int(ws.Port())), nil))
}
