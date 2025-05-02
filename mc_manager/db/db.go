package db

import (
	"database/sql"
	"embed"
	"fmt"
	"io/fs"
	"mc_manager/logging"

	_ "github.com/mattn/go-sqlite3"
	"github.com/pressly/goose/v3"
)

//go:embed sql/*
var sql_files embed.FS

type DB struct {
	Connection      *sql.DB
	files           *embed.FS
	statement_cache map[string]*sql.Stmt
}

func NewDatabase() (*DB, error) {
	return InitializeDatabase("main.db")
}

func InitializeDatabase(file_name string) (*DB, error) {
	connection, err := sql.Open("sqlite3", file_name)
	if err != nil {
		return nil, err
	}

	embed_migrations, err := fs.Sub(sql_files, "sql")
	if err != nil {
		return nil, err
	}
	goose.SetBaseFS(embed_migrations)

	if err := goose.SetDialect("sqlite3"); err != nil {
		return nil, err
	}

	logging.Debug("Running migrations")
	if err := goose.Up(connection, "migrations"); err != nil {
		return nil, err
	}

	db := &DB{
		Connection:      connection,
		files:           &sql_files,
		statement_cache: make(map[string]*sql.Stmt),
	}

	return db, nil
}

func (db *DB) GetStmt(query_name string) (*sql.Stmt, error) {
	if stmt, ok := db.statement_cache[query_name]; ok {
		return stmt, nil
	}

	filePath := fmt.Sprintf("sql/queries/%s.sql", query_name)
	query, err := db.files.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	stmt, err := db.Connection.Prepare(string(query))
	if err != nil {
		return nil, err
	}

	db.statement_cache[query_name] = stmt
	return stmt, nil
}

func (db *DB) NewTx() (*Tx, error) {
	tx, err := db.Connection.Begin()
	if err != nil {
		return nil, err
	}

	return &Tx{
		Tx:              tx,
		files:           db.files,
		statement_cache: make(map[string]*sql.Stmt),
	}, nil
}

type Tx struct {
	Tx              *sql.Tx
	files           *embed.FS
	statement_cache map[string]*sql.Stmt
}

func (tx *Tx) GetStmt(query_name string) (*sql.Stmt, error) {
	if stmt, ok := tx.statement_cache[query_name]; ok {
		return stmt, nil
	}

	filePath := fmt.Sprintf("sql/queries/%s.sql", query_name)
	query, err := tx.files.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	stmt, err := tx.Tx.Prepare(string(query))
	if err != nil {
		return nil, err
	}

	tx.statement_cache[query_name] = stmt
	return stmt, nil
}
func (tx *Tx) Commit() error {
	if err := tx.Tx.Commit(); err != nil {
		return err
	}
	tx.Tx = nil
	return nil
}
func (tx *Tx) Rollback() error {
	if err := tx.Tx.Rollback(); err != nil {
		return err
	}
	tx.Tx = nil
	return nil
}
func (tx *Tx) Close() error {
	if tx.Tx != nil {
		if err := tx.Tx.Rollback(); err != nil {
			return err
		}
	}
	tx.Tx = nil
	return nil
}
