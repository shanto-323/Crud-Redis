package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	_ "github.com/lib/pq"
	"github.com/redis/go-redis/v9"
)

type Database interface {
	CreateAccount(*User) error
	GetAccounts() ([]*User, error)
}

type Store struct {
	Db    *sql.DB
	Redis *redis.Client
}

func NewStore() (*Store, error) {
	defaultSrt := "user=postgres password=1234 host=localhost port=5432 sslmode=disable"
	dbName := "my_users"

	defaultDB, err := sql.Open("postgres", defaultSrt)
	if err != nil {
		return nil, err
	}
	defer defaultDB.Close()

	var exists bool
	err = defaultDB.QueryRow("SELECT EXISTS (SELECT 1 FROM pg_database WHERE datname = $1)", dbName).Scan(&exists)
	if err != nil {
		return nil, err
	}

	if !exists {
		_, err = defaultDB.Exec(fmt.Sprintf("CREATE DATABASE %s", dbName))
		if err != nil {
			return nil, err
		}
	}

	connStr := fmt.Sprintf("user=postgres dbname=%s password=1234 sslmode=disable", dbName)
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}
	if err = db.Ping(); err != nil {
		return nil, err
	}

	// Redis client
	redis := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "1234",
		DB:       0,
	})

	return &Store{Db: db, Redis: redis}, nil
}

func (p *Store) createTable() error {
	quary := `  
    CREATE TABLE IF NOT EXISTS users (
      id SERIAL PRIMARY KEY,
      name VARCHAR(30)
    )`
	_, err := p.Db.Exec(quary)
	return err
}

func (p *Store) init() error {
	return p.createTable()
}

func (p *Store) CreateAccount(acc *User) error {
	quary := `  
    INSERT INTO users (name)
    VALUES($1)
  `
	_, err := p.Db.Query(quary, acc.Name)
	if err != nil {
		return err
	}
	ctx := context.Background()
	p.Redis.Del(ctx, "list")

	return nil
}

func (p *Store) GetAccounts() ([]*User, error) {
	ctx := context.Background()
	accounts := []*User{}

	//Try get data from redis
	cacheKey := "list"
	cacheData, err := p.Redis.Get(ctx, cacheKey).Result()
	if err == nil {
		err := json.Unmarshal([]byte(cacheData), &accounts)
		if err == nil {
			fmt.Println("data from redis")
			return accounts, nil
		}
	}

	rows, err := p.Db.Query("SELECT * FROM users")
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		account, err := getAccount(rows)
		if err != nil {
			return nil, err
		}
		accounts = append(accounts, account)
	}

	//Set data in resis
	jsonData, err := json.Marshal(accounts)
	if err == nil {
		p.Redis.Set(ctx, cacheKey, jsonData, 0)
	}
	return accounts, nil
}

func getAccount(rows *sql.Rows) (*User, error) {
	account := &User{}
	err := rows.Scan(
		&account.Id,
		&account.Name)
	if err != nil {
		return nil, err
	}
	return account, err
}
