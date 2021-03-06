package internal

import (
	"testing"

	"github.com/kinecosystem/go/clients/horizon"
	"github.com/stretchr/testify/assert"

	"sync"
)

func TestFriendbot_Pay(t *testing.T) {
	mockSubmitTransaction := func(bot *Bot, channel chan TxResult, signed string) {
		txSuccess := horizon.TransactionSuccess{Env: signed}
		// we don't want to actually submit the tx here but emulate a success instead
		channel <- TxResult{
			maybeTransactionSuccess: &txSuccess,
			maybeErr:                nil,
		}
	}

	fb := &Bot{
		Secret:            "SAQWC7EPIYF3XGILYVJM4LVAVSLZKT27CTEI3AFBHU2VRCMQ3P3INPG5",
		Network:           "Test SDF Network ; September 2015",
		StartingBalance:   "10000",
		SubmitTransaction: mockSubmitTransaction,
		sequence:          2,
	}

	txSuccess, err := fb.Pay("GDJIN6W6PLTPKLLM57UW65ZH4BITUXUMYQHIMAZFYXF45PZVAWDBI77Z", fb.StartingBalance, false)
	if !assert.NoError(t, err) {
		return
	}
	expectedTxn := "AAAAAPuYf7x7KGvFX9fjCR9WIaoTX3yHJYwX6ZSx6w76HPjEAAAAZAAAAAAAAAADAAAAAAAAAAAAAAABAAAAAAAAAAAAAAAA" +
		"0ob63nrm9S1s7+lvdyfgUTpejMQOhgMlxcvOvzUFhhQAAAAAO5rKAAAAAAAAAAAB+hz4xAAAAECzNV2yXevMYKzm7OhXX2gYwmLZ5V37yeRHUX" +
		"3Vhb6eT8wkUtpj2vJsUwzLWjdKMyGonFCPkaG4twRFUVqBRLEH"
	assert.Equal(t, expectedTxn, txSuccess.Env)

	txSuccess, err = fb.Pay("GDJIN6W6PLTPKLLM57UW65ZH4BITUXUMYQHIMAZFYXF45PZVAWDBI77Z", fb.StartingBalance, true)
	if !assert.NoError(t, err) {
		return
	}
	expectedTxn = "AAAAAPuYf7x7KGvFX9fjCR9WIaoTX3yHJYwX6ZSx6w76HPjEAAAAZAAAAAAAAAAEAAAAAAAAAAAAAAABAAAAAAAAAAEAAAAA0o" +
		"b63nrm9S1s7+lvdyfgUTpejMQOhgMlxcvOvzUFhhQAAAAAAAAAADuaygAAAAAAAAAAAfoc+MQAAABAYOKOlEGCt+A6DGPi4GXbRWtQnyNSUn" +
		"V/hE4vcfAe0Z6z/d0U4BoqlwUx6/bPHq5+VMaKysU8xeJJ2GZMxWhNCg=="
	assert.Equal(t, expectedTxn, txSuccess.Env)

	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		_, err := fb.Pay("GDJIN6W6PLTPKLLM57UW65ZH4BITUXUMYQHIMAZFYXF45PZVAWDBI77Z", fb.StartingBalance, false)
		// don't assert on the txn value here because the ordering is not guaranteed between these 2 goroutines
		assert.NoError(t, err)
		wg.Done()
	}()
	go func() {
		_, err := fb.Pay("GDJIN6W6PLTPKLLM57UW65ZH4BITUXUMYQHIMAZFYXF45PZVAWDBI77Z", fb.StartingBalance, false)
		// don't assert on the txn value here because the ordering is not guaranteed between these 2 goroutines
		assert.NoError(t, err)
		wg.Done()
	}()
	wg.Wait()
}
