package controllers

import (
	"context"
	"os"
	"strconv"
	"time"
	"github.com/vtrod/veecomm-api/database"
	"github.com/vtrod/veecomm-api/ent"
	"github.com/vtrod/veecomm-api/ent/user"

	"github.com/gofiber/fiber/v3"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// LoginRequest contém as credenciais para autenticação
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// RegisterRequest contém os dados para criação de usuário
type RegisterRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

// UserResponse representa os dados de usuário retornados ao cliente
type UserResponse struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	IsAdmin   bool      `json:"isAdmin"`
	CreatedAt time.Time `json:"createdAt"`
}

// ProfileUpdateRequest contém os dados para atualização de perfil
type ProfileUpdateRequest struct {
	Name            string `json:"name"`
	CurrentPassword string `json:"currentPassword"`
	NewPassword     string `json:"newPassword"`
}

// GetAllUsers retorna todos os usuários
func GetAllUsers(c fiber.Ctx) error {
	client, err := database.NewClient()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Erro ao conectar ao banco de dados",
		})
	}
	defer client.Close()

	users, err := client.User.Query().All(context.Background())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Erro ao buscar usuários",
		})
	}

	return c.JSON(users)
}

// GetUser retorna um usuário pelo ID
func GetUser(c fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "ID inválido",
		})
	}

	client, err := database.NewClient()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Erro ao conectar ao banco de dados",
		})
	}
	defer client.Close()

	user, err := client.User.Get(context.Background(), id)
	if err != nil {
		if ent.IsNotFound(err) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Usuário não encontrado",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Erro ao buscar usuário",
		})
	}

	return c.JSON(user)
}

// LoginUser autentica um usuário e retorna um token JWT
func LoginUser(c fiber.Ctx) error {
	var loginReq LoginRequest
	if err := c.BodyParser(&loginReq); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Erro ao processar requisição",
			"error":   err.Error(),
		})
	}

	// Validar dados
	if loginReq.Email == "" || loginReq.Password == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Email e senha são obrigatórios",
		})
	}

	// Buscar usuário pelo email
	client := c.Locals("dbClient").(*ent.Client)
	ctx := context.Background()
	
	u, err := client.User.Query().
		Where(user.Email(loginReq.Email)).
		First(ctx)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "Credenciais inválidas",
		})
	}

	// Verificar senha
	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(loginReq.Password)); err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "Credenciais inválidas",
		})
	}

	// Gerar token JWT
	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)
	claims["userId"] = u.ID
	claims["email"] = u.Email
	claims["isAdmin"] = u.IsAdmin
	claims["exp"] = time.Now().Add(24 * time.Hour).Unix() // Expira em 24h

	// Assinar token
	tokenString, err := token.SignedString([]byte(os.Getenv("JWT_SECRET")))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Erro ao gerar token",
			"error":   err.Error(),
		})
	}

	// Retornar token e dados do usuário
	return c.JSON(fiber.Map{
		"token": tokenString,
		"user": UserResponse{
			ID:        u.ID,
			Name:      u.Name,
			Email:     u.Email,
			IsAdmin:   u.IsAdmin,
			CreatedAt: u.CreatedAt,
		},
	})
}

// RegisterUser cria um novo usuário
func RegisterUser(c fiber.Ctx) error {
	var registerReq RegisterRequest
	if err := c.BodyParser(&registerReq); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Erro ao processar requisição",
			"error":   err.Error(),
		})
	}

	// Validar dados
	if registerReq.Email == "" || registerReq.Password == "" || registerReq.Name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Nome, email e senha são obrigatórios",
		})
	}

	// Verificar se email já existe
	client := c.Locals("dbClient").(*ent.Client)
	ctx := context.Background()
	
	exists, err := client.User.Query().
		Where(user.Email(registerReq.Email)).
		Exist(ctx)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Erro ao verificar email",
			"error":   err.Error(),
		})
	}
	if exists {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{
			"message": "Email já está em uso",
		})
	}

	// Hash da senha
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(registerReq.Password), bcrypt.DefaultCost)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Erro ao processar senha",
			"error":   err.Error(),
		})
	}

	// Criar usuário
	u, err := client.User.Create().
		SetName(registerReq.Name).
		SetEmail(registerReq.Email).
		SetPasswordHash(string(hashedPassword)).
		SetIsAdmin(false).
		Save(ctx)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Erro ao criar usuário",
			"error":   err.Error(),
		})
	}

	// Gerar token JWT
	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)
	claims["userId"] = u.ID
	claims["email"] = u.Email
	claims["isAdmin"] = u.IsAdmin
	claims["exp"] = time.Now().Add(24 * time.Hour).Unix() // Expira em 24h

	// Assinar token
	tokenString, err := token.SignedString([]byte(os.Getenv("JWT_SECRET")))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Erro ao gerar token",
			"error":   err.Error(),
		})
	}

	// Retornar token e dados do usuário
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"token": tokenString,
		"user": UserResponse{
			ID:        u.ID,
			Name:      u.Name,
			Email:     u.Email,
			IsAdmin:   u.IsAdmin,
			CreatedAt: u.CreatedAt,
		},
	})
}

// GetUserProfile retorna o perfil do usuário autenticado
func GetUserProfile(c fiber.Ctx) error {
	// Obter ID do usuário a partir do middleware de autenticação
	userId, ok := c.Locals("userId").(int)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "Não autenticado",
		})
	}

	// Buscar usuário no banco
	client := c.Locals("dbClient").(*ent.Client)
	ctx := context.Background()
	
	u, err := client.User.Get(ctx, userId)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"message": "Usuário não encontrado",
			"error":   err.Error(),
		})
	}

	// Retornar dados do usuário (sem senha)
	return c.JSON(UserResponse{
		ID:        u.ID,
		Name:      u.Name,
		Email:     u.Email,
		IsAdmin:   u.IsAdmin,
		CreatedAt: u.CreatedAt,
	})
}

// UpdateUserProfile atualiza o perfil do usuário autenticado
func UpdateUserProfile(c fiber.Ctx) error {
	var updateReq ProfileUpdateRequest
	if err := c.BodyParser(&updateReq); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Erro ao processar requisição",
			"error":   err.Error(),
		})
	}

	// Obter ID do usuário a partir do middleware de autenticação
	userId, ok := c.Locals("userId").(int)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "Não autenticado",
		})
	}

	// Buscar usuário no banco
	client := c.Locals("dbClient").(*ent.Client)
	ctx := context.Background()
	
	u, err := client.User.Get(ctx, userId)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"message": "Usuário não encontrado",
			"error":   err.Error(),
		})
	}

	// Iniciar construção da atualização
	update := client.User.UpdateOneID(userId)
	
	// Atualizar nome se fornecido
	if updateReq.Name != "" {
		update = update.SetName(updateReq.Name)
	}

	// Atualizar senha se fornecida
	if updateReq.CurrentPassword != "" && updateReq.NewPassword != "" {
		// Verificar senha atual
		if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(updateReq.CurrentPassword)); err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"message": "Senha atual incorreta",
			})
		}

		// Gerar hash da nova senha
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(updateReq.NewPassword), bcrypt.DefaultCost)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"message": "Erro ao processar nova senha",
				"error":   err.Error(),
			})
		}

		update = update.SetPasswordHash(string(hashedPassword))
	}

	// Executar atualização
	updatedUser, err := update.Save(ctx)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Erro ao atualizar perfil",
			"error":   err.Error(),
		})
	}

	// Retornar dados atualizados
	return c.JSON(UserResponse{
		ID:        updatedUser.ID,
		Name:      updatedUser.Name,
		Email:     updatedUser.Email,
		IsAdmin:   updatedUser.IsAdmin,
		CreatedAt: updatedUser.CreatedAt,
	})
}

// UpdateUser atualiza um usuário existente
func UpdateUser(c fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "ID inválido",
		})
	}

	type userRequest struct {
		Nome  string `json:"nome"`
		Email string `json:"email"`
		Senha string `json:"senha"`
	}

	var input userRequest
	if err := c.Bind().Body(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Dados inválidos",
		})
	}

	client, err := database.NewClient()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Erro ao conectar ao banco de dados",
		})
	}
	defer client.Close()

	user, err := client.User.UpdateOneID(id).
		SetNome(input.Nome).
		SetEmail(input.Email).
		SetSenha(input.Senha).
		Save(context.Background())
	if err != nil {
		if ent.IsNotFound(err) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Usuário não encontrado",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Erro ao atualizar usuário",
		})
	}

	return c.JSON(user)
}

// DeleteUser remove um usuário
func DeleteUser(c fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "ID inválido",
		})
	}

	client, err := database.NewClient()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Erro ao conectar ao banco de dados",
		})
	}
	defer client.Close()

	err = client.User.DeleteOneID(id).Exec(context.Background())
	if err != nil {
		if ent.IsNotFound(err) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Usuário não encontrado",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Erro ao remover usuário",
		})
	}

	return c.SendStatus(fiber.StatusNoContent)
} 