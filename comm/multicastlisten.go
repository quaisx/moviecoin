package comm

import (
	"log"
	"net"
)

const (
	maxDatagramSize = 8192
)

func MulticastListen(address string, handler func(*net.UDPAddr, int, []byte)) {
	addr, err := net.ResolveUDPAddr("udp4", address)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("ListenMulticastUDP %s", addr)
	conn, err := net.ListenMulticastUDP("udp4", nil, addr)
	if err != nil {
		log.Fatal(err)
	}

	conn.SetReadBuffer(maxDatagramSize)

	for {
		buffer := make([]byte, maxDatagramSize)
		numBytes, src, err := conn.ReadFromUDP(buffer)
		if err != nil {
			log.Fatal("ReadFromUDP failed:", err)
		}
		log.Printf("Received from %s %d bytes", src, numBytes)
		handler(src, numBytes, buffer)
	}
}
