package test

import (
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

// TEST LOGIC WRITE

type TxPool struct {
	From     string `json:"from"`
	To       string `json:"to"`
	Gas      string `json:"gas"`
	GasPrice string `json:"gas_price"`
	Value    string `json:"value"`
	Data     string `json:"data"`
	Nonce    string `json:"nonce"`
}

func TestInsertTx(t *testing.T) {
	err := os.MkdirAll("./store", 0755)
	if err != nil {
		fmt.Println("Error creating directory:", err)
		return
	}

	tx1 := new(TxPool)
	tx1 = &TxPool{
		From:     "haha",
		To:       "hihi",
		Gas:      "hihi",
		GasPrice: "hih",
		Value:    "sds",
		Data:     "ss",
		Nonce:    "ss",
	}

	file, err := os.OpenFile("store/transactions.txt", os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	assert.Nil(t, err)

	marshal, err := json.Marshal(tx1)
	assert.Nil(t, err)

	_, err = file.WriteString(string(marshal))
	assert.Nil(t, err)

}
