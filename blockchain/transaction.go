package blockchain

import (
	"encoding/json"
	"fmt"
	"strings"
)

type Transaction struct {
	sender   string
	receiver string
	amount   float32
}

type TransactionRequest struct {
	SenderAddress   *string  `json:"sender_address"`
	ReceiverAddress *string  `json:"receiver_address"`
	SenderPublicKey *string  `json:"sender_public_key"`
	Amount          *float32 `json:"amount"`
	Signature       *string  `json:"signature"`
}

func NewTransaction(sender string, recipient string, value float32) *Transaction {
	return &Transaction{sender, recipient, value}
}

func (t *Transaction) String() string {
	output := fmt.Sprintf("%s\n", strings.Repeat("-", 40))
	output += fmt.Sprintf(" sender_address     %s\n", t.sender)
	output += fmt.Sprintf(" receiver_address   %s\n", t.receiver)
	output += fmt.Sprintf(" amount             %.1f\n", t.amount)
	return output
}

func (t *Transaction) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		SenderAddress   string  `json:"sender_address"`
		ReceiverAddress string  `json:"recipient_address"`
		Amount          float32 `json:"amount"`
	}{
		SenderAddress:   t.sender,
		ReceiverAddress: t.receiver,
		Amount:          t.amount,
	})
}

func (t *Transaction) UnmarshalJSON(data []byte) error {
	v := &struct {
		SenderAddress   *string  `json:"sender_address"`
		ReceiverAddress *string  `json:"recipient_address"`
		Amount          *float32 `json:"amount"`
	}{
		SenderAddress:   &t.sender,
		ReceiverAddress: &t.receiver,
		Amount:          &t.amount,
	}
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}
	return nil
}

func (tr *TransactionRequest) Validate() bool {
	if tr.SenderAddress == nil ||
		tr.ReceiverAddress == nil ||
		tr.SenderPublicKey == nil ||
		tr.Amount == nil ||
		tr.Signature == nil {
		return false
	}
	return true
}
