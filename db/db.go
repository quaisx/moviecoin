package db

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/jackc/pgx/v4/pgxpool"
)

const (
	host        = "localhost" //docker PostgreSql is running on -p 5432:5432
	port        = 5432
	user        = "postgres"
	password    = "password"
	dbname      = "active"
	PGXPOOL_MIN = 3
	PGXPOOL_MAX = 10
)

type Database struct {
	database *pgxpool.Pool
	ctx      context.Context
}

func (db *Database) CreateTable(table string, layout []string) (status bool) {
	stmt := "CREATE TABLE IF NOT EXIST $1 (%s);"
	layout_ := strings.Join(layout[:len(layout)-1], ",") + layout[len(layout)-1]
	_, err := db.database.Exec(db.ctx, fmt.Sprintf(stmt, layout_), table)
	if err == nil {
		status = true
	}
	return
}

func (db *Database) DropTable(table string) (status bool) {
	stmt := fmt.Sprintf("DROP TABLE IF EXISTS %s;", table)
	_, err := db.database.Exec(db.ctx, stmt)
	if err == nil {
		status = true
	}
	return
}

func (db *Database) TableExists(table string) (exists bool) {
	var count int
	db.database.QueryRow(db.ctx, `SELECT COUNT(table_name)
		FROM
			information_schema.tables
		WHERE
			table_schema LIKE 'public' AND
			table_type LIKE 'BASE TABLE' AND
			table_name = $1;`, table).Scan(&count)
	if count == 1 {
		exists = true
	}
	return
}

func (db *Database) NewConnection(conn_str string) (err error) {
	config, err := pgxpool.ParseConfig(conn_str)
	if err != nil {
		log.Printf("Parse connection string error: %+v", err)
		return err
	}
	config.MinConns, config.MaxConns = PGXPOOL_MIN, PGXPOOL_MAX
	db.database, err = pgxpool.ConnectConfig(context.Background(), config)
	if err != nil {
		log.Printf("New Postgres connection establishment failed: %v", err)
		db.database = nil
		return err
	}
	return nil
}

func (db *Database) Close() {
	if db != nil && db.database != nil {
		db.database.Close()
	}
}
