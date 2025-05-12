package controllers

import (
	"context"
	"time"
	"github.com/vtrod/veecomm-api/ent"
	"github.com/vtrod/veecomm-api/ent/category"
	"github.com/vtrod/veecomm-api/ent/product"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

// Estrutura para criar/atualizar categoria
type CategoryRequest struct {
	Name        string `json:"name"`
	Icon        string `json:"icon"`
	Image       string `json:"image,omitempty"`
	Description string `json:"description,omitempty"`
}

// GetAllCategories retorna todas as categorias
// GET /api/categories
func GetAllCategories(c fiber.Ctx) error {
	client := c.Locals("dbClient").(*ent.Client)
	ctx := context.Background()

	// Buscar todas as categorias
	categories, err := client.Category.
		Query().
		Order(ent.Asc("name")).
		All(ctx)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Erro ao buscar categorias",
			"error":   err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"categories": categories,
	})
}

// GetCategory retorna uma categoria pelo ID
// GET /api/categories/:id
func GetCategory(c fiber.Ctx) error {
	id := c.Params("id")
	client := c.Locals("dbClient").(*ent.Client)
	ctx := context.Background()

	// Buscar categoria por ID
	cat, err := client.Category.
		Query().
		Where(category.ID(id)).
		First(ctx)

	if err != nil {
		if ent.IsNotFound(err) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"message": "Categoria não encontrada",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Erro ao buscar categoria",
			"error":   err.Error(),
		})
	}

	// Contar produtos da categoria
	count, err := client.Product.
		Query().
		Where(product.CategoryID(id)).
		Count(ctx)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Erro ao contar produtos da categoria",
			"error":   err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"category":        cat,
		"products_count":  count,
	})
}

// CreateCategory cria uma nova categoria
// POST /api/categories
func CreateCategory(c fiber.Ctx) error {
	client := c.Locals("dbClient").(*ent.Client)
	ctx := context.Background()

	// Extrair dados do request
	var req CategoryRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Dados inválidos",
			"error":   err.Error(),
		})
	}

	// Validar campos obrigatórios
	if req.Name == "" || req.Icon == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Nome e ícone são obrigatórios",
		})
	}

	// Verificar se já existe categoria com o mesmo nome
	exists, err := client.Category.
		Query().
		Where(category.Name(req.Name)).
		Exist(ctx)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Erro ao verificar categoria",
			"error":   err.Error(),
		})
	}

	if exists {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{
			"message": "Já existe uma categoria com este nome",
		})
	}

	// Criar categoria
	cat, err := client.Category.
		Create().
		SetID(uuid.New().String()).
		SetName(req.Name).
		SetIcon(req.Icon).
		SetNillableImage(nilIfEmpty(req.Image)).
		SetNillableDescription(nilIfEmpty(req.Description)).
		Save(ctx)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Erro ao criar categoria",
			"error":   err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message":  "Categoria criada com sucesso",
		"category": cat,
	})
}

// UpdateCategory atualiza uma categoria existente
// PUT /api/categories/:id
func UpdateCategory(c fiber.Ctx) error {
	id := c.Params("id")
	client := c.Locals("dbClient").(*ent.Client)
	ctx := context.Background()

	// Verificar se a categoria existe
	exists, err := client.Category.
		Query().
		Where(category.ID(id)).
		Exist(ctx)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Erro ao verificar categoria",
			"error":   err.Error(),
		})
	}

	if !exists {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"message": "Categoria não encontrada",
		})
	}

	// Extrair dados do request
	var req CategoryRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Dados inválidos",
			"error":   err.Error(),
		})
	}

	// Se o nome for alterado, verificar unicidade
	if req.Name != "" {
		nameExists, err := client.Category.
			Query().
			Where(
				category.NameEQ(req.Name),
				category.IDNEQ(id),
			).
			Exist(ctx)

		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"message": "Erro ao verificar categoria",
				"error":   err.Error(),
			})
		}

		if nameExists {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{
				"message": "Já existe uma categoria com este nome",
			})
		}
	}

	// Iniciar atualização
	update := client.Category.
		UpdateOneID(id).
		SetUpdatedAt(time.Now())

	// Aplicar cada campo que foi enviado
	if req.Name != "" {
		update = update.SetName(req.Name)
	}
	if req.Icon != "" {
		update = update.SetIcon(req.Icon)
	}
	update = update.SetNillableImage(nilIfEmpty(req.Image))
	update = update.SetNillableDescription(nilIfEmpty(req.Description))

	// Salvar atualização
	cat, err := update.Save(ctx)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Erro ao atualizar categoria",
			"error":   err.Error(),
		})
	}

	// Se o nome foi alterado, atualizar produtos relacionados
	if req.Name != "" {
		_, err = client.Product.
			Update().
			Where(product.CategoryID(id)).
			SetCategoryName(req.Name).
			Save(ctx)

		if err != nil {
			// Não falhar a operação, apenas registrar o erro
			// O usuário ainda pode receber um sucesso, mas os produtos não serão atualizados
			// Uma tarefa de sincronização posterior pode corrigir isso
			return c.Status(fiber.StatusOK).JSON(fiber.Map{
				"message":   "Categoria atualizada com sucesso, mas houve erro ao atualizar produtos relacionados",
				"category":  cat,
				"error":     err.Error(),
			})
		}
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message":  "Categoria atualizada com sucesso",
		"category": cat,
	})
}

// DeleteCategory remove uma categoria
// DELETE /api/categories/:id
func DeleteCategory(c fiber.Ctx) error {
	id := c.Params("id")
	client := c.Locals("dbClient").(*ent.Client)
	ctx := context.Background()

	// Verificar se existem produtos nesta categoria
	count, err := client.Product.
		Query().
		Where(product.CategoryID(id)).
		Count(ctx)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Erro ao verificar produtos da categoria",
			"error":   err.Error(),
		})
	}

	if count > 0 {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{
			"message": "Não é possível excluir a categoria pois ela contém produtos",
			"products_count": count,
		})
	}

	// Excluir a categoria
	err = client.Category.
		DeleteOneID(id).
		Exec(ctx)

	if err != nil {
		if ent.IsNotFound(err) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"message": "Categoria não encontrada",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Erro ao excluir categoria",
			"error":   err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Categoria excluída com sucesso",
	})
}

// Helper para converter string vazia em nil
func nilIfEmpty(s string) *string {
	if s == "" {
		return nil
	}
	return &s
} 