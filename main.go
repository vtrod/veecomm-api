package main

import (
	"context"
	"log"
	"os"
	"github.com/vtrod/veecomm-api/database"
	"github.com/vtrod/veecomm-api/middleware"
	"github.com/vtrod/veecomm-api/routes"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/gofiber/fiber/v3/middleware/logger"
	"github.com/gofiber/fiber/v3/middleware/recover"
	"github.com/joho/godotenv"
)

func main() {
	// Carregar variáveis de ambiente
	if err := godotenv.Load(); err != nil {
		log.Println("Arquivo .env não encontrado, usando variáveis de ambiente do sistema")
	}

	// Inicializar o banco de dados
	client, err := database.NewClient()
	if err != nil {
		log.Fatalf("Falha ao conectar ao banco de dados: %v", err)
	}
	defer client.Close()
	
	// Executar migração automática
	if err := client.Schema.Create(context.Background()); err != nil {
		log.Fatalf("Falha ao criar schema: %v", err)
	}

	// Inicializar aplicação Fiber
	app := fiber.New(fiber.Config{
		AppName:      "VeeComm API",
		ErrorHandler: customErrorHandler,
	})

	// Configurar middlewares globais
	app.Use(recover.New())
	app.Use(logger.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins:     os.Getenv("CORS_ALLOW_ORIGINS"),
		AllowMethods:     "GET,POST,PUT,DELETE,OPTIONS",
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization",
		AllowCredentials: true,
	}))

	// Middleware para injetar o cliente do banco de dados
	app.Use(func(c fiber.Ctx) error {
		c.Locals("dbClient", client)
		return c.Next()
	})

	// Aplicar middleware de autenticação para todas as rotas
	app.Use(middleware.New(middleware.Config{
		JWTSecret: os.Getenv("JWT_SECRET"),
	}))

	// Configurar rotas
	routes.SetupRoutes(app)

	// Determinar porta
	port := os.Getenv("PORT")
	if port == "" {
		port = "8001"
	}

	// Iniciar servidor
	log.Fatal(app.Listen(":" + port))
}

// customErrorHandler lida com erros da aplicação
func customErrorHandler(c fiber.Ctx, err error) error {
	// Status code padrão
	code := fiber.StatusInternalServerError

	// Verificar se é um erro do Fiber
	if e, ok := err.(*fiber.Error); ok {
		code = e.Code
	}

	// Responder com erro em formato JSON
	return c.Status(code).JSON(fiber.Map{
		"message": err.Error(),
	})
}
