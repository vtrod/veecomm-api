package database

import (
	"log"
	"github.com/vtrod/veecomm-api/ent"
	_ "github.com/go-sql-driver/mysql"
)

// NewClient cria uma nova conexão com o banco de dados
func NewClient() (*ent.Client, error) {
	// Conexão com MariaDB
	// DSN formato: [username[:password]@][protocol[(address)]]/dbname[?param=value]
	dsn := "root:root@tcp(localhost:3306)/veecomm?parseTime=True"
	client, err := ent.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("Falha ao abrir conexão com MariaDB: %v", err)
		return nil, err
	}
	
	return client, nil
} 