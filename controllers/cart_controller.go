package controllers

import (
	"context"
	"strconv"
	"github.com/vtrod/veecomm-api/database"
	"github.com/vtrod/veecomm-api/ent"
	"github.com/vtrod/veecomm-api/ent/cart"
	"github.com/vtrod/veecomm-api/ent/cart_item"
	"github.com/vtrod/veecomm-api/ent/product"
	"github.com/vtrod/veecomm-api/ent/coupon"
	"github.com/vtrod/veecomm-api/ent/address"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

// Estrutura para adicionar/atualizar item no carrinho
type CartItemRequest struct {
	ProductID string `json:"product_id"`
	Quantity  int    `json:"quantity"`
}

// Estrutura para aplicar cupom
type CouponRequest struct {
	Code string `json:"code"`
}

// Estrutura para atualizar endereço de entrega
type ShippingAddressRequest struct {
	AddressID string `json:"address_id"`
}

// GetCart retorna o carrinho do usuário atual
// GET /api/cart
func GetCart(c fiber.Ctx) error {
	client := c.Locals("dbClient").(*ent.Client)
	ctx := context.Background()
	
	// Obter usuário do contexto de autenticação
	userId := getUserIdFromContext(c)
	if userId == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "Usuário não autenticado",
		})
	}

	// Buscar ou criar carrinho para o usuário
	cartObj, err := getOrCreateCart(ctx, client, userId)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Erro ao buscar carrinho",
			"error":   err.Error(),
		})
	}

	// Buscar itens do carrinho
	items, err := client.CartItem.
		Query().
		Where(cart_item.CartID(cartObj.ID)).
		All(ctx)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Erro ao buscar itens do carrinho",
			"error":   err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"cart":  cartObj,
		"items": items,
	})
}

// AddToCart adiciona um item ao carrinho
// POST /api/cart/items
func AddToCart(c fiber.Ctx) error {
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
	var req CartItemRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Dados inválidos",
			"error":   err.Error(),
		})
	}

	// Validar quantidade positiva
	if req.Quantity <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "A quantidade deve ser maior que zero",
		})
	}

	// Verificar se o produto existe
	prod, err := client.Product.
		Query().
		Where(product.ID(req.ProductID)).
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

	// Buscar ou criar carrinho para o usuário
	cartObj, err := getOrCreateCart(ctx, client, userId)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Erro ao buscar carrinho",
			"error":   err.Error(),
		})
	}

	// Verificar se o produto já está no carrinho
	existingItem, err := client.CartItem.
		Query().
		Where(
			cart_item.CartID(cartObj.ID),
			cart_item.ProductID(req.ProductID),
		).
		First(ctx)

	// Atualizar quantidade se já existir ou criar novo item
	var item *ent.CartItem
	if err == nil && existingItem != nil {
		// Atualizar quantidade
		item, err = client.CartItem.
			UpdateOne(existingItem).
			SetQuantity(existingItem.Quantity + req.Quantity).
			Save(ctx)
	} else {
		// Criar novo item
		item, err = client.CartItem.
			Create().
			SetID(uuid.New().String()).
			SetCartID(cartObj.ID).
			SetProductID(req.ProductID).
			SetName(prod.Name).
			SetPrice(prod.Price).
			SetImage(prod.Image).
			SetQuantity(req.Quantity).
			Save(ctx)
	}

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Erro ao adicionar item ao carrinho",
			"error":   err.Error(),
		})
	}

	// Atualizar totais do carrinho
	updatedCart, err := updateCartTotals(ctx, client, cartObj.ID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Item adicionado com sucesso, mas houve erro ao atualizar totais do carrinho",
			"error":   err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "Item adicionado ao carrinho com sucesso",
		"item":    item,
		"cart":    updatedCart,
	})
}

// UpdateCartItem atualiza a quantidade de um item do carrinho
// PUT /api/cart/items/:itemId
func UpdateCartItem(c fiber.Ctx) error {
	itemId := c.Params("itemId")
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
	var req CartItemRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Dados inválidos",
			"error":   err.Error(),
		})
	}

	// Validar quantidade positiva
	if req.Quantity <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "A quantidade deve ser maior que zero",
		})
	}

	// Buscar carrinho do usuário
	cartObj, err := client.Cart.
		Query().
		Where(cart.UserID(userId)).
		First(ctx)

	if err != nil {
		if ent.IsNotFound(err) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"message": "Carrinho não encontrado",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Erro ao buscar carrinho",
			"error":   err.Error(),
		})
	}

	// Verificar se o item existe e pertence ao carrinho do usuário
	item, err := client.CartItem.
		Query().
		Where(
			cart_item.ID(itemId),
			cart_item.HasCartWith(cart.ID(cartObj.ID)),
		).
		First(ctx)

	if err != nil {
		if ent.IsNotFound(err) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"message": "Item não encontrado no carrinho",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Erro ao buscar item do carrinho",
			"error":   err.Error(),
		})
	}

	// Atualizar quantidade do item
	updatedItem, err := client.CartItem.
		UpdateOne(item).
		SetQuantity(req.Quantity).
		Save(ctx)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Erro ao atualizar item do carrinho",
			"error":   err.Error(),
		})
	}

	// Atualizar totais do carrinho
	updatedCart, err := updateCartTotals(ctx, client, cartObj.ID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Item atualizado com sucesso, mas houve erro ao atualizar totais do carrinho",
			"error":   err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Item atualizado com sucesso",
		"item":    updatedItem,
		"cart":    updatedCart,
	})
}

// RemoveCartItem remove um item do carrinho
// DELETE /api/cart/items/:itemId
func RemoveCartItem(c fiber.Ctx) error {
	itemId := c.Params("itemId")
	client := c.Locals("dbClient").(*ent.Client)
	ctx := context.Background()
	
	// Obter usuário do contexto de autenticação
	userId := getUserIdFromContext(c)
	if userId == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "Usuário não autenticado",
		})
	}

	// Buscar carrinho do usuário
	cartObj, err := client.Cart.
		Query().
		Where(cart.UserID(userId)).
		First(ctx)

	if err != nil {
		if ent.IsNotFound(err) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"message": "Carrinho não encontrado",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Erro ao buscar carrinho",
			"error":   err.Error(),
		})
	}

	// Verificar se o item existe e pertence ao carrinho do usuário
	exists, err := client.CartItem.
		Query().
		Where(
			cart_item.ID(itemId),
			cart_item.HasCartWith(cart.ID(cartObj.ID)),
		).
		Exist(ctx)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Erro ao verificar item do carrinho",
			"error":   err.Error(),
		})
	}

	if !exists {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"message": "Item não encontrado no carrinho",
		})
	}

	// Remover o item
	err = client.CartItem.
		DeleteOneID(itemId).
		Exec(ctx)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Erro ao remover item do carrinho",
			"error":   err.Error(),
		})
	}

	// Atualizar totais do carrinho
	updatedCart, err := updateCartTotals(ctx, client, cartObj.ID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Item removido com sucesso, mas houve erro ao atualizar totais do carrinho",
			"error":   err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Item removido com sucesso",
		"cart":    updatedCart,
	})
}

// ApplyCoupon aplica um cupom de desconto ao carrinho
// POST /api/cart/coupon
func ApplyCoupon(c fiber.Ctx) error {
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
	var req CouponRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Dados inválidos",
			"error":   err.Error(),
		})
	}

	// Validar código do cupom
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
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Erro ao verificar cupom",
			"error":   err.Error(),
		})
	}

	// Verificar se o cupom expirou
	if couponObj.ExpiresAt != nil && couponObj.ExpiresAt.Before(time.Now()) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Cupom expirado",
		})
	}

	// Verificar número máximo de usos
	if couponObj.MaxUses != nil && *couponObj.MaxUses <= couponObj.TimesUsed {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Limite de uso do cupom excedido",
		})
	}

	// Buscar carrinho do usuário
	cartObj, err := client.Cart.
		Query().
		Where(cart.UserID(userId)).
		First(ctx)

	if err != nil {
		if ent.IsNotFound(err) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"message": "Carrinho não encontrado",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Erro ao buscar carrinho",
			"error":   err.Error(),
		})
	}

	// Verificar valor mínimo de compra
	if cartObj.Subtotal < couponObj.MinPurchase {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Valor mínimo para uso do cupom não atingido",
			"min_purchase": couponObj.MinPurchase,
		})
	}

	// Aplicar o cupom ao carrinho
	updateCart := client.Cart.
		UpdateOne(cartObj).
		SetAppliedCoupon(true).
		SetCouponCode(req.Code)

	// Calcular desconto
	var discount float64
	if couponObj.DiscountType == "percentage" {
		discount = cartObj.Subtotal * (couponObj.DiscountValue / 100)
	} else { // fixed
		discount = couponObj.DiscountValue
		if discount > cartObj.Subtotal {
			discount = cartObj.Subtotal
		}
	}

	// Atualizar desconto e total
	updateCart = updateCart.
		SetDiscount(discount).
		SetTotal(cartObj.Subtotal - discount + cartObj.Shipping)

	// Salvar atualização
	updatedCart, err := updateCart.Save(ctx)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Erro ao aplicar cupom ao carrinho",
			"error":   err.Error(),
		})
	}

	// Incrementar número de usos do cupom
	_, err = client.Coupon.
		UpdateOne(couponObj).
		SetTimesUsed(couponObj.TimesUsed + 1).
		Save(ctx)

	if err != nil {
		// Não falhar a operação, mas registrar o erro
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"message": "Cupom aplicado com sucesso, mas houve erro ao atualizar contador de usos",
			"cart":    updatedCart,
			"error":   err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Cupom aplicado com sucesso",
		"cart":    updatedCart,
	})
}

// RemoveCoupon remove o cupom de desconto do carrinho
// DELETE /api/cart/coupon
func RemoveCoupon(c fiber.Ctx) error {
	client := c.Locals("dbClient").(*ent.Client)
	ctx := context.Background()
	
	// Obter usuário do contexto de autenticação
	userId := getUserIdFromContext(c)
	if userId == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "Usuário não autenticado",
		})
	}

	// Buscar carrinho do usuário
	cartObj, err := client.Cart.
		Query().
		Where(cart.UserID(userId)).
		First(ctx)

	if err != nil {
		if ent.IsNotFound(err) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"message": "Carrinho não encontrado",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Erro ao buscar carrinho",
			"error":   err.Error(),
		})
	}

	// Verificar se há cupom aplicado
	if !cartObj.AppliedCoupon {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Não há cupom aplicado ao carrinho",
		})
	}

	// Armazenar código do cupom para decrementar uso
	couponCode := cartObj.CouponCode

	// Remover o cupom do carrinho
	updatedCart, err := client.Cart.
		UpdateOne(cartObj).
		SetAppliedCoupon(false).
		SetCouponCode("").
		SetDiscount(0).
		SetTotal(cartObj.Subtotal + cartObj.Shipping).
		Save(ctx)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Erro ao remover cupom do carrinho",
			"error":   err.Error(),
		})
	}

	// Decrementar número de usos do cupom
	if couponCode != "" {
		couponObj, err := client.Coupon.
			Query().
			Where(coupon.Code(couponCode)).
			First(ctx)

		if err == nil && couponObj != nil && couponObj.TimesUsed > 0 {
			_, err = client.Coupon.
				UpdateOne(couponObj).
				SetTimesUsed(couponObj.TimesUsed - 1).
				Save(ctx)

			if err != nil {
				// Não falhar a operação, mas registrar o erro
				return c.Status(fiber.StatusOK).JSON(fiber.Map{
					"message": "Cupom removido com sucesso, mas houve erro ao atualizar contador de usos",
					"cart":    updatedCart,
					"error":   err.Error(),
				})
			}
		}
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Cupom removido com sucesso",
		"cart":    updatedCart,
	})
}

// ClearCart remove todos os itens do carrinho
// DELETE /api/cart
func ClearCart(c fiber.Ctx) error {
	client := c.Locals("dbClient").(*ent.Client)
	ctx := context.Background()
	
	// Obter usuário do contexto de autenticação
	userId := getUserIdFromContext(c)
	if userId == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "Usuário não autenticado",
		})
	}

	// Buscar carrinho do usuário
	cartObj, err := client.Cart.
		Query().
		Where(cart.UserID(userId)).
		First(ctx)

	if err != nil {
		if ent.IsNotFound(err) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"message": "Carrinho não encontrado",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Erro ao buscar carrinho",
			"error":   err.Error(),
		})
	}

	// Remover todos os itens do carrinho
	_, err = client.CartItem.
		Delete().
		Where(cart_item.CartID(cartObj.ID)).
		Exec(ctx)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Erro ao remover itens do carrinho",
			"error":   err.Error(),
		})
	}

	// Remover cupom se estiver aplicado
	couponCode := ""
	if cartObj.AppliedCoupon {
		couponCode = cartObj.CouponCode
	}

	// Resetar o carrinho
	updatedCart, err := client.Cart.
		UpdateOne(cartObj).
		SetSubtotal(0).
		SetDiscount(0).
		SetTotal(cartObj.Shipping). // Mantém apenas o frete
		SetAppliedCoupon(false).
		SetCouponCode("").
		Save(ctx)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Erro ao resetar carrinho",
			"error":   err.Error(),
		})
	}

	// Decrementar número de usos do cupom se estava aplicado
	if couponCode != "" {
		couponObj, err := client.Coupon.
			Query().
			Where(coupon.Code(couponCode)).
			First(ctx)

		if err == nil && couponObj != nil && couponObj.TimesUsed > 0 {
			_, err = client.Coupon.
				UpdateOne(couponObj).
				SetTimesUsed(couponObj.TimesUsed - 1).
				Save(ctx)

			if err != nil {
				// Não falhar a operação, mas registrar o erro
				return c.Status(fiber.StatusOK).JSON(fiber.Map{
					"message": "Carrinho limpo com sucesso, mas houve erro ao atualizar contador de usos do cupom",
					"cart":    updatedCart,
					"error":   err.Error(),
				})
			}
		}
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Carrinho limpo com sucesso",
		"cart":    updatedCart,
	})
}

// Helper para buscar ou criar um carrinho para o usuário
func getOrCreateCart(ctx context.Context, client *ent.Client, userId string) (*ent.Cart, error) {
	// Buscar carrinho existente
	cartObj, err := client.Cart.
		Query().
		Where(cart.UserID(userId)).
		First(ctx)

	// Se encontrou, retornar
	if err == nil {
		return cartObj, nil
	}

	// Se o erro não for "não encontrado", retornar o erro
	if !ent.IsNotFound(err) {
		return nil, err
	}

	// Criar novo carrinho
	return client.Cart.
		Create().
		SetID(uuid.New().String()).
		SetUserID(userId).
		SetSubtotal(0).
		SetDiscount(0).
		SetTotal(0).
		SetAppliedCoupon(false).
		SetCreatedAt(time.Now()).
		SetUpdatedAt(time.Now()).
		Save(ctx)
}

// Helper para atualizar os totais do carrinho
func updateCartTotals(ctx context.Context, client *ent.Client, cartId string) (*ent.Cart, error) {
	// Buscar carrinho
	cartObj, err := client.Cart.Get(ctx, cartId)
	if err != nil {
		return nil, err
	}

	// Buscar todos os itens do carrinho
	items, err := client.CartItem.
		Query().
		Where(cart_item.CartID(cartId)).
		All(ctx)

	if err != nil {
		return nil, err
	}

	// Calcular subtotal
	var subtotal float64
	for _, item := range items {
		subtotal += item.Price * float64(item.Quantity)
	}

	// Calcular desconto (se tiver cupom aplicado)
	discount := 0.0
	if cartObj.AppliedCoupon && cartObj.CouponCode != "" {
		couponObj, err := client.Coupon.
			Query().
			Where(coupon.Code(cartObj.CouponCode)).
			First(ctx)

		if err == nil && couponObj != nil {
			if couponObj.DiscountType == "percentage" {
				discount = subtotal * (couponObj.DiscountValue / 100)
			} else { // fixed
				discount = couponObj.DiscountValue
				if discount > subtotal {
					discount = subtotal
				}
			}
		}
	}

	// Calcular total
	total := subtotal - discount
	if cartObj.Shipping > 0 {
		total += cartObj.Shipping
	}

	// Atualizar carrinho
	return client.Cart.
		UpdateOne(cartObj).
		SetSubtotal(subtotal).
		SetDiscount(discount).
		SetTotal(total).
		SetUpdatedAt(time.Now()).
		Save(ctx)
} 