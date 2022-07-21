package blockchain

import "encoding/json"

type AmountResponse struct {
	Amount float32 `json:"amount"`
}

func (ar *AmountResponse) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Amount float32 `json:"amount"`
	}{
		Amount: ar.Amount,
	})
}

func (ar *AmountResponse) UnmarshalJSON(data []byte) error {
	v := &struct {
		Amount *float32 `json:"amount"`
	}{
		Amount: &ar.Amount,
	}
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}
	return nil
}
