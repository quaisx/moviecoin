package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"
)

const (
	ERROR_EXIT_CODE = 1
)

func init() {
	log.SetPrefix("<WalletServer>: ")
}

func main() {
	port := flag.Uint("port", 8080, "TCP Port Number for Wallet Server")
	gateway := flag.String("gateway", "localhost", "Blockchain Gateway")
	gateway_port := flag.Uint("gateway_port", 5000, "Blockchain Gateway Port")

	flag.Parse()

	addrs, err := net.LookupHost(*gateway)

	if err != nil {
		log.Fatal(fmt.Sprintf("Invalid gateway address: %s", err))
		os.Exit(ERROR_EXIT_CODE)
	}

	for _, addr := range addrs {
		ip := net.ParseIP(addr)
		ipv4 := ip.To4()
		if ipv4 == nil {
			continue
		}
		*gateway = addr
		break
	}

	app := NewWalletServer(uint16(*port), *gateway, uint16(*gateway_port))
	app.Run()
}
