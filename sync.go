package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

func SyncToDynamics(cfg Config, po PurchaseOrder) error {
	payload := map[string]interface{}{
		"PurchaseOrderNumber": po.ID,
		"VendorAccountNumber": po.VendorID,
		"TotalAmount":         po.Amount,
		"Currency":            po.Currency,
	}

	jsonPayload, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", cfg.Dynamics365.APIURL, bytes.NewBuffer(jsonPayload))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("dynamics API Error: %s", resp.Status)
	}

	return nil
}
