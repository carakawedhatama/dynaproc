package main

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestReportErrorToGlitchTip(t *testing.T) {
	// Test case 1: Successful error report with full validation
	t.Run("Successfully report error with full validation", func(t *testing.T) {
		var receivedPayload map[string]string
		requestReceived := false

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestReceived = true

			assert.Equal(t, "POST", r.Method)
			assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

			body, err := io.ReadAll(r.Body)
			assert.NoError(t, err)
			err = json.Unmarshal(body, &receivedPayload)
			assert.NoError(t, err)

			w.WriteHeader(http.StatusOK)
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"status":"ok"}`))
		}))
		defer server.Close()

		cfg := Config{
			GlitchTip: GlitchTipConfig{
				APIURL: server.URL,
			},
		}

		testError := errors.New("test sync error")
		testPOID := "PO123"

		ReportErrorToGlitchTip(cfg, testPOID, testError)

		assert.True(t, requestReceived, "Request was not received by the server")

		expectedTitle := "Purchase Order Sync Failed"
		expectedMessage := "Failed to sync PO: PO123, Error: test sync error"

		assert.Equal(t, expectedTitle, receivedPayload["title"])
		assert.Equal(t, expectedMessage, receivedPayload["message"])
	})

	// Test case 2: Error report with empty PO ID
	t.Run("Report error with empty PO ID", func(t *testing.T) {
		var receivedPayload map[string]string

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			body, _ := io.ReadAll(r.Body)
			json.Unmarshal(body, &receivedPayload)
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		cfg := Config{
			GlitchTip: GlitchTipConfig{
				APIURL: server.URL,
			},
		}

		testError := errors.New("test error")
		ReportErrorToGlitchTip(cfg, "", testError)

		expectedMessage := "Failed to sync PO: , Error: test error"
		assert.Equal(t, expectedMessage, receivedPayload["message"])
	})

	// Test case 3: Error report with nil error
	t.Run("Report with nil error", func(t *testing.T) {
		var receivedPayload map[string]string

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			body, _ := io.ReadAll(r.Body)
			json.Unmarshal(body, &receivedPayload)
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		cfg := Config{
			GlitchTip: GlitchTipConfig{
				APIURL: server.URL,
			},
		}

		ReportErrorToGlitchTip(cfg, "PO123", nil)

		expectedMessage := "Failed to sync PO: PO123, Error: <nil>"
		assert.Equal(t, expectedMessage, receivedPayload["message"])
	})

	// Test case 4: Server returns error
	t.Run("Server returns error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"error":"internal server error"}`))
		}))
		defer server.Close()

		cfg := Config{
			GlitchTip: GlitchTipConfig{
				APIURL: server.URL,
			},
		}

		ReportErrorToGlitchTip(cfg, "PO123", errors.New("test error"))
	})

	// Test case 5: Network timeout
	t.Run("Network timeout", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			time.Sleep(100 * time.Millisecond)
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		cfg := Config{
			GlitchTip: GlitchTipConfig{
				APIURL: server.URL,
			},
		}

		ReportErrorToGlitchTip(cfg, "PO123", errors.New("test error"))
	})

	// Test case 6: Invalid URL
	t.Run("Invalid URL", func(t *testing.T) {
		cfg := Config{
			GlitchTip: GlitchTipConfig{
				APIURL: "not-a-valid-url",
			},
		}

		ReportErrorToGlitchTip(cfg, "PO123", errors.New("test error"))
	})

	// Test case 7: Empty URL
	t.Run("Empty URL", func(t *testing.T) {
		cfg := Config{
			GlitchTip: GlitchTipConfig{
				APIURL: "",
			},
		}

		ReportErrorToGlitchTip(cfg, "PO123", errors.New("test error"))
	})
}
