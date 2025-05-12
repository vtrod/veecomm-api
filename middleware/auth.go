package middleware

import (
	"context"
	"strings"
	"github.com/vtrod/veecomm-api/ent"
	"github.com/vtrod/veecomm-api/ent/user"

	"github.com/gofiber/fiber/v3"
	"github.com/golang-jwt/jwt/v5"
)

// Config armazena as configurações do middleware de autenticação
type Config struct {
	JWTSecret string
}

// DefaultConfig retorna uma configuração padrão
func DefaultConfig() Config {
	return Config{
		JWTSecret: "your-secret-key", // Em produção, use variável de ambiente
	}
}

// New cria uma nova instância do middleware de autenticação
func New(config Config) fiber.Handler {
	// Validar configuração
	if config.JWTSecret == "" {
		config.JWTSecret = DefaultConfig().JWTSecret
	}

	// Retornar middleware handler
	return func(c fiber.Ctx) error {
		// Injetar o cliente do banco de dados no contexto
		client := c.Locals("dbClient").(*ent.Client)
		
		// Obter token do cabeçalho de autorização
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			// Definir usuário não autenticado
			c.Locals("authenticated", false)
			c.Locals("userId", "")
			c.Locals("isAdmin", false)
			return c.Next()
		}

		// Verificar formato do token (Bearer)
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			// Definir usuário não autenticado
			c.Locals("authenticated", false)
			c.Locals("userId", "")
			c.Locals("isAdmin", false)
			return c.Next()
		}

		// Extrair token
		tokenString := parts[1]

		// Parse do token JWT
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			// Verificar método de assinatura
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fiber.NewError(fiber.StatusUnauthorized, "Método de assinatura inválido")
			}
			return []byte(config.JWTSecret), nil
		})

		if err != nil || !token.Valid {
			// Definir usuário não autenticado
			c.Locals("authenticated", false)
			c.Locals("userId", "")
			c.Locals("isAdmin", false)
			return c.Next()
		}

		// Extrair claims do token
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			// Definir usuário não autenticado
			c.Locals("authenticated", false)
			c.Locals("userId", "")
			c.Locals("isAdmin", false)
			return c.Next()
		}

		// Extrair ID do usuário das claims
		userID, ok := claims["userId"].(string)
		if !ok || userID == "" {
			// Definir usuário não autenticado
			c.Locals("authenticated", false)
			c.Locals("userId", "")
			c.Locals("isAdmin", false)
			return c.Next()
		}

		// Verificar se o usuário existe no banco de dados
		ctx := context.Background()
		userObj, err := client.User.Query().Where(user.ID(userID)).First(ctx)
		if err != nil {
			// Definir usuário não autenticado
			c.Locals("authenticated", false)
			c.Locals("userId", "")
			c.Locals("isAdmin", false)
			return c.Next()
		}

		// Verificar se é admin (exemplo simples)
		isAdmin := false
		if role, ok := claims["role"].(string); ok && role == "admin" {
			isAdmin = true
		}

		// Definir informações do usuário no contexto
		c.Locals("authenticated", true)
		c.Locals("userId", userID)
		c.Locals("user", userObj)
		c.Locals("isAdmin", isAdmin)

		// Continuar com a próxima middleware/handler
		return c.Next()
	}
}

// Protected verifica se o usuário está autenticado
func Protected() fiber.Handler {
	return func(c fiber.Ctx) error {
		// Verificar se o usuário está autenticado
		authenticated, ok := c.Locals("authenticated").(bool)
		if !ok || !authenticated {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"message": "Não autorizado",
			})
		}

		// Continuar com a próxima middleware/handler
		return c.Next()
	}
}

// AdminOnly verifica se o usuário é administrador
func AdminOnly() fiber.Handler {
	return func(c fiber.Ctx) error {
		// Verificar se o usuário é admin
		isAdmin, ok := c.Locals("isAdmin").(bool)
		if !ok || !isAdmin {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"message": "Acesso restrito a administradores",
			})
		}

		// Continuar com a próxima middleware/handler
		return c.Next()
	}
} 