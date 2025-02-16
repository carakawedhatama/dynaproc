package main

import (
	"database/sql"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func TestFetchPendingOrders(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock database: %v", err)
	}
	defer mockDB.Close()

	oldDB := db
	db = mockDB
	defer func() { db = oldDB }()

	// Test case 1: Fetch pending orders successfully
	t.Run("Fetch pending orders successfully", func(t *testing.T) {

		rows := sqlmock.NewRows([]string{"id", "vendor_id", "amount", "currency"}).
			AddRow("PO001", "V001", 100.50, "USD").
			AddRow("PO002", "V002", 200.75, "EUR")

		mock.ExpectQuery("SELECT id, vendor_id, amount, currency FROM purchase_orders WHERE status = 'APPROVED' AND synced = FALSE").
			WillReturnRows(rows)

		orders, err := FetchPendingOrders()
		assert.NoError(t, err)
		assert.Len(t, orders, 2)

		assert.Equal(t, "PO001", orders[0].ID)
		assert.Equal(t, "V001", orders[0].VendorID)
		assert.Equal(t, 100.50, orders[0].Amount)
		assert.Equal(t, "USD", orders[0].Currency)

		assert.Equal(t, "PO002", orders[1].ID)
		assert.Equal(t, "V002", orders[1].VendorID)
		assert.Equal(t, 200.75, orders[1].Amount)
		assert.Equal(t, "EUR", orders[1].Currency)

		assert.NoError(t, mock.ExpectationsWereMet())
	})

	// Test case 2: No pending orders
	t.Run("No pending orders", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "vendor_id", "amount", "currency"})

		mock.ExpectQuery("SELECT id, vendor_id, amount, currency FROM purchase_orders WHERE status = 'APPROVED' AND synced = FALSE").
			WillReturnRows(rows)

		orders, err := FetchPendingOrders()
		assert.NoError(t, err)
		assert.Empty(t, orders)

		assert.NoError(t, mock.ExpectationsWereMet())
	})

	// Test case 3: Database error
	t.Run("Database error", func(t *testing.T) {
		mock.ExpectQuery("SELECT id, vendor_id, amount, currency FROM purchase_orders WHERE status = 'APPROVED' AND synced = FALSE").
			WillReturnError(sql.ErrConnDone)

		orders, err := FetchPendingOrders()
		assert.Error(t, err)
		assert.Nil(t, orders)

		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestInitDB(t *testing.T) {
	t.Run("Database connection string format", func(t *testing.T) {
		cfg := Config{
			Database: DatabaseConfig{
				Host:     "testhost",
				Port:     5432,
				User:     "testuser",
				Password: "testpass",
				Name:     "testdb",
				SSLMode:  "disable",
			},
		}

		mockDB, _, err := sqlmock.New()
		if err != nil {
			t.Fatalf("Failed to create mock database: %v", err)
		}
		defer mockDB.Close()

		oldDB := db
		db = mockDB
		defer func() { db = oldDB }()

		InitDB(cfg)
		assert.NotNil(t, db)
	})
}
