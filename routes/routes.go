package routes

import (
	"github.com/vtrod/veecomm-api/controllers"
	"github.com/vtrod/veecomm-api/middleware"

	"github.com/gofiber/fiber/v3"
)

// SetupRoutes configura todas as rotas da aplicação
func SetupRoutes(app *fiber.App) {
	// API grupo de rotas
	api := app.Group("/api")
	
	// 1. Rotas de Produtos (Products)
	products := api.Group("/products")
	products.Get("/", controllers.GetAllProducts)                     // Listar todos os produtos
	products.Get("/:id", controllers.GetProduct)                      // Obter detalhes de um produto
	products.Get("/category/:categoryId", controllers.GetProductsByCategory) // Listar produtos por categoria
	products.Get("/promotions", controllers.GetPromotionProducts)     // Listar produtos em promoção
	products.Post("/", middleware.Protected, middleware.AdminOnly, controllers.CreateProduct)                     // Criar novo produto
	products.Put("/:id", middleware.Protected, middleware.AdminOnly, controllers.UpdateProduct)                   // Atualizar produto
	products.Delete("/:id", middleware.Protected, middleware.AdminOnly, controllers.DeleteProduct)                // Deletar produto

	// 2. Rotas de Categorias (Categories)
	categories := api.Group("/categories")
	categories.Get("/", controllers.GetAllCategories)                 // Listar todas as categorias
	categories.Get("/:id", controllers.GetCategory)                   // Obter detalhes de uma categoria
	categories.Post("/", middleware.Protected, middleware.AdminOnly, controllers.CreateCategory)                  // Criar nova categoria
	categories.Put("/:id", middleware.Protected, middleware.AdminOnly, controllers.UpdateCategory)                // Atualizar categoria
	categories.Delete("/:id", middleware.Protected, middleware.AdminOnly, controllers.DeleteCategory)             // Deletar categoria

	// 3. Rotas de Carrinho (Cart)
	cart := api.Group("/cart", middleware.Protected)
	cart.Get("/", controllers.GetCart)                                // Obter itens do carrinho
	cart.Post("/items", controllers.AddToCart)                        // Adicionar item ao carrinho
	cart.Put("/items/:itemId", controllers.UpdateCartItem)            // Atualizar quantidade de item
	cart.Delete("/items/:itemId", controllers.RemoveCartItem)         // Remover item do carrinho
	cart.Post("/coupon", controllers.ApplyCoupon)                     // Aplicar cupom de desconto
	cart.Delete("/coupon", controllers.RemoveCoupon)                  // Remover cupom de desconto
	cart.Delete("/", controllers.ClearCart)                           // Limpar carrinho

	// 4. Rotas de Pedidos (Orders)
	orders := api.Group("/orders", middleware.Protected)
	orders.Get("/", controllers.GetUserOrders)                        // Listar pedidos do usuário
	orders.Get("/:id", controllers.GetOrder)                          // Obter detalhes de um pedido
	orders.Post("/", controllers.CreateOrder)                         // Criar novo pedido
	orders.Put("/:id/status", middleware.AdminOnly, controllers.UpdateOrderStatus)          // Atualizar status do pedido
	orders.Delete("/:id", controllers.CancelOrder)                    // Cancelar/deletar pedido

	// 5. Rotas de Avaliações (Reviews)
	api.Get("/products/:productId/reviews", controllers.GetProductReviews)          // Listar avaliações de um produto
	api.Post("/products/:productId/reviews", middleware.Protected, controllers.AddProductReview)          // Adicionar avaliação a um produto
	api.Put("/products/:productId/reviews/:reviewId", middleware.Protected, controllers.UpdateProductReview) // Atualizar avaliação
	api.Delete("/products/:productId/reviews/:reviewId", middleware.Protected, controllers.DeleteProductReview) // Deletar avaliação

	// 6. Rotas de Endereços (Addresses)
	addresses := api.Group("/addresses", middleware.Protected)
	addresses.Get("/", controllers.GetUserAddresses)                  // Listar endereços do usuário
	addresses.Get("/:id", controllers.GetAddress)                     // Obter detalhes de um endereço
	addresses.Post("/", controllers.CreateAddress)                    // Criar novo endereço
	addresses.Put("/:id", controllers.UpdateAddress)                  // Atualizar endereço
	addresses.Put("/:id/default", controllers.SetDefaultAddress)      // Definir endereço como padrão
	addresses.Delete("/:id", controllers.DeleteAddress)               // Deletar endereço

	// 7. Rotas de Usuários (Users)
	auth := api.Group("/auth")
	auth.Post("/login", controllers.LoginUser)                        // Login de usuário
	auth.Post("/register", controllers.RegisterUser)                  // Registrar novo usuário
	
	users := api.Group("/users", middleware.Protected)
	users.Get("/profile", controllers.GetUserProfile)                 // Obter perfil do usuário
	users.Put("/profile", controllers.UpdateUserProfile)              // Atualizar perfil do usuário

	// 8. Rotas de Frete e Entrega (Shipping)
	shipping := api.Group("/shipping")
	shipping.Post("/calculate", controllers.CalculateShipping)        // Calcular custo de frete
	shipping.Get("/:orderId/track", middleware.Protected, controllers.TrackShipping)        // Rastrear entrega
	
	// 9. Rotas de Cupons (Coupons)
	coupons := api.Group("/coupons")
	coupons.Get("/", middleware.Protected, middleware.AdminOnly, controllers.GetAllCoupons)                       // Listar todos os cupons (admin)
	coupons.Get("/:id", middleware.Protected, middleware.AdminOnly, controllers.GetCoupon)                        // Obter detalhes de um cupom (admin)
	coupons.Post("/validate", controllers.ValidateCoupon)             // Validar cupom
	coupons.Post("/", middleware.Protected, middleware.AdminOnly, controllers.CreateCoupon)                       // Criar novo cupom (admin)
	coupons.Put("/:id", middleware.Protected, middleware.AdminOnly, controllers.UpdateCoupon)                     // Atualizar cupom (admin)
	coupons.Delete("/:id", middleware.Protected, middleware.AdminOnly, controllers.DeleteCoupon)                  // Deletar cupom (admin)

	// 10. Rota de Administração (Dashboard)
	admin := api.Group("/admin", middleware.Protected, middleware.AdminOnly)
	admin.Get("/dashboard", controllers.GetDashboardData)             // Obter dados do dashboard
	admin.Get("/users", controllers.GetAllUsers)                      // Listar todos os usuários
	admin.Get("/orders", controllers.GetAllOrders)                    // Listar todos os pedidos
} 