package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

func ReportErrorToGlitchTip(cfg Config, poID string, err error) {
	payload := map[string]string{
		"title":   "Purchase Order Sync Failed",
		"message": fmt.Sprintf("Failed to sync PO: %s, Error: %v", poID, err),
	}
	jsonPayload, _ := json.Marshal(payload)

	http.Post(cfg.GlitchTip.APIURL, "application/json", bytes.NewBuffer(jsonPayload))
}
