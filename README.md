# guIA Backend

Uma plataforma de rede social voltada para viajantes e autônomos que gostem de viagem, onde os usuários podem compartilhar experiências, criar roteiros detalhados e descobrir novos destinos.

## 🚀 Funcionalidades

### Versão Atual (MVP)
- ✅ **Autenticação JWT** - Registro, login e gerenciamento de sessões
- ✅ **Perfis de Usuário** - Perfis pessoais e empresariais com verificação
- ✅ **Upload de Mídia** - Upload de imagens e vídeos para posts
- ✅ **Posts Sociais** - Compartilhamento de texto, imagens e vídeos
- ✅ **Sistema de Curtidas** - Interação com posts
- ✅ **Sistema de Seguidores** - Seguir/deixar de seguir usuários
- ✅ **Roteiros de Viagem** - Criação e gerenciamento de itinerários detalhados
- ✅ **Avaliações** - Sistema de ratings para roteiros
- ✅ **Busca** - Busca de usuários, posts e roteiros
- ✅ **Categorização** - Roteiros organizados por categorias (aventura, cultural, gastronômico, etc.)
- ✅ **Geolocalização** - Suporte a coordenadas GPS em posts e roteiros

### Funcionalidades Futuras
- 🔄 **IA para Recomendações** - Sugestões personalizadas de roteiros
- 🔄 **Parcerias Empresariais** - Roteiros corporativos (iFood, XP Investimentos, etc.)
- 🔄 **Chat e Mensagens** - Sistema de mensagens privadas
- 🔄 **Comentários** - Sistema de comentários em posts e roteiros
- 🔄 **Notificações** - Sistema de notificações em tempo real
- 🔄 **Processamento de Imagem** - Redimensionamento e otimização automática

## 🛠 Tecnologias

- **Backend**: Go 1.21+ com Gin Framework
- **Banco de Dados**: PostgreSQL 15+
- **ORM**: GORM
- **Autenticação**: JWT (golang-jwt/jwt)
- **Containerização**: Docker & Docker Compose
- **Cache**: Redis (configurado para uso futuro)

## 📋 Pré-requisitos

- Go 1.21 ou superior
- PostgreSQL 15 ou superior
- Docker e Docker Compose (opcional)
- Make (opcional, mas recomendado)

## 🚀 Instalação

### Opção 1: Docker (Recomendado)

```bash
# Clone o repositório
git clone https://github.com/Ulpio/guIA-backend.git
cd guIA-backend

# Configure o ambiente
make setup

# Inicie os serviços
make docker-run
```

### Opção 2: Instalação Local

```bash
# Clone o repositório
git clone https://github.com/Ulpio/guIA-backend.git
cd guIA-backend

# Instale as dependências
make deps

# Configure o arquivo .env
cp .env.example .env
# Edite o arquivo .env com suas configurações

# Configure o banco de dados PostgreSQL
# Crie um banco chamado 'guia_db'

# Execute a aplicação
make run
```

## ⚙️ Configuração

### Variáveis de Ambiente

Copie o arquivo `.env.example` para `.env` e configure as seguintes variáveis:

```env
# Servidor
PORT=8080
ENVIRONMENT=development

# Banco de Dados
DATABASE_URL=postgres://username:password@localhost:5432/guia_db?sslmode=disable

# JWT
JWT_SECRET=your-super-secret-jwt-key-change-this-in-production
```

### Banco de Dados

As migrações são executadas automaticamente ao iniciar a aplicação. Os seguintes modelos são criados:

- `users` - Usuários da plataforma
- `posts` - Posts dos usuários
- `post_likes` - Curtidas nos posts
- `comments` - Comentários (preparado para implementação futura)
- `itineraries` - Roteiros de viagem
- `itinerary_days` - Dias dos roteiros
- `itinerary_locations` - Locais dos roteiros
- `itinerary_ratings` - Avaliações dos roteiros
- `follows` - Relacionamentos de seguidor

## 📚 API Documentation

### Autenticação

#### Registro
```http
POST /api/v1/auth/register
Content-Type: application/json

{
  "username": "joao123",
  "email": "joao@example.com",
  "password": "senha123",
  "first_name": "João",
  "last_name": "Silva",
  "user_type": "normal"
}
```

#### Login
```http
POST /api/v1/auth/login
Content-Type: application/json

{
  "login": "joao@example.com",
  "password": "senha123"
}
```

### Posts

#### Criar Post
```http
POST /api/v1/posts
Authorization: Bearer {token}
Content-Type: application/json

{
  "content": "Que viagem incrível para o Rio!",
  "post_type": "text",
  "location": "Rio de Janeiro, RJ",
  "latitude": -22.9068,
  "longitude": -43.1729
}
```

#### Feed
```http
GET /api/v1/posts?limit=20&offset=0
Authorization: Bearer {token}
```

### Roteiros

#### Criar Roteiro
```http
POST /api/v1/itineraries
Authorization: Bearer {token}
Content-Type: application/json

{
  "title": "3 Dias no Rio de Janeiro",
  "description": "Roteiro completo para conhecer o Rio",
  "category": "urban",
  "duration": 3,
  "country": "Brasil",
  "city": "Rio de Janeiro",
  "estimated_cost": 1500.00,
  "currency": "BRL",
  "is_public": true
}
```

### Usuários

#### Perfil
```http
GET /api/v1/users/profile
Authorization: Bearer {token}
```

#### Seguir Usuário
```http
POST /api/v1/users/{id}/follow
Authorization: Bearer {token}
```

### Upload de Mídia

#### Upload de Imagem
```http
POST /api/v1/media/upload/image
Authorization: Bearer {token}
Content-Type: multipart/form-data

file: [arquivo_imagem.jpg]
```

#### Upload de Vídeo
```http
POST /api/v1/media/upload/video
Authorization: Bearer {token}
Content-Type: multipart/form-data

file: [arquivo_video.mp4]
```

#### Upload Múltiplo
```http
POST /api/v1/media/upload/multiple
Authorization: Bearer {token}
Content-Type: multipart/form-data

files: [arquivo1.jpg, arquivo2.png, video.mp4]
type: image (opcional - filtra apenas imagens)
```

#### Criar Post com Mídia
```http
POST /api/v1/posts
Authorization: Bearer {token}
Content-Type: application/json

{
  "content": "Confira essas fotos da viagem!",
  "post_type": "image",
  "media_urls": [
    "http://localhost:8080/uploads/images/123_1640995200_abc12345.jpg",
    "http://localhost:8080/uploads/images/123_1640995201_def67890.jpg"
  ],
  "location": "Rio de Janeiro, RJ"
}
```

## 🏗 Arquitetura

O projeto segue os princípios da Clean Architecture:

```
cmd/
├── main.go                 # Ponto de entrada da aplicação

internal/
├── config/                 # Configurações
├── database/              # Conexão e migrações do banco
├── handlers/              # Controllers HTTP
├── middleware/            # Middlewares (auth, cors, etc.)
├── models/               # Modelos de dados (structs)
├── repositories/         # Camada de acesso aos dados
└── services/             # Lógica de negócio

pkg/                      # Pacotes reutilizáveis (futuro)
```

### Fluxo de Dados

```
HTTP Request → Handler → Service → Repository → Database
                  ↓
HTTP Response ← Handler ← Service ← Repository ← Database
```

## 🔧 Comandos Úteis

```bash
# Desenvolvimento
make run              # Executa a aplicação
make dev              # Executa com hot reload (air)
make test             # Executa testes
make lint             # Executa linting

# Docker
make docker-run       # Inicia com Docker Compose
make docker-stop      # Para os serviços
make docker-logs      # Mostra logs

# Banco de dados
make db-setup         # Configura o banco
make db-reset         # Reseta o banco (⚠️ dados perdidos)

# Build
make build            # Compila a aplicação
make build-prod       # Build otimizado para produção

# Limpeza
make clean            # Remove arquivos de build
make clean-all        # Remove tudo (build + Docker)
```

## 🧪 Testes

```bash
# Executar todos os testes
make test

# Testes com coverage
make test-coverage

# Benchmarks
make benchmark
```

## 📦 Deploy

### Docker

```bash
# Build da imagem
make docker-build

# Deploy com Docker Compose
docker-compose up -d
```

### Produção

```bash
# Build otimizado
make build-prod

# A aplicação estará em bin/main
./bin/main
```

## 🤝 Contribuindo

1. Fork o projeto
2. Crie uma branch para sua feature (`git checkout -b feature/AmazingFeature`)
3. Commit suas mudanças (`git commit -m 'Add some AmazingFeature'`)
4. Push para a branch (`git push origin feature/AmazingFeature`)
5. Abra um Pull Request

### Padrões de Código

- Siga as convenções do Go (gofmt, golint)
- Escreva testes para novas funcionalidades
- Mantenha a cobertura de testes acima de 80%
- Use nomes descritivos para variáveis e funções
- Documente funções públicas

## 📝 Roadmap

### v1.0 - MVP (Atual)
- [x] Sistema de autenticação
- [x] CRUD de usuários
- [x] Posts e interações básicas
- [x] Roteiros de viagem
- [x] Sistema de avaliações

### v1.1 - Melhorias Sociais
- [ ] Sistema de comentários
- [ ] Chat/mensagens privadas
- [ ] Notificações push
- [ ] Upload de imagens/vídeos

### v1.2 - IA e Recomendações
- [ ] IA para sugestões de roteiros
- [ ] Recomendações personalizadas
- [ ] Análise de preferências

### v1.3 - Parcerias Empresariais
- [ ] Dashboard empresarial
- [ ] Roteiros patrocinados
- [ ] Analytics para empresas

## 📄 Licença

Este projeto está licenciado sob a Licença MIT - veja o arquivo [LICENSE](LICENSE) para detalhes.

## 👥 Equipe

- **Ulpio** - Desenvolvimento inicial - [@Ulpio](https://github.com/Ulpio)

## 📞 Suporte

- 📧 Email: [suporte@guia.app](mailto:suporte@guia.app)
- 💬 Issues: [GitHub Issues](https://github.com/Ulpio/guIA-backend/issues)

---

⭐ Se este projeto te ajudou, considere dar uma estrela!

## 🔗 Links Úteis

- [Documentação da API](docs/api.md) (em desenvolvimento)
- [Guia de Contribuição](CONTRIBUTING.md) (em desenvolvimento)
- [Changelog](CHANGELOG.md) (em desenvolvimento)