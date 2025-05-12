package controllers

import (
	"context"
	"time"
	"github.com/vtrod/veecomm-api/ent"
	"github.com/vtrod/veecomm-api/ent/product"
	"github.com/vtrod/veecomm-api/ent/user"
	"github.com/vtrod/veecomm-api/ent/avaliation"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

// Estrutura para criar/atualizar avaliação
type ReviewRequest struct {
	Rating  int      `json:"rating"`
	Comment string   `json:"comment"`
	Images  []string `json:"images,omitempty"`
}

// GetProductReviews retorna as avaliações de um produto
// GET /api/products/:productId/reviews
func GetProductReviews(c fiber.Ctx) error {
	productId := c.Params("productId")
	client := c.Locals("dbClient").(*ent.Client)
	ctx := context.Background()

	// Verificar se o produto existe
	exists, err := client.Product.
		Query().
		Where(product.ID(productId)).
		Exist(ctx)

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

	// Buscar avaliações do produto
	reviews, err := client.Avaliation.
		Query().
		Where(avaliation.ProductID(productId)).
		Order(ent.Desc("date")).
		All(ctx)

	if err != nil && !ent.IsNotFound(err) {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Erro ao buscar avaliações",
			"error":   err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"reviews": reviews,
	})
}

// AddProductReview adiciona uma avaliação a um produto
// POST /api/products/:productId/reviews
func AddProductReview(c fiber.Ctx) error {
	productId := c.Params("productId")
	client := c.Locals("dbClient").(*ent.Client)
	ctx := context.Background()
	
	// Obter usuário do contexto de autenticação
	userId := getUserIdFromContext(c)
	if userId == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "Usuário não autenticado",
		})
	}

	// Verificar se o produto existe
	prod, err := client.Product.
		Query().
		Where(product.ID(productId)).
		First(ctx)

	if err != nil {
		if ent.IsNotFound(err) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"message": "Produto não encontrado",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Erro ao verificar produto",
			"error":   err.Error(),
		})
	}

	// Verificar se o usuário existe
	userObj, err := client.User.
		Query().
		Where(user.ID(userId)).
		First(ctx)

	if err != nil {
		if ent.IsNotFound(err) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"message": "Usuário não encontrado",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Erro ao verificar usuário",
			"error":   err.Error(),
		})
	}

	// Verificar se o usuário já avaliou este produto
	exists, err := client.Avaliation.
		Query().
		Where(
			avaliation.ProductID(productId),
			avaliation.UserID(userId),
		).
		Exist(ctx)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Erro ao verificar avaliação",
			"error":   err.Error(),
		})
	}

	if exists {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{
			"message": "Você já avaliou este produto",
		})
	}

	// Extrair dados do request
	var req ReviewRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Dados inválidos",
			"error":   err.Error(),
		})
	}

	// Validar rating entre 1 e 5
	if req.Rating < 1 || req.Rating > 5 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "A avaliação deve ser entre 1 e 5 estrelas",
		})
	}

	// Validar comentário
	if req.Comment == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "O comentário é obrigatório",
		})
	}

	// Criar avaliação
	review, err := client.Avaliation.
		Create().
		SetID(uuid.New().String()).
		SetProductID(productId).
		SetUserID(userId).
		SetUserName(userObj.Name).
		SetRating(req.Rating).
		SetComment(req.Comment).
		SetDate(time.Now()).
		SetNillableImages(req.Images).
		Save(ctx)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Erro ao criar avaliação",
			"error":   err.Error(),
		})
	}

	// Atualizar média e contagem de avaliações do produto
	reviews, err := client.Avaliation.
		Query().
		Where(avaliation.ProductID(productId)).
		All(ctx)

	if err != nil {
		// Não falhar a operação, apenas registrar o erro
		return c.Status(fiber.StatusCreated).JSON(fiber.Map{
			"message": "Avaliação criada com sucesso, mas houve erro ao atualizar estatísticas do produto",
			"review":  review,
			"error":   err.Error(),
		})
	}

	// Calcular nova média
	var sum float64
	for _, r := range reviews {
		sum += float64(r.Rating)
	}
	
	avgRating := sum / float64(len(reviews))
	totalAvaliations := len(reviews)

	// Atualizar produto
	_, err = client.Product.
		UpdateOneID(productId).
		SetAverageRating(avgRating).
		SetTotalAvaliations(totalAvaliations).
		Save(ctx)

	if err != nil {
		// Não falhar a operação, apenas registrar o erro
		return c.Status(fiber.StatusCreated).JSON(fiber.Map{
			"message": "Avaliação criada com sucesso, mas houve erro ao atualizar estatísticas do produto",
			"review":  review,
			"error":   err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "Avaliação criada com sucesso",
		"review":  review,
	})
}

// UpdateProductReview atualiza uma avaliação
// PUT /api/products/:productId/reviews/:reviewId
func UpdateProductReview(c fiber.Ctx) error {
	productId := c.Params("productId")
	reviewId := c.Params("reviewId")
	client := c.Locals("dbClient").(*ent.Client)
	ctx := context.Background()
	
	// Obter usuário do contexto de autenticação
	userId := getUserIdFromContext(c)
	if userId == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "Usuário não autenticado",
		})
	}

	// Buscar a avaliação
	review, err := client.Avaliation.
		Query().
		Where(
			avaliation.ID(reviewId),
			avaliation.ProductID(productId),
		).
		First(ctx)

	if err != nil {
		if ent.IsNotFound(err) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"message": "Avaliação não encontrada",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Erro ao buscar avaliação",
			"error":   err.Error(),
		})
	}

	// Verificar se a avaliação pertence ao usuário
	if review.UserID != userId {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"message": "Você não tem permissão para editar esta avaliação",
		})
	}

	// Extrair dados do request
	var req ReviewRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Dados inválidos",
			"error":   err.Error(),
		})
	}

	// Validar rating entre 1 e 5
	if req.Rating < 1 || req.Rating > 5 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "A avaliação deve ser entre 1 e 5 estrelas",
		})
	}

	// Iniciar atualização
	update := client.Avaliation.
		UpdateOneID(reviewId).
		SetUpdatedAt(time.Now())

	// Aplicar cada campo que foi enviado
	if req.Rating > 0 {
		update = update.SetRating(req.Rating)
	}
	if req.Comment != "" {
		update = update.SetComment(req.Comment)
	}
	if req.Images != nil {
		update = update.SetImages(req.Images)
	}

	// Salvar atualização
	updatedReview, err := update.Save(ctx)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Erro ao atualizar avaliação",
			"error":   err.Error(),
		})
	}

	// Atualizar média e contagem de avaliações do produto se o rating mudou
	if req.Rating > 0 && req.Rating != review.Rating {
		reviews, err := client.Avaliation.
			Query().
			Where(avaliation.ProductID(productId)).
			All(ctx)

		if err != nil {
			// Não falhar a operação, apenas registrar o erro
			return c.Status(fiber.StatusOK).JSON(fiber.Map{
				"message": "Avaliação atualizada com sucesso, mas houve erro ao atualizar estatísticas do produto",
				"review":  updatedReview,
				"error":   err.Error(),
			})
		}

		// Calcular nova média
		var sum float64
		for _, r := range reviews {
			sum += float64(r.Rating)
		}
		
		avgRating := sum / float64(len(reviews))

		// Atualizar produto
		_, err = client.Product.
			UpdateOneID(productId).
			SetAverageRating(avgRating).
			Save(ctx)

		if err != nil {
			// Não falhar a operação, apenas registrar o erro
			return c.Status(fiber.StatusOK).JSON(fiber.Map{
				"message": "Avaliação atualizada com sucesso, mas houve erro ao atualizar estatísticas do produto",
				"review":  updatedReview,
				"error":   err.Error(),
			})
		}
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Avaliação atualizada com sucesso",
		"review":  updatedReview,
	})
}

// DeleteProductReview remove uma avaliação
// DELETE /api/products/:productId/reviews/:reviewId
func DeleteProductReview(c fiber.Ctx) error {
	productId := c.Params("productId")
	reviewId := c.Params("reviewId")
	client := c.Locals("dbClient").(*ent.Client)
	ctx := context.Background()
	
	// Obter usuário do contexto de autenticação
	userId := getUserIdFromContext(c)
	if userId == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "Usuário não autenticado",
		})
	}

	// Buscar a avaliação
	review, err := client.Avaliation.
		Query().
		Where(
			avaliation.ID(reviewId),
			avaliation.ProductID(productId),
		).
		First(ctx)

	if err != nil {
		if ent.IsNotFound(err) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"message": "Avaliação não encontrada",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Erro ao buscar avaliação",
			"error":   err.Error(),
		})
	}

	// Verificar se a avaliação pertence ao usuário (ou se é admin)
	isAdmin := c.Locals("isAdmin").(bool)
	if review.UserID != userId && !isAdmin {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"message": "Você não tem permissão para excluir esta avaliação",
		})
	}

	// Excluir a avaliação
	err = client.Avaliation.
		DeleteOneID(reviewId).
		Exec(ctx)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Erro ao excluir avaliação",
			"error":   err.Error(),
		})
	}

	// Atualizar média e contagem de avaliações do produto
	reviews, err := client.Avaliation.
		Query().
		Where(avaliation.ProductID(productId)).
		All(ctx)

	if err != nil {
		// Não falhar a operação, apenas registrar o erro
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"message": "Avaliação excluída com sucesso, mas houve erro ao atualizar estatísticas do produto",
			"error":   err.Error(),
		})
	}

	// Calcular nova média ou zerar se não houver mais avaliações
	avgRating := 0.0
	totalAvaliations := len(reviews)
	
	if totalAvaliations > 0 {
		var sum float64
		for _, r := range reviews {
			sum += float64(r.Rating)
		}
		avgRating = sum / float64(totalAvaliations)
	}

	// Atualizar produto
	_, err = client.Product.
		UpdateOneID(productId).
		SetAverageRating(avgRating).
		SetTotalAvaliations(totalAvaliations).
		Save(ctx)

	if err != nil {
		// Não falhar a operação, apenas registrar o erro
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"message": "Avaliação excluída com sucesso, mas houve erro ao atualizar estatísticas do produto",
			"error":   err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Avaliação excluída com sucesso",
	})
}

// Helper para obter ID do usuário do contexto
func getUserIdFromContext(c fiber.Ctx) string {
	// Assumindo que o ID do usuário está no contexto após autenticação
	userId, ok := c.Locals("userId").(string)
	if !ok {
		return ""
	}
	return userId
} 