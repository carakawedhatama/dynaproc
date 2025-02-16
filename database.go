package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
)

var db *sql.DB

func InitDB(cfg Config) {
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.Name,
		cfg.Database.SSLMode,
	)
	var err error
	db, err = sql.Open("postgres", dsn)
	if err != nil {
		log.Fatal(err)
	}
}

func FetchPendingOrders() ([]PurchaseOrder, error) {
	rows, err := db.Query("SELECT id, vendor_id, amount, currency FROM purchase_orders WHERE status = 'APPROVED' AND synced = FALSE")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []PurchaseOrder
	for rows.Next() {
		var po PurchaseOrder
		err := rows.Scan(&po.ID, &po.VendorID, &po.Amount, &po.Currency)
		if err != nil {
			return nil, err
		}
		orders = append(orders, po)
	}
	return orders, nil
}
