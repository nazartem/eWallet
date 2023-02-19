package sqlite

import (
	"context"
	"database/sql"
	"ewallet/internal/model"
	"fmt"
	"math/rand"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type Storage struct {
	db *sql.DB
}

// createAddress creates a 64 digit wallet address.
func createAddress() string {
	rand.Seed(time.Now().UnixNano())
	chars := []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZ" +
		"abcdefghijklmnopqrstuvwxyz" +
		"0123456789")
	length := 64
	var b strings.Builder
	for i := 0; i < length; i++ {
		b.WriteRune(chars[rand.Intn(len(chars))])
	}
	str := b.String()

	return str
}

// New creates new SQLite storage.
func New(path string) (*Storage, error) {
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, fmt.Errorf("can't open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("can't connect to database: %w", err)
	}

	return &Storage{db: db}, nil
}

// SaveWallet saves wallet to storage.
func (s *Storage) SaveWallet(ctx context.Context, w *model.Wallet) error {
	q := `INSERT INTO wallets (address, balance) VALUES (?, ?)`

	if _, err := s.db.ExecContext(ctx, q, w.Address, w.Balance); err != nil {
		return fmt.Errorf("can't save wallet: %w", err)
	}

	return nil
}

// SaveTransaction saves transaction to storage.
func (s *Storage) SaveTransaction(ctx context.Context, t *model.Transaction) error {
	q := `INSERT INTO transactions (amount, creation_time, sender, recipient) VALUES (?, ?, ?, ?)`

	if _, err := s.db.ExecContext(ctx, q, t.Amount, t.CreatedAt, t.SenderAddress, t.DestinationAddress); err != nil {
		return fmt.Errorf("can't save wallet: %w", err)
	}

	return nil
}

// GetLastTransactions returns N last transactions from the database.
func (s *Storage) GetLastTransactions(ctx context.Context, n int) ([]model.Transaction, error) {
	q := `SELECT * FROM transactions ORDER BY creation_time DESC LIMIT ?`

	rows, err := s.db.QueryContext(ctx, q, n)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("no saved transactions")
	}
	if err != nil {
		return nil, fmt.Errorf("can't get last transations: %w", err)
	}

	var transactions []model.Transaction

	for rows.Next() {
		t := model.Transaction{}
		err := rows.Scan(&t.Amount, &t.CreatedAt, &t.SenderAddress, &t.DestinationAddress)
		if err != nil {
			return nil, fmt.Errorf("can't scan attributes into transation: %w", err)
		}
		transactions = append(transactions, t)
	}

	return transactions, nil
}

// SendMoney transfers funds from one wallet to another.
func (s *Storage) SendMoney(ctx context.Context, sender, receiver string, amount float64) error {
	q := `UPDATE wallets SET balance = (balance - ?) WHERE address = ?`

	if _, err := s.db.ExecContext(ctx, q, amount, sender); err != nil {
		return fmt.Errorf("can't send money from %s : %w", sender, err)
	}

	q = `UPDATE wallets SET balance = (balance + ?) WHERE address = ?`

	if _, err := s.db.ExecContext(ctx, q, amount, receiver); err != nil {
		return fmt.Errorf("can't accept money on %s : %w", receiver, err)
	}

	return nil
}

// GetBalance returns wallet balance.
func (s *Storage) GetBalance(ctx context.Context, address string) (float64, error) {
	q := `SELECT balance FROM wallets WHERE address = ?`

	var balance float64

	err := s.db.QueryRowContext(ctx, q, address).Scan(&balance)
	if err == sql.ErrNoRows {
		return 0, fmt.Errorf("unable to find sender's wallet")
	}
	if err != nil {
		return 0, fmt.Errorf("can't get balance: %w", err)
	}

	return balance, nil
}

// RemoveWallet removes wallets from storage.
func (s *Storage) RemoveWallet(ctx context.Context, address string) error {
	q := `DELETE FROM wallets WHERE address = ?`
	if _, err := s.db.ExecContext(ctx, q, address); err != nil {
		return fmt.Errorf("can't remove wallet: %w", err)
	}

	return nil
}

// RemoveTransaction removes transaction from storage.
func (s *Storage) RemoveTransaction(ctx context.Context, t *model.Transaction) error {
	q := `DELETE FROM transactions WHERE sender = ? AND creation_time = ? AND recipient = ?`
	if _, err := s.db.ExecContext(ctx, q, t.SenderAddress, t.CreatedAt, t.DestinationAddress); err != nil {
		return fmt.Errorf("can't remove transaction: %w", err)
	}

	return nil
}

// IsExistsWallet checks if wallet exists in storage.
func (s *Storage) IsExistsWallet(ctx context.Context, address string) (bool, error) {
	q := `SELECT COUNT(*) FROM wallets WHERE address = ?`

	var count int

	if err := s.db.QueryRowContext(ctx, q, address).Scan(&count); err != nil {
		return false, fmt.Errorf("can't check if wallet exists: %w", err)
	}

	return count > 0, nil
}

func (s *Storage) Init(ctx context.Context) error {
	q1 := `CREATE TABLE IF NOT EXISTS wallets (address TEXT, balance REAL);`
	q2 := `CREATE TABLE IF NOT EXISTS transactions (amount REAL, creation_time TEXT, sender TEXT, recipient TEXT)`

	_, err := s.db.ExecContext(ctx, q1)
	if err != nil {
		return fmt.Errorf("can't create table wallets: %w", err)
	}

	q := `SELECT COUNT(*) FROM wallets`

	var count int

	if err := s.db.QueryRowContext(ctx, q).Scan(&count); err != nil {
		return fmt.Errorf("can't check wallet's count: %w", err)
	}

	if count == 0 {
		for i := 0; i < 10; i++ {
			w := &model.Wallet{
				Address: createAddress(),
				Balance: 100,
			}

			s.SaveWallet(ctx, w)
		}
	}

	_, err = s.db.ExecContext(ctx, q2)
	if err != nil {
		return fmt.Errorf("can't create table transactions: %w", err)
	}

	return nil
}
