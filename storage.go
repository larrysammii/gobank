package main

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
	// the underscore is used to import a package for its side effects only
)

type Storage interface {
	CreateAccount(*Account) error
	DeleteAccount(int) error
	UpdateAccount(*Account) error
	GetAccounts() ([]*Account, error)
	GetAccountByID(int) (*Account, error)
}

type PostgresStore struct {
	db *sql.DB
}

// Instanciate a postgres container:
// ```
// docker run --name some-postgres -e POSTGRES_PASSWORD=gobank -p 5432:5432 -d postgres
// ```
// map port 5432 to 5432 or whatever is available.
//
// Connect to postgres container to use psql bash commands:
// ```
// docker exec -it some-postgres bash
// psql -h localhost -U postgres
// ```
//
// I had error when I run the server: "pq: role "postgres" does not exist",
// turns out there was postgres running on my machine, and probably clashes with the port,
// if you installed postgres via homebrew, kill it as below:
// ```
// brew services stop postgres
// ```
// or just map it to another port idk.

// To check connection:
// ```
// docker ps
// telnet localhost {whatever the port you used, eg. 5432}
// ```

func NewPostgresStore() (*PostgresStore, error) {
	connStr := "user=postgres dbname=postgres password=gobank sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		return nil, err
	}

	return &PostgresStore{db: db}, nil
}

func (s *PostgresStore) Init() error {
	return s.createAccountTable()
}

func (s *PostgresStore) createAccountTable() error {
	query := `CREATE TABLE IF NOT EXISTS account (
	id SERIAL PRIMARY KEY,
	first_name VARCHAR(50),
    last_name VARCHAR(50),
    number SERIAL,
    balance SERIAL,
    created_at TIMESTAMP
	)`

	_, err := s.db.Exec(query)
	return err
}

func (s *PostgresStore) CreateAccount(acc *Account) error {
	query := `
	INSERT INTO account (first_name, last_name, number, balance, created_at)
	VALUES ($1, $2, $3, $4, $5)
    `
	resp, err := s.db.Query(
		query,
		acc.FirstName,
		acc.LastName,
		acc.Number,
		acc.Balance,
		acc.CreatedAt,
	)

	if err != nil {
		return err
	}

	fmt.Printf("%+v\n", resp)

	return nil
}

func (s *PostgresStore) UpdateAccount(*Account) error {
	return nil
}

func (s *PostgresStore) DeleteAccount(id int) error {
	return nil
}

func (s *PostgresStore) GetAccountByID(id int) (*Account, error) {
	return nil, nil
}

func (s *PostgresStore) GetAccounts() ([]*Account, error) {
	rows, err := s.db.Query("SELECT * FROM account")
	if err != nil {
		return nil, err
	}

	accounts := []*Account{}

	for rows.Next() {
		account := new(Account)
		err = rows.Scan(
			&account.ID,
			&account.FirstName,
			&account.LastName,
			&account.Number,
			&account.Balance,
			&account.CreatedAt,
		)

		if err != nil {
			return nil, err
		}

		accounts = append(accounts, account)

	}
	return accounts, nil
}
