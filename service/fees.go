package service

import "net/http"

type feePayload struct {
	FeePerByte  int64 `json:"feePerByte"`
	BlockHeight int64 `json:"blockHeight"`
}

func (s *service) getFeesHandler(w http.ResponseWriter, r *http.Request) {

	if !acceptHeaderFound(w, r) {
		return
	}

	fees := s.Fees()

	payload := make([]feePayload, len(fees))
	for i, fee := range fees {
		payload[i] = feePayload{
			FeePerByte:  fee.feePerByte,
			BlockHeight: fee.blockHeight,
		}
	}

	sendPayload(w, http.StatusCreated, "fees", "", payload)
}
