package controllers

import (
	"context"
	"github.com/vtrod/veecomm-api/ent"
	"github.com/vtrod/veecomm-api/ent/coupon"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

// Estrutura para criar/atualizar cupom
type CouponRequest struct {
	Code          string     `json:"code"`
	DiscountType  string     `json:"discount_type"`
	DiscountValue float64    `json:"discount_value"`
	MinPurchase   float64    `json:"min_purchase"`
	ExpiresAt     *time.Time `json:"expires_at,omitempty"`
	IsActive      bool       `json:"is_active"`
	MaxUses       *int       `json:"max_uses,omitempty"`
}

// GetAllCoupons retorna todos os cupons
// GET /api/coupons
func GetAllCoupons(c fiber.Ctx) error {
	client := c.Locals("dbClient").(*ent.Client)
	ctx := context.Background()
	
	// Verificar se o usuário é admin
	isAdmin := c.Locals("isAdmin").(bool)
	if !isAdmin {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"message": "Apenas administradores podem listar todos os cupons",
		})
	}

	// Buscar todos os cupons
	coupons, err := client.Coupon.
		Query().
		Order(ent.Desc("created_at")).
		All(ctx)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Erro ao buscar cupons",
			"error":   err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"coupons": coupons,
	})
}

// GetCoupon retorna um cupom pelo ID
// GET /api/coupons/:id
func GetCoupon(c fiber.Ctx) error {
	id := c.Params("id")
	client := c.Locals("dbClient").(*ent.Client)
	ctx := context.Background()
	
	// Verificar se o usuário é admin
	isAdmin := c.Locals("isAdmin").(bool)
	if !isAdmin {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"message": "Apenas administradores podem visualizar detalhes de cupons",
		})
	}

	// Buscar cupom por ID
	couponObj, err := client.Coupon.
		Query().
		Where(coupon.ID(id)).
		First(ctx)

	if err != nil {
		if ent.IsNotFound(err) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"message": "Cupom não encontrado",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Erro ao buscar cupom",
			"error":   err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"coupon": couponObj,
	})
}

// ValidateCoupon verifica se um cupom é válido
// POST /api/coupons/validate
func ValidateCoupon(c fiber.Ctx) error {
	client := c.Locals("dbClient").(*ent.Client)
	ctx := context.Background()
	
	// Extrair código do cupom
	var req struct {
		Code       string  `json:"code"`
		CartTotal  float64 `json:"cart_total"`
	}
	
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Dados inválidos",
			"error":   err.Error(),
		})
	}

	if req.Code == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Código do cupom é obrigatório",
		})
	}

	// Buscar cupom pelo código
	couponObj, err := client.Coupon.
		Query().
		Where(
			coupon.Code(req.Code),
			coupon.IsActive(true),
		).
		First(ctx)

	if err != nil {
		if ent.IsNotFound(err) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"message": "Cupom não encontrado ou inativo",
				"valid":   false,
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Erro ao buscar cupom",
			"error":   err.Error(),
		})
	}

	// Verificar se expirou
	if couponObj.ExpiresAt != nil && couponObj.ExpiresAt.Before(time.Now()) {
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"message": "Cupom expirado",
			"valid":   false,
		})
	}

	// Verificar número máximo de usos
	if couponObj.MaxUses != nil && couponObj.TimesUsed >= *couponObj.MaxUses {
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"message": "Cupom atingiu o limite máximo de usos",
			"valid":   false,
		})
	}

	// Verificar valor mínimo de compra
	if req.CartTotal < couponObj.MinPurchase {
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"message":     "Valor mínimo para uso do cupom não atingido",
			"valid":       false,
			"min_purchase": couponObj.MinPurchase,
		})
	}

	// Calcular valor do desconto
	var discount float64
	if couponObj.DiscountType == "percentage" {
		discount = req.CartTotal * (couponObj.DiscountValue / 100)
	} else { // fixed
		discount = couponObj.DiscountValue
		if discount > req.CartTotal {
			discount = req.CartTotal
		}
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message":  "Cupom válido",
		"valid":    true,
		"coupon":   couponObj,
		"discount": discount,
	})
}

// CreateCoupon cria um novo cupom
// POST /api/coupons
func CreateCoupon(c fiber.Ctx) error {
	client := c.Locals("dbClient").(*ent.Client)
	ctx := context.Background()
	
	// Verificar se o usuário é admin
	isAdmin := c.Locals("isAdmin").(bool)
	if !isAdmin {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"message": "Apenas administradores podem criar cupons",
		})
	}

	// Extrair dados do request
	var req CouponRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Dados inválidos",
			"error":   err.Error(),
		})
	}

	// Validar campos obrigatórios
	if req.Code == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Código do cupom é obrigatório",
		})
	}

	// Validar tipo de desconto
	if req.DiscountType != "percentage" && req.DiscountType != "fixed" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Tipo de desconto inválido. Use 'percentage' ou 'fixed'",
		})
	}

	// Validar valor do desconto
	if req.DiscountValue <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Valor do desconto deve ser maior que zero",
		})
	}

	// Validar valor mínimo de compra
	if req.MinPurchase < 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Valor mínimo de compra não pode ser negativo",
		})
	}

	// Verificar se já existe cupom com o mesmo código
	exists, err := client.Coupon.
		Query().
		Where(coupon.Code(req.Code)).
		Exist(ctx)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Erro ao verificar cupom",
			"error":   err.Error(),
		})
	}

	if exists {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{
			"message": "Já existe um cupom com este código",
		})
	}

	// Criar cupom
	couponBuilder := client.Coupon.
		Create().
		SetID(uuid.New().String()).
		SetCode(req.Code).
		SetDiscountType(req.DiscountType).
		SetDiscountValue(req.DiscountValue).
		SetMinPurchase(req.MinPurchase).
		SetIsActive(req.IsActive).
		SetTimesUsed(0)

	// Adicionar campos opcionais
	if req.ExpiresAt != nil {
		couponBuilder = couponBuilder.SetExpiresAt(*req.ExpiresAt)
	}
	if req.MaxUses != nil {
		couponBuilder = couponBuilder.SetMaxUses(*req.MaxUses)
	}

	// Salvar cupom
	couponObj, err := couponBuilder.Save(ctx)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Erro ao criar cupom",
			"error":   err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "Cupom criado com sucesso",
		"coupon":  couponObj,
	})
}

// UpdateCoupon atualiza um cupom existente
// PUT /api/coupons/:id
func UpdateCoupon(c fiber.Ctx) error {
	id := c.Params("id")
	client := c.Locals("dbClient").(*ent.Client)
	ctx := context.Background()
	
	// Verificar se o usuário é admin
	isAdmin := c.Locals("isAdmin").(bool)
	if !isAdmin {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"message": "Apenas administradores podem atualizar cupons",
		})
	}

	// Verificar se o cupom existe
	couponObj, err := client.Coupon.
		Query().
		Where(coupon.ID(id)).
		First(ctx)

	if err != nil {
		if ent.IsNotFound(err) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"message": "Cupom não encontrado",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Erro ao buscar cupom",
			"error":   err.Error(),
		})
	}

	// Extrair dados do request
	var req CouponRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Dados inválidos",
			"error":   err.Error(),
		})
	}

	// Validar tipo de desconto
	if req.DiscountType != "" && req.DiscountType != "percentage" && req.DiscountType != "fixed" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Tipo de desconto inválido. Use 'percentage' ou 'fixed'",
		})
	}

	// Validar valor do desconto
	if req.DiscountValue < 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Valor do desconto não pode ser negativo",
		})
	}

	// Validar valor mínimo de compra
	if req.MinPurchase < 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Valor mínimo de compra não pode ser negativo",
		})
	}

	// Verificar unicidade do código se for alterado
	if req.Code != "" && req.Code != couponObj.Code {
		exists, err := client.Coupon.
			Query().
			Where(
				coupon.Code(req.Code),
				coupon.IDNEQ(id),
			).
			Exist(ctx)

		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"message": "Erro ao verificar código do cupom",
				"error":   err.Error(),
			})
		}

		if exists {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{
				"message": "Já existe um cupom com este código",
			})
		}
	}

	// Iniciar atualização
	update := client.Coupon.
		UpdateOne(couponObj).
		SetUpdatedAt(time.Now())

	// Aplicar cada campo que foi enviado
	if req.Code != "" {
		update = update.SetCode(req.Code)
	}
	if req.DiscountType != "" {
		update = update.SetDiscountType(req.DiscountType)
	}
	if req.DiscountValue > 0 {
		update = update.SetDiscountValue(req.DiscountValue)
	}
	if req.MinPurchase >= 0 {
		update = update.SetMinPurchase(req.MinPurchase)
	}
	// IsActive pode ser true ou false, então sempre atualizamos
	update = update.SetIsActive(req.IsActive)
	if req.ExpiresAt != nil {
		update = update.SetExpiresAt(*req.ExpiresAt)
	}
	if req.MaxUses != nil {
		update = update.SetMaxUses(*req.MaxUses)
	}

	// Salvar atualização
	updatedCoupon, err := update.Save(ctx)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Erro ao atualizar cupom",
			"error":   err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Cupom atualizado com sucesso",
		"coupon":  updatedCoupon,
	})
}

// DeleteCoupon remove um cupom
// DELETE /api/coupons/:id
func DeleteCoupon(c fiber.Ctx) error {
	id := c.Params("id")
	client := c.Locals("dbClient").(*ent.Client)
	ctx := context.Background()
	
	// Verificar se o usuário é admin
	isAdmin := c.Locals("isAdmin").(bool)
	if !isAdmin {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"message": "Apenas administradores podem excluir cupons",
		})
	}

	// Verificar se o cupom existe
	exists, err := client.Coupon.
		Query().
		Where(coupon.ID(id)).
		Exist(ctx)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Erro ao verificar cupom",
			"error":   err.Error(),
		})
	}

	if !exists {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"message": "Cupom não encontrado",
		})
	}

	// Excluir o cupom
	err = client.Coupon.
		DeleteOneID(id).
		Exec(ctx)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Erro ao excluir cupom",
			"error":   err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Cupom excluído com sucesso",
	})
} 