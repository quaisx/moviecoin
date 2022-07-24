package blockchain

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"moviecoin/utils"
	"net/http"
	"strings"
	"sync"
	"time"
)

// @TODO - temp values that need to be regrouped
// MINING_* stuff will be set by the blockchain network dynamically
// blockchain network will need to go through a discovery mechanism to
// discover new nodes and add them automatically to a cache
const (
	// @TODO - group 1
	MINING_DIFFICULTY = 3
	MINING_SENDER     = "MOVIECOIN BLOCKCHAIN"
	MINING_REWARD     = 1.0
	MINING_TIMER_SEC  = 30 // default mining time lapse
	// @TODO - group 2
	BLOCKCHAIN_NEIGHBOR_SYNC_TIME_SEC = 20
)

type Blockchain struct {
	transactionPool   []*Transaction
	chain             []*Block
	blockchainAddress string
	port              uint16
	mux               sync.Mutex
	neighbors         []string
	muxNeighbors      sync.Mutex
}

func NewBlockchain(blockchainAddress string, port uint16) *Blockchain {
	bc := new(Blockchain)
	bc.blockchainAddress = blockchainAddress
	//create genesis block
	bc.CreateBlock(0, new(Block).Hash()) //<- hash of all zeros block
	bc.port = port
	return bc
}

func (bc *Blockchain) CreateBlock(nonce int, previousHash [32]byte) *Block {
	b := NewBlock(nonce, previousHash, bc.transactionPool)
	bc.chain = append(bc.chain, b)
	// when a new block is created, the transaction pool is reset!
	bc.transactionPool = []*Transaction{}
	// When a new block is mined, tell all other nodes to delete transaction pools
	for _, n := range bc.neighbors {
		endpoint := fmt.Sprintf("http://%s/transactions", n)
		// new http client connection to talk to each node separately
		client := &http.Client{}
		// send out a delete request to delete transaction pool
		req, _ := http.NewRequest("DELETE", endpoint, nil)
		resp, _ := client.Do(req)
		log.Printf("%v", resp)
	}
	return b
}

func (bc *Blockchain) Chain() []*Block {
	return bc.chain
}

func (bc *Blockchain) Run() {

	bc.MulticastPresence()
	bc.ListenNeighbors()
	bc.StartNotifyNeighbors()
	bc.StartSyncNeighbors()
	bc.ResolveConflicts()
	bc.StartMining()
}

// Periodically notify other nodes of us being alive
func (bc *Blockchain) StartNotifyNeighbors() {
	bc.NotifyNeighbors()
	_ = time.AfterFunc(time.Second*BLOCKCHAIN_NEIGHBOR_SYNC_TIME_SEC, bc.StartNotifyNeighbors)
}

// keep looking for new nodes at BLOCKCHAIN_NEIGHBOR_SYNC_TIME_SEC intervals
// Note: the master node should communicate this interval value instead of hard coding it
func (bc *Blockchain) StartSyncNeighbors() {
	bc.SyncNeighbors() //let's discover mining hosts only once
	_ = time.AfterFunc(time.Second*BLOCKCHAIN_NEIGHBOR_SYNC_TIME_SEC, bc.StartSyncNeighbors)
}

//Resolve mining conflicts
func (bc *Blockchain) ResolveConflicts() bool {
	var longestChain []*Block = nil
	maxLength := len(bc.chain)
	// go through neighbors and ask for the chains
	for _, n := range bc.neighbors {
		endpoint := fmt.Sprintf("http://%s/chain", n)
		resp, _ := http.Get(endpoint)
		if resp.StatusCode == 200 {
			var bcResp Blockchain
			decoder := json.NewDecoder(resp.Body)
			// unmarshal the neighbor's blockchain
			_ = decoder.Decode(&bcResp)
			// get their's blockchain
			chain := bcResp.Chain()
			// if their chain is longer - use their chain instead
			if len(chain) > maxLength && bc.ValidChain(chain) {
				maxLength = len(chain)
				longestChain = chain
			}
		}
	}

	if longestChain != nil {
		bc.chain = longestChain
		log.Printf("Conflict resolved: adopt a new blockchain")
		return true
	}
	log.Printf("Conflict resolved: keep my blockchain")
	return false
}

func (bc *Blockchain) NotifyNeighbors() {
	utils.NotifyNeighbors(utils.GetHost(), bc.port)
}

func (bc *Blockchain) SetNeighbors() {
	// I am a blockchain running on a host with port x
	// discover and set all other nodes
	bc.neighbors = utils.FindNeighbors()
	log.Printf("%v", bc.neighbors)
}

func (bc *Blockchain) ListenNeighbors() {
	log.Println("Listening for other neighbors...")
	utils.ListenNeighbors()
}

func (bc *Blockchain) AnouncePresence() {
	utils.NotifyNeighbors(utils.GetHost(), bc.port)
}

func (bc *Blockchain) MulticastPresence() {
	bc.muxNeighbors.Lock()
	defer bc.muxNeighbors.Unlock()
	bc.AnouncePresence()
}

func (bc *Blockchain) SyncNeighbors() {
	bc.muxNeighbors.Lock()
	defer bc.muxNeighbors.Unlock()
	bc.SetNeighbors()
}

func (bc *Blockchain) TransactionPool() []*Transaction {
	return bc.transactionPool
}

func (bc *Blockchain) ClearTransactionPool() {
	bc.transactionPool = bc.transactionPool[:0]
}

func (bc *Blockchain) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Blocks []*Block `json:"chain"`
	}{
		Blocks: bc.chain,
	})
}

func (bc *Blockchain) UnmarshalJSON(data []byte) error {
	v := &struct {
		Blocks *[]*Block `json:"chain"`
	}{
		Blocks: &bc.chain,
	}
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}
	return nil
}

func (bc *Blockchain) LastBlock() *Block {
	return bc.chain[len(bc.chain)-1]
}

func (bc *Blockchain) String() string {
	var output string
	for i, block := range bc.chain {
		output += fmt.Sprintf("%s Chain %d %s\n", strings.Repeat("=", 25), i,
			strings.Repeat("=", 25))
		output += fmt.Sprintf("%v", block)
	}
	output += fmt.Sprintf("%s\n", strings.Repeat("*", 25))
	return output
}

// New transaction
func (bc *Blockchain) CreateTransaction(sender string, receiver string, amount float32,
	senderPublicKey *ecdsa.PublicKey, s *utils.Signature) bool {
	isTransacted := bc.AddTransaction(sender, receiver, amount, senderPublicKey, s)

	if isTransacted {
		for _, n := range bc.neighbors {
			publicKeyStr := fmt.Sprintf("%064x%064x", senderPublicKey.X.Bytes(),
				senderPublicKey.Y.Bytes())
			signatureStr := s.String()
			bt := &TransactionRequest{
				&sender, &receiver, &publicKeyStr, &amount, &signatureStr}
			m, _ := json.Marshal(bt)
			buf := bytes.NewBuffer(m)
			endpoint := fmt.Sprintf("http://%s/transactions", n)
			client := &http.Client{}
			req, _ := http.NewRequest("PUT", endpoint, buf)
			resp, _ := client.Do(req)
			log.Printf("%v", resp)
		}
	}

	return isTransacted
}

func (bc *Blockchain) AddTransaction(sender string, receiver string, amount float32,
	senderPublicKey *ecdsa.PublicKey, s *utils.Signature) bool {
	t := NewTransaction(sender, receiver, amount)

	if sender == MINING_SENDER {
		bc.transactionPool = append(bc.transactionPool, t)
		return true
	}

	if bc.VerifyTransactionSignature(senderPublicKey, s, t) {
		if bc.CalculateTotalAmount(sender) < amount {
			log.Println("ERROR: Insufficient funds")
			return false
		}
		bc.transactionPool = append(bc.transactionPool, t)
		return true
	} else {
		log.Println("ERROR: Invalid transaction")
	}
	return false

}

func (bc *Blockchain) VerifyTransactionSignature(
	senderPublicKey *ecdsa.PublicKey, s *utils.Signature, t *Transaction) bool {
	m, _ := json.Marshal(t)
	h := sha256.Sum256([]byte(m))
	return ecdsa.Verify(senderPublicKey, h[:], s.R, s.S)
}

func (bc *Blockchain) CopyTransactionPool() []*Transaction {
	transactions := make([]*Transaction, 0)
	for _, t := range bc.transactionPool {
		transactions = append(transactions,
			NewTransaction(t.sender,
				t.receiver,
				t.amount))
	}
	return transactions
}

func (bc *Blockchain) ValidProof(nonce int, previousHash [32]byte, transactions []*Transaction, difficulty int) bool {
	zeros := strings.Repeat("0", difficulty)
	guessBlock := Block{0, nonce, previousHash, transactions}
	guessHashStr := fmt.Sprintf("%x", guessBlock.Hash())
	return guessHashStr[:difficulty] == zeros
}

func (bc *Blockchain) ProofOfWork() int {
	transactions := bc.CopyTransactionPool()
	previousHash := bc.LastBlock().Hash()
	nonce := 0
	for !bc.ValidProof(nonce, previousHash, transactions, MINING_DIFFICULTY) {
		nonce += 1
	}
	return nonce
}

func (bc *Blockchain) Mining() bool {
	bc.mux.Lock()
	defer bc.mux.Unlock()
	// Transaction pool must contain transactions in order to mine
	if len(bc.transactionPool) > 0 {
		// add a reward transaction to the pool
		bc.AddTransaction(MINING_SENDER, bc.blockchainAddress, MINING_REWARD, nil, nil)
		nonce := bc.ProofOfWork()
		previousHash := bc.LastBlock().Hash()
		// POW done, mint a new block
		bc.CreateBlock(nonce, previousHash)
		log.Println("Mining is done. New block created.")
		// run consensus across all mining nodes
		for _, n := range bc.neighbors {
			endpoint := fmt.Sprintf("http://%s/consensus", n)
			client := &http.Client{}
			req, _ := http.NewRequest("PUT", endpoint, nil)
			resp, _ := client.Do(req)
			log.Printf("%v", resp)
		}
		return true
	}
	log.Println(":( Nothing to mine ... Taking a nap (zzz.zz.z)")
	return false
}

func (bc *Blockchain) StartMining() {
	bc.Mining()
	// Mininig at MINING_TIMER_SEC interval
	_ = time.AfterFunc(time.Second*MINING_TIMER_SEC, bc.StartMining)
}

func (bc *Blockchain) CalculateTotalAmount(blockchainAddress string) float32 {
	var totalAmount float32 = 0.0
	if blockchainAddress == MINING_SENDER {
		//for now, let's assume it's infinite supply of coins
		totalAmount = float32(math.MaxFloat32)
		return totalAmount
	}
	for _, b := range bc.chain {
		for _, t := range b.transactions {
			amount := t.amount
			// credit the receiver
			if blockchainAddress == t.receiver {
				totalAmount += amount
			}
			// debit the sender
			if blockchainAddress == t.sender {
				totalAmount -= amount
			}
		}
	}
	return totalAmount
}

func (bc *Blockchain) ValidChain(chain []*Block) bool {
	preBlock := chain[0]
	currentIndex := 1
	for currentIndex < len(chain) {
		b := chain[currentIndex]
		if b.previousHash != preBlock.Hash() {
			return false
		}
		// Note: mining difficulty may vary across different blocks in a blockchain.
		// @TODO add logic to handle validating POW with a variable mining difficulty factor
		if !bc.ValidProof(b.Nonce(), b.PreviousHash(), b.Transactions(), MINING_DIFFICULTY) {
			return false
		}
		preBlock = b
		currentIndex += 1
	}
	return true
}
