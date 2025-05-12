package controllers

import (
	"context"
	"strconv"
	"time"
	"github.com/vtrod/veecomm-api/ent"
	"github.com/vtrod/veecomm-api/ent/category"
	"github.com/vtrod/veecomm-api/ent/product"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

// Estrutura para criar/atualizar produto
type ProductRequest struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Price       float64                `json:"price"`
	OldPrice    *float64               `json:"old_price,omitempty"`
	Image       string                 `json:"image"`
	CategoryID  string                 `json:"category_id"`
	Featured    bool                   `json:"featured"`
	Details     map[string]interface{} `json:"details,omitempty"`
}

// GetAllProducts retorna todos os produtos
// GET /api/products
func GetAllProducts(c fiber.Ctx) error {
	client := c.Locals("dbClient").(*ent.Client)
	ctx := context.Background()

	// Parâmetros de paginação
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "20"))
	offset := (page - 1) * limit

	// Buscar produtos com paginação
	products, err := client.Product.
		Query().
		Limit(limit).
		Offset(offset).
		Order(ent.Desc("created_at")).
		All(ctx)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Erro ao buscar produtos",
			"error":   err.Error(),
		})
	}

	// Contar total para paginação
	total, err := client.Product.Query().Count(ctx)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Erro ao contar produtos",
			"error":   err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"data": products,
		"meta": fiber.Map{
			"total":       total,
			"page":        page,
			"limit":       limit,
			"total_pages": (total + limit - 1) / limit,
		},
	})
}

// GetProduct retorna um produto pelo ID
// GET /api/products/:id
func GetProduct(c fiber.Ctx) error {
	id := c.Params("id")
	client := c.Locals("dbClient").(*ent.Client)
	ctx := context.Background()

	// Buscar produto por ID
	prod, err := client.Product.
		Query().
		Where(product.ID(id)).
		WithCategory().
		First(ctx)

	if err != nil {
		if ent.IsNotFound(err) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"message": "Produto não encontrado",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Erro ao buscar produto",
			"error":   err.Error(),
		})
	}

	// Buscar avaliações do produto
	avaliations, err := client.Avaliation.
		Query().
		Where(ent.HasProductWith(product.ID(id))).
		Order(ent.Desc("date")).
		Limit(10).
		All(ctx)

	if err != nil && !ent.IsNotFound(err) {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Erro ao buscar avaliações",
			"error":   err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"product":     prod,
		"avaliations": avaliations,
	})
}

// GetProductsByCategory retorna produtos de uma categoria específica
// GET /api/products/category/:categoryId
func GetProductsByCategory(c fiber.Ctx) error {
	categoryId := c.Params("categoryId")
	client := c.Locals("dbClient").(*ent.Client)
	ctx := context.Background()

	// Verificar se a categoria existe
	cat, err := client.Category.Get(ctx, categoryId)
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

	// Parâmetros de paginação
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "20"))
	offset := (page - 1) * limit

	// Buscar produtos por categoria
	products, err := client.Product.
		Query().
		Where(product.CategoryID(categoryId)).
		Limit(limit).
		Offset(offset).
		Order(ent.Desc("created_at")).
		All(ctx)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Erro ao buscar produtos da categoria",
			"error":   err.Error(),
		})
	}

	// Contar total para paginação
	total, err := client.Product.
		Query().
		Where(product.CategoryID(categoryId)).
		Count(ctx)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Erro ao contar produtos da categoria",
			"error":   err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"category": cat,
		"products": products,
		"meta": fiber.Map{
			"total":       total,
			"page":        page,
			"limit":       limit,
			"total_pages": (total + limit - 1) / limit,
		},
	})
}

// GetPromotionProducts retorna produtos em promoção
// GET /api/products/promotions
func GetPromotionProducts(c fiber.Ctx) error {
	client := c.Locals("dbClient").(*ent.Client)
	ctx := context.Background()

	// Parâmetros de paginação
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "20"))
	offset := (page - 1) * limit

	// Buscar produtos em promoção (com old_price não nulo)
	products, err := client.Product.
		Query().
		Where(product.OldPriceNotNil()).
		Limit(limit).
		Offset(offset).
		Order(ent.Desc("created_at")).
		All(ctx)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Erro ao buscar produtos em promoção",
			"error":   err.Error(),
		})
	}

	// Contar total para paginação
	total, err := client.Product.
		Query().
		Where(product.OldPriceNotNil()).
		Count(ctx)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Erro ao contar produtos em promoção",
			"error":   err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"products": products,
		"meta": fiber.Map{
			"total":       total,
			"page":        page,
			"limit":       limit,
			"total_pages": (total + limit - 1) / limit,
		},
	})
}

// CreateProduct cria um novo produto
// POST /api/products
func CreateProduct(c fiber.Ctx) error {
	client := c.Locals("dbClient").(*ent.Client)
	ctx := context.Background()

	// Extrair dados do request
	var req ProductRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Dados inválidos",
			"error":   err.Error(),
		})
	}

	// Validar preço positivo
	if req.Price <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "O preço deve ser maior que zero",
		})
	}

	// Verificar se a categoria existe e buscar nome
	category, err := client.Category.Get(ctx, req.CategoryID)
	if err != nil {
		if ent.IsNotFound(err) {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"message": "Categoria não encontrada",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Erro ao verificar categoria",
			"error":   err.Error(),
		})
	}

	// Criar produto
	prod, err := client.Product.
		Create().
		SetID(uuid.New().String()).
		SetName(req.Name).
		SetDescription(req.Description).
		SetPrice(req.Price).
		SetNillableOldPrice(req.OldPrice).
		SetImage(req.Image).
		SetCategoryID(req.CategoryID).
		SetCategoryName(category.Name).
		SetFeatured(req.Featured).
		SetNillableDetails(req.Details).
		SetAverageRating(0).
		SetTotalAvaliations(0).
		Save(ctx)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Erro ao criar produto",
			"error":   err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "Produto criado com sucesso",
		"product": prod,
	})
}

// UpdateProduct atualiza um produto existente
// PUT /api/products/:id
func UpdateProduct(c fiber.Ctx) error {
	id := c.Params("id")
	client := c.Locals("dbClient").(*ent.Client)
	ctx := context.Background()

	// Verificar se o produto existe
	exists, err := client.Product.Query().Where(product.ID(id)).Exist(ctx)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Erro ao verificar produto",
			"error":   err.Error(),
		})
	}
	if !exists {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"message": "Produto não encontrado",
		})
	}

	// Extrair dados do request
	var req ProductRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Dados inválidos",
			"error":   err.Error(),
		})
	}

	// Validar preço positivo
	if req.Price <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "O preço deve ser maior que zero",
		})
	}

	// Verificar se houve mudança de categoria e buscar nome
	var categoryName string
	if req.CategoryID != "" {
		category, err := client.Category.Get(ctx, req.CategoryID)
		if err != nil {
			if ent.IsNotFound(err) {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"message": "Categoria não encontrada",
				})
			}
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"message": "Erro ao verificar categoria",
				"error":   err.Error(),
			})
		}
		categoryName = category.Name
	}

	// Iniciar atualização
	update := client.Product.UpdateOneID(id).
		SetUpdatedAt(time.Now())

	// Aplicar cada campo que foi enviado
	if req.Name != "" {
		update = update.SetName(req.Name)
	}
	if req.Description != "" {
		update = update.SetDescription(req.Description)
	}
	if req.Price > 0 {
		update = update.SetPrice(req.Price)
	}
	if req.OldPrice != nil {
		update = update.SetNillableOldPrice(req.OldPrice)
	}
	if req.Image != "" {
		update = update.SetImage(req.Image)
	}
	if req.CategoryID != "" {
		update = update.SetCategoryID(req.CategoryID).
			SetCategoryName(categoryName)
	}
	// Featured pode ser true ou false, então sempre atualizamos
	update = update.SetFeatured(req.Featured)
	if req.Details != nil {
		update = update.SetDetails(req.Details)
	}

	// Salvar atualização
	prod, err := update.Save(ctx)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Erro ao atualizar produto",
			"error":   err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Produto atualizado com sucesso",
		"product": prod,
	})
}

// DeleteProduct remove um produto
// DELETE /api/products/:id
func DeleteProduct(c fiber.Ctx) error {
	id := c.Params("id")
	client := c.Locals("dbClient").(*ent.Client)
	ctx := context.Background()

	// Verificar se existem referencias ao produto em carrinhos ou pedidos
	cartItemExists, err := client.CartItem.
		Query().
		Where(ent.HasProductWith(product.ID(id))).
		Exist(ctx)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Erro ao verificar relações do produto",
			"error":   err.Error(),
		})
	}

	if cartItemExists {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{
			"message": "Não é possível excluir o produto pois ele está em carrinhos de compra",
		})
	}

	orderItemExists, err := client.OrderItem.
		Query().
		Where(ent.HasProductWith(product.ID(id))).
		Exist(ctx)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Erro ao verificar relações do produto",
			"error":   err.Error(),
		})
	}

	if orderItemExists {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{
			"message": "Não é possível excluir o produto pois ele está em pedidos",
		})
	}

	// Excluir avaliações do produto primeiro
	_, err = client.Avaliation.
		Delete().
		Where(ent.HasProductWith(product.ID(id))).
		Exec(ctx)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Erro ao excluir avaliações do produto",
			"error":   err.Error(),
		})
	}

	// Excluir o produto
	err = client.Product.
		DeleteOneID(id).
		Exec(ctx)

	if err != nil {
		if ent.IsNotFound(err) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"message": "Produto não encontrado",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Erro ao excluir produto",
			"error":   err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Produto excluído com sucesso",
	})
} 