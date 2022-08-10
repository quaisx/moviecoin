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

var (
	local_addresses = []string{"localhost", "127.0.0.1", "::1"}
)

func init() {
	log.SetPrefix("<WalletServer>: ")
}

func isLocalAddr(addr string) bool {
	for _, v := range local_addresses {
		if v == addr {
			return true
		}
	}
	return false
}

func main() {
	port := flag.Uint("port", 8080, "TCP Port Number for Wallet Server")
	node := flag.String("node", "localhost", "Blockchain Node")
	node_port := flag.Uint("node_port", 5000, "Blockchain Node Port")

	flag.Parse()

	node_addr := "http://"
	if !isLocalAddr(*node) {
		addrs, err := net.LookupHost(*node)
		if err != nil {
			log.Fatalf(fmt.Sprintf("Invalid blockchain node address: %s", err))
			os.Exit(ERROR_EXIT_CODE)
		}

		for _, addr := range addrs {
			ip := net.ParseIP(addr)
			ipv4 := ip.To4()
			if ipv4 == nil {
				continue
			}
			node_addr += addr
			break
		}
	} else {
		node_addr += *node
	}
	app := NewWalletServer(uint16(*port), node_addr, uint16(*node_port))
	app.Run()
}
