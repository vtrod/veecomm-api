package ent

import (
	"context"
	"fmt"
	"log"

	"entgo.io/ent/dialect/sql/schema"
)

// MigrateDatabase executará a migração com as opções corretas
func MigrateDatabase(client *Client) error {
	ctx := context.Background()
	
	// Criando um plano de migração
	err := client.Schema.Create(
		ctx,
		schema.WithDropIndex(true),
		schema.WithDropColumn(true),
		schema.WithAtlas(true),
	)
	
	if err != nil {
		return fmt.Errorf("failed creating schema resources: %w", err)
	}
	
	log.Println("Database migration completed successfully!")
	return nil
} 