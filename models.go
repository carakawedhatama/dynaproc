package main

type PurchaseOrder struct {
	ID       string  `json:"id"`
	VendorID string  `json:"vendor_id"`
	Amount   float64 `json:"amount"`
	Currency string  `json:"currency"`
}
