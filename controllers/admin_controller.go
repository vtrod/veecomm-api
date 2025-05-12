package controllers

import (
	"context"
	"time"
	"github.com/vtrod/veecomm-api/ent"
	"github.com/vtrod/veecomm-api/ent/order"
	"github.com/vtrod/veecomm-api/ent/user"

	"github.com/gofiber/fiber/v3"
)

// DashboardData contém estatísticas para o dashboard administrativo
type DashboardData struct {
	Revenue          float64 `json:"revenue"`
	Orders           int     `json:"orders"`
	Users            int     `json:"users"`
	Products         int     `json:"products"`
	RecentOrders     []Order `json:"recentOrders"`
	TopSellingItems  []Item  `json:"topSellingItems"`
	MonthlyRevenue   []float64 `json:"monthlyRevenue"`
}

// Item representa um produto com suas estatísticas de vendas
type Item struct {
	ID          int     `json:"id"`
	Name        string  `json:"name"`
	Quantity    int     `json:"quantity"`
	TotalSales  float64 `json:"totalSales"`
}

// GetDashboardData retorna dados estatísticos para o dashboard administrativo
func GetDashboardData(c fiber.Ctx) error {
	client := c.Locals("dbClient").(*ent.Client)
	ctx := context.Background()

	// Período de análise - último mês
	endDate := time.Now()
	startDate := endDate.AddDate(0, -1, 0)

	// 1. Obter receita total
	var revenue float64
	err := client.Order.Query().
		Where(order.StatusIn(
			order.StatusCompleted,
			order.StatusProcessed,
			order.StatusShipped,
		)).
		Aggregate(
			ent.Sum(order.FieldTotalAmount),
		).
		Scan(ctx, &revenue)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Erro ao calcular receita total",
			"error":   err.Error(),
		})
	}

	// 2. Contagem de pedidos
	orderCount, err := client.Order.Query().Count(ctx)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Erro ao contar pedidos",
			"error":   err.Error(),
		})
	}

	// 3. Contagem de usuários
	userCount, err := client.User.Query().Count(ctx)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Erro ao contar usuários",
			"error":   err.Error(),
		})
	}

	// 4. Contagem de produtos
	productCount, err := client.Product.Query().Count(ctx)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Erro ao contar produtos",
			"error":   err.Error(),
		})
	}

	// 5. Pedidos recentes
	recentOrders, err := client.Order.Query().
		Order(ent.Desc(order.FieldCreatedAt)).
		Limit(5).
		WithOrderItems().
		WithUser().
		All(ctx)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Erro ao obter pedidos recentes",
			"error":   err.Error(),
		})
	}

	// Converter para resposta formatada
	formattedOrders := make([]Order, 0, len(recentOrders))
	for _, o := range recentOrders {
		formattedOrders = append(formattedOrders, formatOrder(o))
	}

	// 6. Receita mensal (últimos 12 meses)
	monthlyRevenue := make([]float64, 12)
	currentMonth := endDate.Month()
	currentYear := endDate.Year()

	// Preencher com dados dos últimos 12 meses
	for i := 0; i < 12; i++ {
		month := int(currentMonth) - i
		year := currentYear
		
		if month <= 0 {
			month += 12
			year--
		}
		
		startOfMonth := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
		endOfMonth := startOfMonth.AddDate(0, 1, 0).Add(-time.Second)
		
		var monthRevenue float64
		err := client.Order.Query().
			Where(
				order.StatusIn(
					order.StatusCompleted,
					order.StatusProcessed,
					order.StatusShipped,
				),
				order.CreatedAtGTE(startOfMonth),
				order.CreatedAtLTE(endOfMonth),
			).
			Aggregate(
				ent.Sum(order.FieldTotalAmount),
			).
			Scan(ctx, &monthRevenue)
		
		if err == nil {
			// Armazenar no índice correto (reverso)
			monthlyRevenue[11-i] = monthRevenue
		}
	}

	// Montar resposta
	dashboardData := DashboardData{
		Revenue:         revenue,
		Orders:          orderCount,
		Users:           userCount,
		Products:        productCount,
		RecentOrders:    formattedOrders,
		MonthlyRevenue:  monthlyRevenue,
	}

	return c.JSON(dashboardData)
}

// GetAllUsers retorna todos os usuários (apenas admin)
func GetAllUsers(c fiber.Ctx) error {
	client := c.Locals("dbClient").(*ent.Client)
	ctx := context.Background()

	// Opcionalmente, aplicar filtros e paginação
	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 20)
	offset := (page - 1) * limit

	// Consulta total de usuários para paginação
	total, err := client.User.Query().Count(ctx)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Erro ao contar usuários",
			"error":   err.Error(),
		})
	}

	// Consulta usuários com paginação
	users, err := client.User.Query().
		Order(ent.Asc(user.FieldID)).
		Offset(offset).
		Limit(limit).
		All(ctx)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Erro ao buscar usuários",
			"error":   err.Error(),
		})
	}

	// Formatando resposta - removendo campos sensíveis
	formattedUsers := make([]fiber.Map, 0, len(users))
	for _, u := range users {
		formattedUsers = append(formattedUsers, fiber.Map{
			"id":        u.ID,
			"name":      u.Name,
			"email":     u.Email,
			"isAdmin":   u.IsAdmin,
			"createdAt": u.CreatedAt.Format(time.RFC3339),
		})
	}

	return c.JSON(fiber.Map{
		"users": formattedUsers,
		"pagination": fiber.Map{
			"total":  total,
			"page":   page,
			"limit":  limit,
			"pages":  (total + limit - 1) / limit,
		},
	})
}

// GetAllOrders retorna todos os pedidos (apenas admin)
func GetAllOrders(c fiber.Ctx) error {
	client := c.Locals("dbClient").(*ent.Client)
	ctx := context.Background()

	// Opcionalmente, aplicar filtros e paginação
	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 20)
	offset := (page - 1) * limit
	status := c.Query("status", "")

	// Construir consulta com filtros opcionais
	query := client.Order.Query()
	if status != "" {
		query = query.Where(order.Status(order.Status(status)))
	}

	// Consulta total de pedidos para paginação
	total, err := query.Clone().Count(ctx)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Erro ao contar pedidos",
			"error":   err.Error(),
		})
	}

	// Consulta pedidos com paginação
	orders, err := query.
		Order(ent.Desc(order.FieldCreatedAt)).
		Offset(offset).
		Limit(limit).
		WithUser().
		WithOrderItems().
		All(ctx)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Erro ao buscar pedidos",
			"error":   err.Error(),
		})
	}

	// Formatando resposta
	formattedOrders := make([]Order, 0, len(orders))
	for _, o := range orders {
		formattedOrders = append(formattedOrders, formatOrder(o))
	}

	return c.JSON(fiber.Map{
		"orders": formattedOrders,
		"pagination": fiber.Map{
			"total":  total,
			"page":   page,
			"limit":  limit,
			"pages":  (total + limit - 1) / limit,
		},
	})
} 