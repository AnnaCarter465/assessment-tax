package database

import (
	"context"
	"database/sql"

	_ "github.com/lib/pq"
)

type DB struct {
	sqlDB *sql.DB
}

func NewDB(dbURL string) (*DB, error) {
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		return nil, err
	}

	return &DB{sqlDB: db}, nil
}

func (db *DB) GetSQLDB() *sql.DB {
	return db.sqlDB
}

func (db *DB) FindAllDefaultAllowances(ctx context.Context) ([]DefaultAllowance, error) {
	var results []DefaultAllowance

	rows, err := db.GetSQLDB().QueryContext(
		ctx,
		`
			SELECT allowance_type, amount FROM default_allowances
		`)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var (
			allowanceType string
			amount        float64
		)

		err = rows.Scan(&allowanceType, &amount)
		if err != nil {
			return nil, err
		}

		results = append(results, DefaultAllowance{
			AllowanceType: allowanceType,
			Amount:        amount,
		})
	}

	return results, nil
}

func (db *DB) UpdateAmountDefaultAllowances(ctx context.Context, allowanceType string, amount float64) (DefaultAllowance, error) {
	var (
		at string
		am float64
	)

	err := db.GetSQLDB().QueryRowContext(ctx,
		`
			UPDATE default_allowances
			SET amount = $2
			WHERE allowance_type = $1
			RETURNING allowance_type, amount
	   	`, allowanceType, amount).Scan(&at, &am)
	if err != nil {
		return DefaultAllowance{}, err
	}

	return DefaultAllowance{
		AllowanceType: at,
		Amount:        am,
	}, nil
}

func (db *DB) FindAllAllowedAllowances(ctx context.Context) ([]AllowedAllowance, error) {
	var results []AllowedAllowance

	rows, err := db.GetSQLDB().QueryContext(
		ctx,
		`
		SELECT allowance_type, max_amount FROM allowed_allowances
		`)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var (
			allowanceType string
			maxAmount     float64
		)

		err = rows.Scan(&allowanceType, &maxAmount)
		if err != nil {
			return nil, err
		}

		results = append(results, AllowedAllowance{
			AllowanceType: allowanceType,
			MaxAmount:     maxAmount,
		})
	}

	return results, nil
}

type DefaultAllowance struct {
	AllowanceType string  `db:"allowance_type"`
	Amount        float64 `db:"amount"`
}

type AllowedAllowance struct {
	AllowanceType string  `db:"allowance_type"`
	MaxAmount     float64 `db:"max_amount"`
}
