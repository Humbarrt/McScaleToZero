package db

import (
	"database/sql"
	"embed"
	"fmt"
	"io/fs"
	"mc_analyser/logging"

	_ "github.com/mattn/go-sqlite3"
	"github.com/pressly/goose/v3"
)

//go:embed sql/*
var sql_files embed.FS

type DB struct {
	Connection      *sql.DB
	files           embed.FS
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
		files:           sql_files,
		statement_cache: make(map[string]*sql.Stmt),
	}

	return db, nil
}

func (db *DB) GetQuery(query_name string) (*sql.Stmt, error) {
	if stmt, ok := db.statement_cache[query_name]; ok {
		return stmt, nil
	}

	file_path := fmt.Sprintf("sql/queries/%s.sql", query_name)
	query, err := db.files.ReadFile(file_path)
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
