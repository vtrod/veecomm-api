# VeeComm API

API de e-commerce completa desenvolvida em Go com GoFiber e Ent ORM.

## Requisitos

- Go 1.24+
- PostgreSQL
- Docker (opcional)

## Configuração

1. Clone o repositório
2. Crie um arquivo `.env` na raiz do projeto (ou use o fornecido)
3. Configure as variáveis de ambiente no arquivo `.env`
4. Execute `go mod download` para instalar dependências
5. Execute `go run main.go` para iniciar o servidor

## Variáveis de Ambiente

```
# Configurações do Servidor
PORT=8001

# Configurações de Banco de Dados
DATABASE_URL=postgres://postgres:postgres@localhost:5432/veecomm

# Configurações de Segurança
JWT_SECRET=seu_segredo_jwt
TOKEN_EXPIRY=24h
PASSWORD_SALT=seu_salt_para_senha

# Configurações CORS
CORS_ALLOW_ORIGINS=http://localhost:3000,http://localhost:5173
```

## Estrutura do Projeto

```
veecomm-api/
├── controllers/      # Controladores da API
├── database/         # Configuração do banco de dados
├── ent/              # Modelos e schema do Ent ORM
├── middleware/       # Middlewares (auth, etc)
├── routes/           # Definição de rotas
├── .env              # Variáveis de ambiente
├── go.mod            # Dependências
├── go.sum            # Checksums das dependências
└── main.go           # Ponto de entrada da aplicação
```

## Endpoints da API

### Autenticação

- `POST /api/auth/login` - Login de usuário
- `POST /api/auth/register` - Registrar novo usuário

### Usuários

- `GET /api/users/profile` - Obter perfil do usuário
- `PUT /api/users/profile` - Atualizar perfil do usuário

### Produtos

- `GET /api/products` - Listar todos os produtos
- `GET /api/products/:id` - Obter detalhes de um produto
- `GET /api/products/category/:categoryId` - Listar produtos por categoria
- `GET /api/products/promotions` - Listar produtos em promoção
- `POST /api/products` - Criar novo produto (admin)
- `PUT /api/products/:id` - Atualizar produto (admin)
- `DELETE /api/products/:id` - Deletar produto (admin)

### Categorias

- `GET /api/categories` - Listar todas as categorias
- `GET /api/categories/:id` - Obter detalhes de uma categoria
- `POST /api/categories` - Criar nova categoria (admin)
- `PUT /api/categories/:id` - Atualizar categoria (admin)
- `DELETE /api/categories/:id` - Deletar categoria (admin)

### Carrinho

- `GET /api/cart` - Obter itens do carrinho
- `POST /api/cart/items` - Adicionar item ao carrinho
- `PUT /api/cart/items/:itemId` - Atualizar quantidade de item
- `DELETE /api/cart/items/:itemId` - Remover item do carrinho
- `POST /api/cart/coupon` - Aplicar cupom de desconto
- `DELETE /api/cart/coupon` - Remover cupom de desconto
- `DELETE /api/cart` - Limpar carrinho

### Pedidos

- `GET /api/orders` - Listar pedidos do usuário
- `GET /api/orders/:id` - Obter detalhes de um pedido
- `POST /api/orders` - Criar novo pedido
- `PUT /api/orders/:id/status` - Atualizar status do pedido (admin)
- `DELETE /api/orders/:id` - Cancelar/deletar pedido

### Avaliações

- `GET /api/products/:productId/reviews` - Listar avaliações de um produto
- `POST /api/products/:productId/reviews` - Adicionar avaliação a um produto
- `PUT /api/products/:productId/reviews/:reviewId` - Atualizar avaliação
- `DELETE /api/products/:productId/reviews/:reviewId` - Deletar avaliação

### Endereços

- `GET /api/addresses` - Listar endereços do usuário
- `GET /api/addresses/:id` - Obter detalhes de um endereço
- `POST /api/addresses` - Criar novo endereço
- `PUT /api/addresses/:id` - Atualizar endereço
- `PUT /api/addresses/:id/default` - Definir endereço como padrão
- `DELETE /api/addresses/:id` - Deletar endereço

### Frete e Entrega

- `POST /api/shipping/calculate` - Calcular custo de frete
- `GET /api/shipping/:orderId/track` - Rastrear entrega

### Cupons

- `GET /api/coupons` - Listar todos os cupons (admin)
- `GET /api/coupons/:id` - Obter detalhes de um cupom (admin)
- `POST /api/coupons/validate` - Validar cupom
- `POST /api/coupons` - Criar novo cupom (admin)
- `PUT /api/coupons/:id` - Atualizar cupom (admin)
- `DELETE /api/coupons/:id` - Deletar cupom (admin)

### Administração

- `GET /api/admin/dashboard` - Obter dados do dashboard
- `GET /api/admin/users` - Listar todos os usuários
- `GET /api/admin/orders` - Listar todos os pedidos

## Autenticação

A API utiliza JWT (JSON Web Token) para autenticação. Para acessar endpoints protegidos, inclua o token no cabeçalho da requisição:

```
Authorization: Bearer seu_token_jwt
```

## Instalação com Docker

1. Certifique-se de que o Docker e Docker Compose estão instalados
2. Execute `docker-compose up -d` para iniciar o servidor e o banco de dados
3. A API estará disponível em `http://localhost:8001`

## Tecnologias Utilizadas

- [Go](https://golang.org/)
- [Fiber](https://gofiber.io/) - Framework web
- [Ent](https://entgo.io/) - ORM
- [JWT](https://github.com/golang-jwt/jwt) - Autenticação
- [PostgreSQL](https://www.postgresql.org/) - Banco de dados 