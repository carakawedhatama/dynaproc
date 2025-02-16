package main

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSyncToDynamics(t *testing.T) {
	// Test case 1: Successful sync
	t.Run("Successfully sync purchase order", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "POST", r.Method)
			assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

			body, err := io.ReadAll(r.Body)
			assert.NoError(t, err)

			var payload map[string]interface{}
			err = json.Unmarshal(body, &payload)
			assert.NoError(t, err)

			assert.Equal(t, "PO123", payload["PurchaseOrderNumber"])
			assert.Equal(t, "V001", payload["VendorAccountNumber"])
			assert.Equal(t, 100.50, payload["TotalAmount"])
			assert.Equal(t, "USD", payload["Currency"])

			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		cfg := Config{
			Dynamics365: Dynamics365Config{
				APIURL: server.URL,
			},
		}

		po := PurchaseOrder{
			ID:       "PO123",
			VendorID: "V001",
			Amount:   100.50,
			Currency: "USD",
		}

		err := SyncToDynamics(cfg, po)
		assert.NoError(t, err)
	})

	// Test case 2: Server error response
	t.Run("Server returns error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Internal Server Error"))
		}))
		defer server.Close()

		cfg := Config{
			Dynamics365: Dynamics365Config{
				APIURL: server.URL,
			},
		}

		po := PurchaseOrder{
			ID:       "PO123",
			VendorID: "V001",
			Amount:   100.50,
			Currency: "USD",
		}

		err := SyncToDynamics(cfg, po)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "500")
	})

	// Test case 3: Network error
	t.Run("Network error", func(t *testing.T) {
		cfg := Config{
			Dynamics365: Dynamics365Config{
				APIURL: "http://invalid-url-that-does-not-exist.com",
			},
		}

		po := PurchaseOrder{
			ID:       "PO123",
			VendorID: "V001",
			Amount:   100.50,
			Currency: "USD",
		}

		err := SyncToDynamics(cfg, po)
		assert.Error(t, err)
	})

	// Test case 4: Invalid JSON response
	t.Run("Server returns invalid JSON", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("invalid json"))
		}))
		defer server.Close()

		cfg := Config{
			Dynamics365: Dynamics365Config{
				APIURL: server.URL,
			},
		}

		po := PurchaseOrder{
			ID:       "PO123",
			VendorID: "V001",
			Amount:   100.50,
			Currency: "USD",
		}

		err := SyncToDynamics(cfg, po)
		assert.NoError(t, err)
	})

	// Test case 5: Empty API URL
	t.Run("Empty API URL", func(t *testing.T) {
		cfg := Config{
			Dynamics365: Dynamics365Config{
				APIURL: "",
			},
		}

		po := PurchaseOrder{
			ID:       "PO123",
			VendorID: "V001",
			Amount:   100.50,
			Currency: "USD",
		}

		err := SyncToDynamics(cfg, po)
		assert.Error(t, err)
	})

	// Test case 6: Malformed URL
	t.Run("Malformed URL", func(t *testing.T) {
		cfg := Config{
			Dynamics365: Dynamics365Config{
				APIURL: "not-a-valid-url",
			},
		}

		po := PurchaseOrder{
			ID:       "PO123",
			VendorID: "V001",
			Amount:   100.50,
			Currency: "USD",
		}

		err := SyncToDynamics(cfg, po)
		assert.Error(t, err)
	})
}
