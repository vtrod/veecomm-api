package controllers

import (
	"context"
	"github.com/vtrod/veecomm-api/ent"
	"github.com/vtrod/veecomm-api/ent/address"
	"github.com/vtrod/veecomm-api/ent/user"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

// Estrutura para criar/atualizar endereço
type AddressRequest struct {
	CEP         string `json:"cep"`
	Logradouro  string `json:"logradouro"`
	Numero      string `json:"numero"`
	Complemento string `json:"complemento,omitempty"`
	Bairro      string `json:"bairro"`
	Cidade      string `json:"cidade"`
	Estado      string `json:"estado"`
	IsDefault   bool   `json:"is_default"`
}

// GetUserAddresses retorna todos os endereços do usuário
// GET /api/addresses
func GetUserAddresses(c fiber.Ctx) error {
	client := c.Locals("dbClient").(*ent.Client)
	ctx := context.Background()
	
	// Obter usuário do contexto de autenticação
	userId := getUserIdFromContext(c)
	if userId == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "Usuário não autenticado",
		})
	}
	
	// Verificar se o usuário existe
	exists, err := client.User.
		Query().
		Where(user.ID(userId)).
		Exist(ctx)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Erro ao verificar usuário",
			"error":   err.Error(),
		})
	}

	if !exists {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"message": "Usuário não encontrado",
		})
	}

	// Buscar endereços do usuário
	addresses, err := client.Address.
		Query().
		Where(address.UserID(userId)).
		All(ctx)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Erro ao buscar endereços",
			"error":   err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"addresses": addresses,
	})
}

// GetAddress retorna um endereço específico
// GET /api/addresses/:id
func GetAddress(c fiber.Ctx) error {
	id := c.Params("id")
	client := c.Locals("dbClient").(*ent.Client)
	ctx := context.Background()
	
	// Obter usuário do contexto de autenticação
	userId := getUserIdFromContext(c)
	if userId == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "Usuário não autenticado",
		})
	}

	// Buscar endereço
	addr, err := client.Address.
		Query().
		Where(
			address.ID(id),
			address.UserID(userId),
		).
		First(ctx)

	if err != nil {
		if ent.IsNotFound(err) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"message": "Endereço não encontrado",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Erro ao buscar endereço",
			"error":   err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"address": addr,
	})
}

// CreateAddress cria um novo endereço
// POST /api/addresses
func CreateAddress(c fiber.Ctx) error {
	client := c.Locals("dbClient").(*ent.Client)
	ctx := context.Background()
	
	// Obter usuário do contexto de autenticação
	userId := getUserIdFromContext(c)
	if userId == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "Usuário não autenticado",
		})
	}

	// Extrair dados do request
	var req AddressRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Dados inválidos",
			"error":   err.Error(),
		})
	}

	// Validar campos obrigatórios
	if req.CEP == "" || req.Logradouro == "" || req.Numero == "" || 
	   req.Bairro == "" || req.Cidade == "" || req.Estado == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Todos os campos são obrigatórios, exceto complemento",
		})
	}

	// Se for endereço padrão, remover padrão dos demais
	if req.IsDefault {
		_, err := client.Address.
			Update().
			Where(
				address.UserID(userId),
				address.IsDefault(true),
			).
			SetIsDefault(false).
			Save(ctx)

		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"message": "Erro ao atualizar endereços existentes",
				"error":   err.Error(),
			})
		}
	}

	// Criar endereço
	addr, err := client.Address.
		Create().
		SetID(uuid.New().String()).
		SetUserID(userId).
		SetCep(req.CEP).
		SetLogradouro(req.Logradouro).
		SetNumero(req.Numero).
		SetNillableComplemento(nilIfEmpty(req.Complemento)).
		SetBairro(req.Bairro).
		SetCidade(req.Cidade).
		SetEstado(req.Estado).
		SetIsDefault(req.IsDefault).
		Save(ctx)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Erro ao criar endereço",
			"error":   err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "Endereço criado com sucesso",
		"address": addr,
	})
}

// UpdateAddress atualiza um endereço existente
// PUT /api/addresses/:id
func UpdateAddress(c fiber.Ctx) error {
	id := c.Params("id")
	client := c.Locals("dbClient").(*ent.Client)
	ctx := context.Background()
	
	// Obter usuário do contexto de autenticação
	userId := getUserIdFromContext(c)
	if userId == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "Usuário não autenticado",
		})
	}

	// Verificar se o endereço existe e pertence ao usuário
	exists, err := client.Address.
		Query().
		Where(
			address.ID(id),
			address.UserID(userId),
		).
		Exist(ctx)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Erro ao verificar endereço",
			"error":   err.Error(),
		})
	}

	if !exists {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"message": "Endereço não encontrado",
		})
	}

	// Extrair dados do request
	var req AddressRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Dados inválidos",
			"error":   err.Error(),
		})
	}

	// Se for endereço padrão, remover padrão dos demais
	if req.IsDefault {
		_, err := client.Address.
			Update().
			Where(
				address.UserID(userId),
				address.IsDefault(true),
				address.IDNEQ(id),
			).
			SetIsDefault(false).
			Save(ctx)

		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"message": "Erro ao atualizar endereços existentes",
				"error":   err.Error(),
			})
		}
	}

	// Iniciar atualização
	update := client.Address.
		UpdateOneID(id)

	// Aplicar cada campo que foi enviado
	if req.CEP != "" {
		update = update.SetCep(req.CEP)
	}
	if req.Logradouro != "" {
		update = update.SetLogradouro(req.Logradouro)
	}
	if req.Numero != "" {
		update = update.SetNumero(req.Numero)
	}
	update = update.SetNillableComplemento(nilIfEmpty(req.Complemento))
	if req.Bairro != "" {
		update = update.SetBairro(req.Bairro)
	}
	if req.Cidade != "" {
		update = update.SetCidade(req.Cidade)
	}
	if req.Estado != "" {
		update = update.SetEstado(req.Estado)
	}
	update = update.SetIsDefault(req.IsDefault)

	// Salvar atualização
	addr, err := update.Save(ctx)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Erro ao atualizar endereço",
			"error":   err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Endereço atualizado com sucesso",
		"address": addr,
	})
}

// SetDefaultAddress define um endereço como padrão
// PUT /api/addresses/:id/default
func SetDefaultAddress(c fiber.Ctx) error {
	id := c.Params("id")
	client := c.Locals("dbClient").(*ent.Client)
	ctx := context.Background()
	
	// Obter usuário do contexto de autenticação
	userId := getUserIdFromContext(c)
	if userId == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "Usuário não autenticado",
		})
	}

	// Verificar se o endereço existe e pertence ao usuário
	exists, err := client.Address.
		Query().
		Where(
			address.ID(id),
			address.UserID(userId),
		).
		Exist(ctx)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Erro ao verificar endereço",
			"error":   err.Error(),
		})
	}

	if !exists {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"message": "Endereço não encontrado",
		})
	}

	// Remover padrão dos demais endereços
	_, err = client.Address.
		Update().
		Where(
			address.UserID(userId),
			address.IsDefault(true),
			address.IDNEQ(id),
		).
		SetIsDefault(false).
		Save(ctx)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Erro ao atualizar endereços existentes",
			"error":   err.Error(),
		})
	}

	// Definir endereço como padrão
	addr, err := client.Address.
		UpdateOneID(id).
		SetIsDefault(true).
		Save(ctx)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Erro ao definir endereço como padrão",
			"error":   err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Endereço definido como padrão com sucesso",
		"address": addr,
	})
}

// DeleteAddress remove um endereço
// DELETE /api/addresses/:id
func DeleteAddress(c fiber.Ctx) error {
	id := c.Params("id")
	client := c.Locals("dbClient").(*ent.Client)
	ctx := context.Background()
	
	// Obter usuário do contexto de autenticação
	userId := getUserIdFromContext(c)
	if userId == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "Usuário não autenticado",
		})
	}

	// Verificar se o endereço existe e pertence ao usuário
	exists, err := client.Address.
		Query().
		Where(
			address.ID(id),
			address.UserID(userId),
		).
		Exist(ctx)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Erro ao verificar endereço",
			"error":   err.Error(),
		})
	}

	if !exists {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"message": "Endereço não encontrado",
		})
	}

	// Verificar se este é o único endereço padrão
	isDefault, err := client.Address.
		Query().
		Where(
			address.ID(id),
			address.IsDefault(true),
		).
		Exist(ctx)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Erro ao verificar endereço",
			"error":   err.Error(),
		})
	}

	// Verificar se existem referências em carrinhos ou pedidos
	cartExists, err := client.Cart.
		Query().
		Where(ent.HasShippingAddressWith(address.ID(id))).
		Exist(ctx)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Erro ao verificar relações do endereço",
			"error":   err.Error(),
		})
	}

	if cartExists {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{
			"message": "Não é possível excluir o endereço pois ele está em uso em um carrinho",
		})
	}

	orderExists, err := client.Order.
		Query().
		Where(ent.HasAddressWith(address.ID(id))).
		Exist(ctx)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Erro ao verificar relações do endereço",
			"error":   err.Error(),
		})
	}

	if orderExists {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{
			"message": "Não é possível excluir o endereço pois ele está em uso em um pedido",
		})
	}

	// Excluir o endereço
	err = client.Address.
		DeleteOneID(id).
		Exec(ctx)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Erro ao excluir endereço",
			"error":   err.Error(),
		})
	}

	// Se era o endereço padrão, definir outro como padrão
	if isDefault {
		// Buscar outro endereço para definir como padrão
		otherAddr, err := client.Address.
			Query().
			Where(
				address.UserID(userId),
			).
			First(ctx)

		if err != nil && !ent.IsNotFound(err) {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"message": "Erro ao buscar outros endereços",
				"error":   err.Error(),
			})
		}

		if otherAddr != nil {
			_, err = client.Address.
				UpdateOne(otherAddr).
				SetIsDefault(true).
				Save(ctx)

			if err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"message": "Endereço excluído com sucesso, mas houve erro ao definir novo endereço padrão",
					"error":   err.Error(),
				})
			}
		}
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Endereço excluído com sucesso",
	})
} 