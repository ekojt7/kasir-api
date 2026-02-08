package repositories

import (
	"database/sql"
	"fmt"
	"kasir-api/models"
)

type TransactionRepository struct {
	db *sql.DB
}

func NewTransactionRepository(db *sql.DB) *TransactionRepository {
	return &TransactionRepository{db: db}
}

func (repo *TransactionRepository) CreateTransaction(items []models.CheckoutItem) (*models.Transaction, error) {
	var (
		res *models.Transaction
	)

	tx, err := repo.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// inisialisasi subtotal -> jumlah total transaksi keseluruhan
	totalAmount := 0
	// inisialisasi modeling transactionDetails -> nanti kita insert ke db
	details := make([]models.TransactionDetail, 0)
	// loop setiap item
	for _, item := range items {
		var productName string
		var productID, price, stock int
		// get product dapet pricing
		err := tx.QueryRow("SELECT id, name, price, stock FROM products WHERE id=$1", item.ProductID).Scan(&productID, &productName, &price, &stock)
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("product id %d not found", item.ProductID)
		}

		if err != nil {
			return nil, err
		}

		// hitung current total = quantity * pricing
		// ditambahin ke dalam subtotal
		subtotal := item.Quantity * price
		totalAmount += subtotal

		// kurangi jumlah stok
		_, err = tx.Exec("UPDATE products SET stock = stock - $1 WHERE id = $2", item.Quantity, productID)
		if err != nil {
			return nil, err
		}

		// item nya dimasukkin ke transactionDetails
		details = append(details, models.TransactionDetail{
			ProductID:   productID,
			ProductName: productName,
			Quantity:    item.Quantity,
			Subtotal:    subtotal,
		})
	}

	// insert transaction
	var transactionID int
	err = tx.QueryRow("INSERT INTO transactions (total_amount) VALUES ($1) RETURNING ID", totalAmount).Scan(&transactionID)
	if err != nil {
		return nil, err
	}

	// insert transaction details
	for i := range details {
		details[i].TransactionID = transactionID
		_, err = tx.Exec("INSERT INTO transaction_details (transaction_id, product_id, quantity, subtotal) VALUES ($1, $2, $3, $4)",
			transactionID, details[i].ProductID, details[i].Quantity, details[i].Subtotal)
		if err != nil {
			return nil, err
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	res = &models.Transaction{
		ID:          transactionID,
		TotalAmount: totalAmount,
		Details:     details,
	}

	return res, nil
}

// GetTodayReport mendapatkan laporan penjualan hari ini
func (repo *TransactionRepository) GetTodayReport() (*models.SalesReport, error) {
	var report models.SalesReport

	// Query untuk mendapatkan total revenue dan total transaksi hari ini
	query := `
		SELECT 
			COALESCE(SUM(total_amount), 0) as total_revenue,
			COUNT(*) as total_transaksi
		FROM transactions
		WHERE DATE(created_at) = CURRENT_DATE
	`

	err := repo.db.QueryRow(query).Scan(&report.TotalRevenue, &report.TotalTransaksi)
	if err != nil {
		return nil, err
	}

	// Query untuk mendapatkan produk terlaris hari ini
	topProductQuery := `
		SELECT 
			p.name,
			COALESCE(SUM(td.quantity), 0) as qty_terjual
		FROM transaction_details td
		JOIN transactions t ON td.transaction_id = t.id
		JOIN products p ON td.product_id = p.id
		WHERE DATE(t.created_at) = CURRENT_DATE
		GROUP BY p.id, p.name
		ORDER BY qty_terjual DESC
		LIMIT 1
	`

	var topProduct models.TopProduct
	err = repo.db.QueryRow(topProductQuery).Scan(&topProduct.Nama, &topProduct.QtyTerjual)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}

	report.ProdukTerlaris = topProduct

	return &report, nil
}

// GetReportByDateRange mendapatkan laporan penjualan berdasarkan range tanggal
func (repo *TransactionRepository) GetReportByDateRange(startDate, endDate string) (*models.SalesReport, error) {
	var report models.SalesReport

	// Query untuk mendapatkan total revenue dan total transaksi dalam range tanggal
	query := `
		SELECT 
			COALESCE(SUM(total_amount), 0) as total_revenue,
			COUNT(*) as total_transaksi
		FROM transactions
		WHERE DATE(created_at) >= $1 AND DATE(created_at) <= $2
	`

	err := repo.db.QueryRow(query, startDate, endDate).Scan(&report.TotalRevenue, &report.TotalTransaksi)
	if err != nil {
		return nil, err
	}

	// Query untuk mendapatkan produk terlaris dalam range tanggal
	topProductQuery := `
		SELECT 
			p.name,
			COALESCE(SUM(td.quantity), 0) as qty_terjual
		FROM transaction_details td
		JOIN transactions t ON td.transaction_id = t.id
		JOIN products p ON td.product_id = p.id
		WHERE DATE(t.created_at) >= $1 AND DATE(t.created_at) <= $2
		GROUP BY p.id, p.name
		ORDER BY qty_terjual DESC
		LIMIT 1
	`

	var topProduct models.TopProduct
	err = repo.db.QueryRow(topProductQuery, startDate, endDate).Scan(&topProduct.Nama, &topProduct.QtyTerjual)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}

	report.ProdukTerlaris = topProduct

	return &report, nil
}