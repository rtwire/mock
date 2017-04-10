package service

import "net/http"

type feePayload struct {
	FeePerByte  int64 `json:"feePerByte"`
	BlockHeight int64 `json:"blockHeight"`
}

func (c *chain) getFeesHandler(w http.ResponseWriter, r *http.Request) {

	if !acceptHeaderFound(w, r) {
		return
	}

	fees := c.Fees()

	payload := make([]feePayload, len(fees))
	for i, fee := range fees {
		payload[i] = feePayload{
			FeePerByte:  fee.feePerByte,
			BlockHeight: fee.blockHeight,
		}
	}

	sendPayload(w, http.StatusOK, "fees", "", payload)
}
