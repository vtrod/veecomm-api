package controllers

import (
	"context"
	"github.com/vtrod/veecomm-api/ent"
	"github.com/vtrod/veecomm-api/ent/order"
	"github.com/vtrod/veecomm-api/ent/address"
	"github.com/vtrod/veecomm-api/ent/cart"
	"github.com/vtrod/veecomm-api/ent/cart_item"
	"github.com/vtrod/veecomm-api/ent/product"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

// Estrutura para criar um novo pedido
type OrderRequest struct {
	AddressID      string `json:"address_id"`
	DeliveryType   string `json:"delivery_type"`
	PaymentMethod  string `json:"payment_method"`
	PaymentStatus  string `json:"payment_status"`
}

// Estrutura para atualizar status de um pedido
type OrderStatusUpdate struct {
	Status string `json:"status"`
}

// GetUserOrders retorna todos os pedidos do usuário
// GET /api/orders
func GetUserOrders(c fiber.Ctx) error {
	client := c.Locals("dbClient").(*ent.Client)
	ctx := context.Background()
	
	// Obter usuário do contexto de autenticação
	userId := getUserIdFromContext(c)
	if userId == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "Usuário não autenticado",
		})
	}

	// Buscar pedidos do usuário
	orders, err := client.Order.
		Query().
		Where(order.UserID(userId)).
		Order(ent.Desc("date")).
		All(ctx)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Erro ao buscar pedidos",
			"error":   err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"orders": orders,
	})
}

// GetOrder retorna os detalhes de um pedido específico
// GET /api/orders/:id
func GetOrder(c fiber.Ctx) error {
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

	// Verificar se o pedido existe e pertence ao usuário
	orderObj, err := client.Order.
		Query().
		Where(
			order.ID(id),
			order.UserID(userId),
		).
		WithAddress().
		First(ctx)

	if err != nil {
		if ent.IsNotFound(err) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"message": "Pedido não encontrado",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Erro ao buscar pedido",
			"error":   err.Error(),
		})
	}

	// Buscar itens do pedido
	items, err := client.OrderItem.
		Query().
		Where(ent.HasOrderWith(order.ID(id))).
		All(ctx)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Erro ao buscar itens do pedido",
			"error":   err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"order": orderObj,
		"items": items,
	})
}

// CreateOrder cria um novo pedido a partir do carrinho
// POST /api/orders
func CreateOrder(c fiber.Ctx) error {
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
	var req OrderRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Dados inválidos",
			"error":   err.Error(),
		})
	}

	// Validar tipo de entrega
	if req.DeliveryType != "pickup" && req.DeliveryType != "delivery" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Tipo de entrega inválido. Use 'pickup' ou 'delivery'",
		})
	}

	// Validar método de pagamento
	if req.PaymentMethod == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Método de pagamento é obrigatório",
		})
	}

	// Validar status de pagamento
	if req.PaymentStatus == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Status de pagamento é obrigatório",
		})
	}

	// Validar endereço de entrega para delivery
	if req.DeliveryType == "delivery" {
		if req.AddressID == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"message": "Endereço de entrega é obrigatório para entregas",
			})
		}

		// Verificar se o endereço existe e pertence ao usuário
		exists, err := client.Address.
			Query().
			Where(
				address.ID(req.AddressID),
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

	// Buscar itens do carrinho
	cartItems, err := client.CartItem.
		Query().
		Where(cart_item.CartID(cartObj.ID)).
		All(ctx)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Erro ao buscar itens do carrinho",
			"error":   err.Error(),
		})
	}

	// Verificar se há itens no carrinho
	if len(cartItems) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Não há itens no carrinho para criar o pedido",
		})
	}

	// Criar o pedido
	orderId := uuid.New().String()
	orderBuilder := client.Order.
		Create().
		SetID(orderId).
		SetUserID(userId).
		SetDate(time.Now()).
		SetTotal(cartObj.Total).
		SetShipping(cartObj.Shipping).
		SetDiscount(cartObj.Discount).
		SetDeliveryType(req.DeliveryType).
		SetStatus("pending").
		SetPaymentMethod(req.PaymentMethod).
		SetPaymentStatus(req.PaymentStatus)

	// Adicionar endereço se for delivery
	if req.DeliveryType == "delivery" && req.AddressID != "" {
		orderBuilder = orderBuilder.SetAddressID(req.AddressID)
	}

	// Adicionar cupom se estiver aplicado
	if cartObj.AppliedCoupon && cartObj.CouponCode != "" {
		orderBuilder = orderBuilder.SetCouponCode(cartObj.CouponCode)
	}

	// Salvar o pedido
	orderObj, err := orderBuilder.Save(ctx)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Erro ao criar pedido",
			"error":   err.Error(),
		})
	}

	// Criar itens do pedido a partir dos itens do carrinho
	for _, item := range cartItems {
		// Buscar produto para pegar nome atualizado
		prod, err := client.Product.Get(ctx, item.ProductID)
		if err != nil && !ent.IsNotFound(err) {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"message": "Erro ao buscar produto",
				"error":   err.Error(),
			})
		}

		// Definir nome do produto (usar nome do carrinho se produto não existir mais)
		prodName := item.Name
		if err == nil && prod != nil {
			prodName = prod.Name
		}

		// Criar item do pedido
		_, err = client.OrderItem.
			Create().
			SetID(uuid.New().String()).
			SetOrderID(orderId).
			SetProductID(item.ProductID).
			SetName(prodName).
			SetQuantity(item.Quantity).
			SetPrice(item.Price).
			Save(ctx)

		if err != nil {
			// Se falhar, continuar tentando os outros itens
			// No final, o usuário terá um pedido incompleto, mas é melhor que nenhum
			continue
		}
	}

	// Limpar o carrinho
	_, err = client.CartItem.
		Delete().
		Where(cart_item.CartID(cartObj.ID)).
		Exec(ctx)

	if err != nil {
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"message": "Pedido criado com sucesso, mas houve erro ao limpar o carrinho",
			"order":   orderObj,
			"error":   err.Error(),
		})
	}

	// Resetar o carrinho
	_, err = client.Cart.
		UpdateOne(cartObj).
		SetSubtotal(0).
		SetDiscount(0).
		SetTotal(0).
		SetAppliedCoupon(false).
		SetCouponCode("").
		Save(ctx)

	if err != nil {
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"message": "Pedido criado com sucesso, mas houve erro ao resetar o carrinho",
			"order":   orderObj,
			"error":   err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "Pedido criado com sucesso",
		"order":   orderObj,
	})
}

// UpdateOrderStatus atualiza o status de um pedido
// PUT /api/orders/:id/status
func UpdateOrderStatus(c fiber.Ctx) error {
	id := c.Params("id")
	client := c.Locals("dbClient").(*ent.Client)
	ctx := context.Background()
	
	// Verificar se o usuário é admin
	isAdmin := c.Locals("isAdmin").(bool)
	if !isAdmin {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"message": "Apenas administradores podem atualizar o status de pedidos",
		})
	}

	// Extrair dados do request
	var req OrderStatusUpdate
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Dados inválidos",
			"error":   err.Error(),
		})
	}

	// Validar status
	validStatus := map[string]bool{
		"pending":    true,
		"processing": true,
		"shipped":    true,
		"delivered":  true,
		"cancelled":  true,
	}

	if !validStatus[req.Status] {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Status inválido",
			"valid_status": []string{
				"pending", "processing", "shipped", "delivered", "cancelled",
			},
		})
	}

	// Verificar se o pedido existe
	exists, err := client.Order.
		Query().
		Where(order.ID(id)).
		Exist(ctx)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Erro ao verificar pedido",
			"error":   err.Error(),
		})
	}

	if !exists {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"message": "Pedido não encontrado",
		})
	}

	// Atualizar status do pedido
	updatedOrder, err := client.Order.
		UpdateOneID(id).
		SetStatus(req.Status).
		Save(ctx)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Erro ao atualizar status do pedido",
			"error":   err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Status do pedido atualizado com sucesso",
		"order":   updatedOrder,
	})
}

// CancelOrder cancela um pedido
// DELETE /api/orders/:id
func CancelOrder(c fiber.Ctx) error {
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

	// Verificar se o pedido existe e pertence ao usuário
	orderObj, err := client.Order.
		Query().
		Where(
			order.ID(id),
			order.UserID(userId),
		).
		First(ctx)

	if err != nil {
		if ent.IsNotFound(err) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"message": "Pedido não encontrado",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Erro ao buscar pedido",
			"error":   err.Error(),
		})
	}

	// Verificar se o pedido pode ser cancelado
	if orderObj.Status == "delivered" || orderObj.Status == "cancelled" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Não é possível cancelar um pedido entregue ou já cancelado",
		})
	}

	// Apenas cancelar o pedido (não excluir)
	updatedOrder, err := client.Order.
		UpdateOne(orderObj).
		SetStatus("cancelled").
		Save(ctx)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Erro ao cancelar pedido",
			"error":   err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Pedido cancelado com sucesso",
		"order":   updatedOrder,
	})
} 