package utils

import (
	"fmt"
	"log"
	"moviecoin/comm"
	"net"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"
)

func IsFoundHost(host string) bool {
	_, err := net.DialTimeout("tcp", host, 1*time.Second)
	if err != nil {
		fmt.Printf("%s %v\n", host, err)
		return false
	}
	return true
}

var PATTERN = regexp.MustCompile(`((25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?\.){3})(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)`)

const (
	defaultMulticastAddress = "239.0.0.0:9999"
	liveNodeHeader          = "<LIVE NODE>"
)

type NeighborCache struct {
	mux   sync.Mutex
	cache map[string]bool
}

var neighborCache *NeighborCache = &NeighborCache{cache: make(map[string]bool, 0)}

func (n *NeighborCache) Update(address string) {
	n.mux.Lock()
	defer n.mux.Unlock()
	_, ok := n.cache[address]
	if !ok {
		n.cache[address] = true
	}
}

func (n *NeighborCache) Reset() {
	n.mux.Lock()
	defer n.mux.Unlock()
	n.cache = make(map[string]bool)
}

func MulticastHandler(src *net.UDPAddr, n int, b []byte) {
	msg := string(b[:n])
	log.Printf("Multicast message received: %s - %s [%d bytes]", src, msg, n)
	tokens := strings.Split(msg, "|")
	if tokens[0] == liveNodeHeader {
		// @TODO: check for ip address correctness
		neighborCache.Update(tokens[1])
	}
	//log.Println(hex.Dump(b[:n]))
}

func RunListener() {
	comm.MulticastListen(defaultMulticastAddress, MulticastHandler)
}

func ListenNeighbors() {
	go RunListener()
}

func (n *NeighborCache) GetNeighborsFromCache() []string {
	n.mux.Lock()
	defer n.mux.Unlock()

	neighbors := make([]string, 0)
	for key := range n.cache {
		neighbors = append(neighbors, key)
	}
	return neighbors
}

func NotifyNeighbors(myHost string, myPort uint16) {
	address := fmt.Sprintf("%s:%d", myHost, myPort)
	conn, err := comm.NewMulticast(defaultMulticastAddress)
	if err != nil {
		log.Fatal(err)
	}
	message := fmt.Sprintf("%s|%s", liveNodeHeader, address)
	// Multicast presence 3 times
	for i := 0; i < 3; i++ {
		log.Printf(" -> Notify neighbors about my presence: %s", address)
		conn.Write([]byte(message))
	}
	conn.Close()
}

func FindNeighbors() []string {
	neighborsFromcCache := neighborCache.GetNeighborsFromCache()
	neighbors := make([]string, 0)
	if len(neighbors) > 0 {
		for _, addr := range neighborsFromcCache {
			if IsFoundHost(addr) {
				neighbors = append(neighbors, addr)
			}
		}
	}
	return neighbors
}

func GetHost() string {
	hostname, err := os.Hostname()
	if err != nil {
		return "127.0.0.1"
	}
	address, err := net.LookupHost(hostname)
	if err != nil {
		return "127.0.0.1"
	}
	return address[0]
}
