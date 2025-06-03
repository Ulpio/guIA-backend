# Variáveis
APP_NAME=guia-backend
BINARY_NAME=main
DOCKER_IMAGE=guia/backend
VERSION?=latest
POSTGRES_CONTAINER=guia_postgres
BACKEND_CONTAINER=guia_backend

# Cores para output
RED=\033[0;31m
GREEN=\033[0;32m
YELLOW=\033[1;33m
BLUE=\033[0;34m
NC=\033[0m # No Color

.PHONY: help build run test clean docker-build docker-run docker-stop setup deps lint format migrate

# Comando padrão
help: ## Mostra este menu de ajuda
	@echo "$(BLUE)Comandos disponíveis para $(APP_NAME):$(NC)"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "$(YELLOW)%-20s$(NC) %s\n", $$1, $$2}'

# Comandos de desenvolvimento
deps: ## Instala dependências Go
	@echo "$(BLUE)Instalando dependências...$(NC)"
	go mod download
	go mod tidy

build: ## Compila a aplicação
	@echo "$(BLUE)Compilando aplicação...$(NC)"
	go build -o bin/$(BINARY_NAME) cmd/main.go
	@echo "$(GREEN)Aplicação compilada com sucesso!$(NC)"

run: ## Executa a aplicação localmente
	@echo "$(BLUE)Executando aplicação...$(NC)"
	go run cmd/main.go

dev: ## Executa com hot reload usando air
	@echo "$(BLUE)Executando em modo desenvolvimento...$(NC)"
	@if command -v air > /dev/null; then \
		air; \
	else \
		echo "$(RED)Air não encontrado. Instale com: go install github.com/cosmtrek/air@latest$(NC)"; \
		echo "$(YELLOW)Executando sem hot reload...$(NC)"; \
		go run cmd/main.go; \
	fi

test: ## Executa todos os testes
	@echo "$(BLUE)Executando testes...$(NC)"
	go test -v ./...

test-coverage: ## Executa testes com coverage
	@echo "$(BLUE)Executando testes com coverage...$(NC)"
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "$(GREEN)Coverage report gerado em coverage.html$(NC)"

benchmark: ## Executa benchmarks
	@echo "$(BLUE)Executando benchmarks...$(NC)"
	go test -bench=. -benchmem ./...

# Comandos de qualidade de código
lint: ## Executa linting do código
	@echo "$(BLUE)Executando linting...$(NC)"
	@if command -v golangci-lint > /dev/null; then \
		golangci-lint run; \
	else \
		echo "$(RED)golangci-lint não encontrado. Instale em: https://golangci-lint.run/usage/install/$(NC)"; \
	fi

format: ## Formata o código
	@echo "$(BLUE)Formatando código...$(NC)"
	go fmt ./...
	@if command -v goimports > /dev/null; then \
		goimports -w .; \
	else \
		echo "$(YELLOW)goimports não encontrado. Instale com: go install golang.org/x/tools/cmd/goimports@latest$(NC)"; \
	fi

vet: ## Executa go vet
	@echo "$(BLUE)Executando go vet...$(NC)"
	go vet ./...

# Comandos Docker
docker-build: ## Constrói a imagem Docker
	@echo "$(BLUE)Construindo imagem Docker...$(NC)"
	docker build -t $(DOCKER_IMAGE):$(VERSION) .
	@echo "$(GREEN)Imagem Docker construída: $(DOCKER_IMAGE):$(VERSION)$(NC)"

docker-run: ## Executa a aplicação via Docker Compose
	@echo "$(BLUE)Iniciando serviços com Docker Compose...$(NC)"
	docker-compose up -d
	@echo "$(GREEN)Serviços iniciados!$(NC)"
	@echo "$(YELLOW)Backend: http://localhost:8080$(NC)"
	@echo "$(YELLOW)Adminer: http://localhost:8081$(NC)"

docker-stop: ## Para todos os serviços Docker
	@echo "$(BLUE)Parando serviços Docker...$(NC)"
	docker-compose down

docker-logs: ## Mostra logs dos containers
	@echo "$(BLUE)Logs dos containers:$(NC)"
	docker-compose logs -f

docker-clean: ## Remove containers, imagens e volumes não utilizados
	@echo "$(BLUE)Limpando Docker...$(NC)"
	docker-compose down -v
	docker system prune -f
	@echo "$(GREEN)Docker limpo!$(NC)"

# Comandos de banco de dados
db-setup: ## Configura o banco de dados
	@echo "$(BLUE)Configurando banco de dados...$(NC)"
	docker-compose up -d postgres
	@echo "$(YELLOW)Aguardando PostgreSQL inicializar...$(NC)"
	@sleep 10
	@echo "$(GREEN)Banco de dados configurado!$(NC)"

db-migrate: ## Executa migrações do banco
	@echo "$(BLUE)Executando migrações...$(NC)"
	go run cmd/main.go migrate
	@echo "$(GREEN)Migrações executadas!$(NC)"

db-reset: ## Reseta o banco de dados
	@echo "$(RED)⚠️  ATENÇÃO: Isso irá deletar todos os dados!$(NC)"
	@read -p "Tem certeza? (y/N): " confirm && [ "$$confirm" = "y" ]
	docker-compose down postgres
	docker volume rm guia-backend_postgres_data 2>/dev/null || true
	docker-compose up -d postgres
	@echo "$(GREEN)Banco resetado!$(NC)"

# Comandos de setup
setup: ## Configura o ambiente de desenvolvimento
	@echo "$(BLUE)Configurando ambiente de desenvolvimento...$(NC)"
	@if [ ! -f .env ]; then \
		cp .env.example .env; \
		echo "$(YELLOW)Arquivo .env criado. Configure as variáveis necessárias.$(NC)"; \
	fi
	make deps
	make db-setup
	@echo "$(GREEN)Ambiente configurado!$(NC)"

install-tools: ## Instala ferramentas de desenvolvimento
	@echo "$(BLUE)Instalando ferramentas...$(NC)"
	go install github.com/cosmtrek/air@latest
	go install golang.org/x/tools/cmd/goimports@latest
	@echo "$(GREEN)Ferramentas instaladas!$(NC)"

# Comandos de deploy
build-prod: ## Constrói para produção
	@echo "$(BLUE)Construindo para produção...$(NC)"
	CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags '-extldflags "-static"' -o bin/$(BINARY_NAME) cmd/main.go
	@echo "$(GREEN)Build de produção concluído!$(NC)"

# Comandos de limpeza
clean: ## Remove arquivos de build
	@echo "$(BLUE)Limpando arquivos de build...$(NC)"
	rm -rf bin/
	rm -f coverage.out coverage.html
	go clean -testcache
	@echo "$(GREEN)Limpeza concluída!$(NC)"

clean-all: clean docker-clean ## Remove tudo (build files + Docker)

# Comandos de monitoramento
logs: ## Mostra logs da aplicação
	@echo "$(BLUE)Logs da aplicação:$(NC)"
	docker-compose logs -f backend

status: ## Mostra status dos serviços
	@echo "$(BLUE)Status dos serviços:$(NC)"
	docker-compose ps

# Comandos de utilitários
check: lint vet test ## Executa todas as verificações de qualidade

release: clean build test ## Prepara um release
	@echo "$(GREEN)Release preparado!$(NC)"

.DEFAULT_GOAL := help