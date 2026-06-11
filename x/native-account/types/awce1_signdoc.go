package types

import "fmt"

type CosmosSignDoc struct {
	AccountNumber uint64     `json:"account_number"`
	ChainID       string     `json:"chain_id"`
	Fee           CosmosFee  `json:"fee"`
	Memo          string     `json:"memo"`
	Msgs          []CosmosMsg `json:"msgs"`
	Sequence      uint64     `json:"sequence"`
}

type CosmosFee struct {
	Amount []CosmosCoin `json:"amount"`
	Gas    string       `json:"gas"`
}

type CosmosCoin struct {
	Denom  string `json:"denom"`
	Amount string `json:"amount"`
}

type CosmosMsg struct {
	TypeURL string `json:"type_url"`
	Value   string `json:"value"`
}

func NewCosmosSignDoc(accountNumber, sequence uint64, chainID string, fee CosmosFee, msgs []CosmosMsg, memo string) *CosmosSignDoc {
	return &CosmosSignDoc{
		AccountNumber: accountNumber,
		ChainID:       chainID,
		Fee:           fee,
		Memo:          memo,
		Msgs:          msgs,
		Sequence:      sequence,
	}
}

func (d *CosmosSignDoc) SignBytes() []byte {
	return []byte(fmt.Sprintf(`{"account_number":"%d","chain_id":"%s","sequence":"%d"}`,
		d.AccountNumber, d.ChainID, d.Sequence))
}

func ValidateCosmosSignDoc(doc *CosmosSignDoc) error {
	if doc.ChainID == "" {
		return fmt.Errorf("AWCE1: SignDoc chain_id must not be empty")
	}
	if len(doc.Msgs) == 0 {
		return fmt.Errorf("AWCE1: SignDoc must have at least one message")
	}
	for i, msg := range doc.Msgs {
		if msg.TypeURL == "" {
			return fmt.Errorf("AWCE1: SignDoc msg[%d] has empty type_url", i)
		}
	}
	return nil
}
