package main

import (
	"flag"
	"log"
)

func init() {
	log.SetPrefix("<WalletServer>: ")
}

func main() {
	port := flag.Uint("port", 8080, "TCP Port Number for Wallet Server")
	gateway := flag.String("gateway", "http://127.0.0.1", "Blockchain Gateway")
	gateway_port := flag.Uint("gateway_port", 5000, "Blockchain Gateway Port")

	flag.Parse()

	app := NewWalletServer(uint16(*port), *gateway, uint16(*gateway_port))
	app.Run()
}
