# 📘 Documentação Técnica Completa - ToDo Microservice

## 📋 Índice

1. [Visão Geral da Arquitetura](#1-visão-geral-da-arquitetura)
2. [Estrutura do Código Go](#2-estrutura-do-código-go)
3. [Camada de Domínio (Models)](#3-camada-de-domínio-models)
4. [Camada de Persistência (Repository Pattern)](#4-camada-de-persistência-repository-pattern)
5. [Camada de Aplicação (Handlers)](#5-camada-de-aplicação-handlers)
6. [Camada de Infraestrutura](#6-camada-de-infraestrutura)
7. [Containerização com Docker](#7-containerização-com-docker)
8. [Orquestração com Kubernetes](#8-orquestração-com-kubernetes)
9. [Fluxo de Dados Completo](#9-fluxo-de-dados-completo)
10. [Guia de Apresentação](#10-guia-de-apresentação)

---

## 1. Visão Geral da Arquitetura

### 1.1 Padrão Arquitetural

Este projeto implementa uma **Arquitetura Limpa (Clean Architecture)** combinada com o padrão **Hexagonal (Ports & Adapters)**:

```
┌─────────────────────────────────────────────────────────────┐
│                    Camada Externa                           │
│  (Delivery Mechanisms: HTTP, CLI, gRPC - futuro)            │
│                                                             │
│  ┌────────────────────────────────────────────────────┐     │
│  │         Camada de Aplicação                        │     │
│  │  (Handlers HTTP - chi router)                      │     │
│  │                                                    │     │
│  │  ┌──────────────────────────────────────────────┐  │     │
│  │  │      Camada de Domínio (Core)                │  │     │
│  │  │  - Entities (ToDo)                           │  │     │
│  │  │  - Interfaces (Repository)                   │  │     │
│  │  │  - Business Rules                            │  │     │
│  │  └──────────────────────────────────────────────┘  │     │
│  │                                                    │     │
│  │  ┌──────────────────────────────────────────────┐  │     │
│  │  │      Camada de Infraestrutura                │  │     │
│  │  │  - MemoryRepository                          │  │     │
│  │  │  - BoltDBRepository                          │  │     │
│  │  │  - Config                                    │  │     │
│  │  │  - HTTP Server                               │  │     │
│  │  └──────────────────────────────────────────────┘  │     │
│  └────────────────────────────────────────────────────┘     │
└─────────────────────────────────────────────────────────────┘
```

### 1.2 Princípios Aplicados

#### **SOLID Principles**

1. **Single Responsibility (SRP)**
   - Cada pacote tem uma única responsabilidade
   - `internal/todo/handler.go` → apenas lida com HTTP requests
   - `internal/todo/repository.go` → apenas define contrato de persistência
   - `internal/config/config.go` → apenas carrega configuração

2. **Open/Closed Principle (OCP)**
   - Repository interface permite adicionar novos backends sem modificar código existente
   - Aberto para extensão (novos repositories), fechado para modificação

3. **Liskov Substitution (LSP)**
   - Qualquer implementação de `Repository` pode substituir outra
   - `MemoryRepository` e `BoltDBRepository` são intercambiáveis

4. **Interface Segregation (ISP)**
   - Interfaces pequenas e específicas
   - Repository não força métodos desnecessários

5. **Dependency Inversion (DIP)**
   - Handlers dependem de interfaces, não de implementações concretas
   - Injeção de dependência via construtores

#### **Domain-Driven Design (DDD)**

- **Entity**: `ToDo` é uma entidade com identidade única (`ID`)
- **Repository**: Abstrai persistência do domínio
- **Service**: Handlers atuam como application services
- **Bounded Context**: Todo domínio isolado no pacote `internal/todo`

---

## 2. Estrutura do Código Go

### 2.1 Layout do Projeto (Padrão golang-standards)

```
arancia/
├── cmd/                          # Aplicações principais
│   └── server/
│       └── main.go              # Entry point da aplicação
│
├── internal/                     # Código privado da aplicação
│   ├── config/                  # Gerenciamento de configuração
│   │   └── config.go
│   ├── health/                  # Health checks
│   │   └── handler.go
│   ├── httpserver/              # Setup do servidor HTTP
│   │   └── router.go
│   └── todo/                    # Domínio de Todo
│       ├── model.go             # Entidades e DTOs
│       ├── repository.go        # Interface do repositório
│       ├── memory_repo.go       # Implementação in-memory
│       ├── boltdb_repo.go       # Implementação BoltDB
│       └── handler.go           # HTTP handlers
│
├── k8s/                         # Manifestos Kubernetes
├── deployments/                 # Alternativa para manifestos
├── Dockerfile                   # Multi-stage build
├── Makefile                    # Automação de build
├── go.mod                      # Definição do módulo Go
└── go.sum                      # Checksums das dependências
```

### 2.2 Pacote `cmd/server/main.go`

**Responsabilidade**: Bootstrap da aplicação

```go
func main() {
    // 1. Carrega configuração do ambiente
    cfg, err := config.Load()
    
    // 2. Inicializa repositório baseado na configuração
    var repo todo.Repository
    switch cfg.StorageBackend {
    case "boltdb":
        repo, err = todo.NewBoltDBRepository(cfg.BoltDBPath)
    case "memory":
        repo = todo.NewMemoryRepository()
    }
    
    // 3. Cria servidor HTTP com dependências injetadas
    server := httpserver.NewServer(cfg.Port, repo)
    
    // 4. Inicia servidor em goroutine
    go func() {
        server.Start()
    }()
    
    // 5. Espera sinal de terminação (SIGINT/SIGTERM)
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit
    
    // 6. Graceful shutdown com timeout de 5 segundos
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    server.Shutdown(ctx)
}
```

**Pontos-chave:**
- ✅ **Fail-fast**: Se config falhar, aplicação não inicia
- ✅ **Injeção de dependência**: Repository injetado no servidor
- ✅ **Graceful shutdown**: Permite requests em andamento finalizarem
- ✅ **Context com timeout**: Previne espera indefinida

---

## 3. Camada de Domínio (Models)

### 3.1 `internal/todo/model.go`

#### **Entity: ToDo**

```go
type ToDo struct {
    ID        string `json:"id"`        // UUID único
    Title     string `json:"title"`     // Descrição da tarefa
    Completed bool   `json:"completed"` // Status de conclusão
}
```

**Design Decisions:**
- `ID` como string (UUID) → facilita distribuição, sem colisões
- `Completed` como bool → simplicidade, extensível para enum futuramente
- Tags JSON → serialização automática

#### **DTOs (Data Transfer Objects)**

**CreateToDoRequest**
```go
type CreateToDoRequest struct {
    Title string `json:"title"`
}
```
- Separado da entidade para não expor campos internos
- Cliente não define ID ou Completed na criação

**UpdateToDoRequest**
```go
type UpdateToDoRequest struct {
    Title     *string `json:"title,omitempty"`
    Completed *bool   `json:"completed,omitempty"`
}
```
- Ponteiros permitem **partial updates** (campos opcionais)
- `omitempty` → JSON não inclui campos nil
- Permite atualizar apenas um campo sem afetar o outro

**Exemplo de uso:**
```json
// Atualizar apenas título
{"title": "Novo título"}

// Atualizar apenas status
{"completed": true}

// Atualizar ambos
{"title": "Finalizado", "completed": true}
```

---

## 4. Camada de Persistência (Repository Pattern)

### 4.1 `internal/todo/repository.go` - Interface

```go
type Repository interface {
    GetAll() ([]*ToDo, error)
    GetByID(id string) (*ToDo, error)
    Create(todo *ToDo) error
    Update(id string, todo *ToDo) error
    Delete(id string) error
    Close() error
}
```

**Por que usar interface?**

1. **Testabilidade**: Podemos criar mocks facilmente
2. **Flexibilidade**: Trocar backend sem modificar handlers
3. **Inversão de Dependência**: Handlers dependem de abstração, não de implementação

**Sentinel Errors:**
```go
var (
    ErrNotFound = errors.New("todo not found")
    ErrInvalid  = errors.New("invalid todo data")
)
```

### 4.2 `internal/todo/memory_repo.go` - Implementação In-Memory

```go
type MemoryRepository struct {
    mu    sync.RWMutex        // Mutex para thread-safety
    todos map[string]*ToDo    // Map para armazenamento
}
```

#### **Thread-Safety com RWMutex**

**Conceito:**
- `RWMutex` permite múltiplos leitores OU um escritor
- Mais eficiente que `Mutex` quando há mais leituras que escritas

**Exemplo de uso:**

```go
// Leitura (múltiplos goroutines podem ler simultaneamente)
func (r *MemoryRepository) GetAll() ([]*ToDo, error) {
    r.mu.RLock()              // Lock de leitura
    defer r.mu.RUnlock()      // Unlock automático ao sair
    
    // Operação de leitura
    result := make([]*ToDo, 0, len(r.todos))
    for _, todo := range r.todos {
        result = append(result, todo)
    }
    return result, nil
}

// Escrita (apenas um goroutine pode escrever por vez)
func (r *MemoryRepository) Create(todo *ToDo) error {
    r.mu.Lock()               // Lock de escrita
    defer r.mu.Unlock()
    
    r.todos[todo.ID] = todo   // Operação de escrita
    return nil
}
```

**Por que isso é importante?**
- ✅ Previne **race conditions** (condições de corrida)
- ✅ Permite múltiplas requisições GET simultâneas
- ✅ Garante consistência dos dados

### 4.3 `internal/todo/boltdb_repo.go` - Implementação BoltDB

#### **O que é BoltDB?**

BoltDB (agora bbolt) é um **banco de dados embarcado** (embedded database):
- ✅ **Sem servidor**: roda no mesmo processo da aplicação
- ✅ **Single file**: todos os dados em um arquivo
- ✅ **ACID transactions**: garantias de consistência
- ✅ **Key-Value store**: armazenamento simples e rápido
- ✅ **File locking**: apenas um processo pode abrir o arquivo

#### **Estrutura de Dados**

```
todos.db (BoltDB File)
└── Bucket: "todos"
    ├── Key: "uuid-1" → Value: JSON(ToDo)
    ├── Key: "uuid-2" → Value: JSON(ToDo)
    └── Key: "uuid-3" → Value: JSON(ToDo)
```

#### **Código Explicado**

```go
type BoltDBRepository struct {
    db *bbolt.DB
}

func NewBoltDBRepository(path string) (*BoltDBRepository, error) {
    // 1. Abre (ou cria) o arquivo de banco de dados
    db, err := bbolt.Open(path, 0600, nil)
    if err != nil {
        return nil, fmt.Errorf("failed to open bolt db: %w", err)
    }

    // 2. Cria bucket "todos" se não existir
    err = db.Update(func(tx *bbolt.Tx) error {
        _, err := tx.CreateBucketIfNotExists([]byte("todos"))
        return err
    })
    if err != nil {
        db.Close()
        return nil, fmt.Errorf("failed to create bucket: %w", err)
    }

    return &BoltDBRepository{db: db}, nil
}
```

**Operação de Leitura (View):**
```go
func (r *BoltDBRepository) GetByID(id string) (*ToDo, error) {
    var todo *ToDo

    // View = transação de leitura (read-only)
    err := r.db.View(func(tx *bbolt.Tx) error {
        bucket := tx.Bucket([]byte("todos"))
        if bucket == nil {
            return ErrNotFound
        }

        // Busca pelo ID (key)
        data := bucket.Get([]byte(id))
        if data == nil {
            return ErrNotFound
        }

        // Deserializa JSON
        todo = &ToDo{}
        return json.Unmarshal(data, todo)
    })

    return todo, err
}
```

**Operação de Escrita (Update):**
```go
func (r *BoltDBRepository) Create(todo *ToDo) error {
    // Update = transação de escrita
    return r.db.Update(func(tx *bbolt.Tx) error {
        bucket := tx.Bucket([]byte("todos"))
        if bucket == nil {
            return fmt.Errorf("bucket not found")
        }

        // Verifica se já existe
        if bucket.Get([]byte(todo.ID)) != nil {
            return ErrInvalid
        }

        // Serializa para JSON
        data, err := json.Marshal(todo)
        if err != nil {
            return err
        }

        // Salva no banco
        return bucket.Put([]byte(todo.ID), data)
    })
}
```

**ACID Transactions:**
- **Atomic**: Toda transação é completa ou nenhuma mudança é feita
- **Consistent**: Dados sempre em estado válido
- **Isolated**: Transações não interferem entre si
- **Durable**: Dados salvos sobrevivem a crashes

---

## 5. Camada de Aplicação (Handlers)

### 5.1 `internal/todo/handler.go`

#### **Estrutura do Handler**

```go
type Handler struct {
    repo Repository  // Dependência injetada via interface
}

func NewHandler(repo Repository) *Handler {
    return &Handler{repo: repo}
}
```

**Injeção de Dependência:**
- Handler recebe Repository no construtor
- Não cria suas próprias dependências (inversão de controle)
- Facilita testes (pode injetar mock)

#### **GET /todos - Listar Todos**

```go
func (h *Handler) GetAll(w http.ResponseWriter, r *http.Request) {
    // 1. Busca todos os itens do repositório
    todos, err := h.repo.GetAll()
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    // 2. Serializa para JSON e retorna
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(todos)
}
```

**Fluxo:**
1. Handler chama `repo.GetAll()`
2. Repository retorna slice de ToDos
3. JSON encoder converte para JSON
4. Response com Content-Type: application/json

#### **POST /todos - Criar Todo**

```go
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
    // 1. Parse do body JSON
    var req CreateToDoRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "invalid request body", http.StatusBadRequest)
        return
    }

    // 2. Validação
    if req.Title == "" {
        http.Error(w, "title is required", http.StatusBadRequest)
        return
    }

    // 3. Cria entidade com UUID único
    todo := &ToDo{
        ID:        uuid.New().String(),
        Title:     req.Title,
        Completed: false,
    }

    // 4. Persiste no repositório
    if err := h.repo.Create(todo); err != nil {
        if err == ErrInvalid {
            http.Error(w, err.Error(), http.StatusBadRequest)
            return
        }
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    // 5. Retorna 201 Created com o objeto criado
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(todo)
}
```

**Decisões de Design:**
- ✅ **UUID gerado no servidor**: Cliente não define ID
- ✅ **Completed sempre false**: Lógica de negócio no servidor
- ✅ **201 Created**: Status code correto para criação
- ✅ **Retorna objeto criado**: Cliente recebe ID gerado

#### **PUT /todos/{id} - Atualizar Todo**

```go
func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
    // 1. Extrai ID da URL
    id := chi.URLParam(r, "id")
    if id == "" {
        http.Error(w, "id is required", http.StatusBadRequest)
        return
    }

    // 2. Busca todo existente
    existing, err := h.repo.GetByID(id)
    if err != nil {
        if err == ErrNotFound {
            http.Error(w, err.Error(), http.StatusNotFound)
            return
        }
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    // 3. Parse do request
    var req UpdateToDoRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "invalid request body", http.StatusBadRequest)
        return
    }

    // 4. Aplica partial updates (apenas campos fornecidos)
    if req.Title != nil {
        existing.Title = *req.Title
    }
    if req.Completed != nil {
        existing.Completed = *req.Completed
    }

    // 5. Persiste mudanças
    if err := h.repo.Update(id, existing); err != nil {
        if err == ErrNotFound {
            http.Error(w, err.Error(), http.StatusNotFound)
            return
        }
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    // 6. Retorna objeto atualizado
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(existing)
}
```

**Partial Updates:**
- Ponteiros (`*string`, `*bool`) distinguem "não fornecido" de "valor zero"
- `nil` = campo não fornecido → mantém valor existente
- Valor presente → atualiza campo

**Exemplo:**
```json
// Request: {"completed": true}
// Resultado: Apenas 'completed' muda, 'title' permanece igual
```

#### **DELETE /todos/{id} - Deletar Todo**

```go
func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
    // 1. Extrai ID
    id := chi.URLParam(r, "id")
    if id == "" {
        http.Error(w, "id is required", http.StatusBadRequest)
        return
    }

    // 2. Deleta do repositório
    if err := h.repo.Delete(id); err != nil {
        if err == ErrNotFound {
            http.Error(w, err.Error(), http.StatusNotFound)
            return
        }
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    // 3. Retorna 204 No Content (sucesso sem body)
    w.WriteHeader(http.StatusNoContent)
}
```

**204 No Content:**
- Indica sucesso mas sem corpo na resposta
- Economiza largura de banda
- Padrão REST para DELETE

---

## 6. Camada de Infraestrutura

### 6.1 `internal/config/config.go` - Configuração

```go
type Config struct {
    Port           int
    StorageBackend string
    BoltDBPath     string
}

func Load() (*Config, error) {
    cfg := &Config{
        Port:           8080,              // Default port
        StorageBackend: "memory",          // Default storage
        BoltDBPath:     "/data/todos.db",  // Default path
    }

    // Sobrescreve com variáveis de ambiente
    if portStr := os.Getenv("PORT"); portStr != "" {
        port, err := strconv.Atoi(portStr)
        if err != nil {
            return nil, fmt.Errorf("invalid PORT: %w", err)
        }
        cfg.Port = port
    }

    if backend := os.Getenv("STORAGE_BACKEND"); backend != "" {
        if backend != "memory" && backend != "boltdb" {
            return nil, fmt.Errorf("invalid STORAGE_BACKEND: must be 'memory' or 'boltdb'")
        }
        cfg.StorageBackend = backend
    }

    if path := os.Getenv("BOLTDB_PATH"); path != "" {
        cfg.BoltDBPath = path
    }

    return cfg, nil
}
```

**12-Factor App:**
- Configuração via variáveis de ambiente
- Defaults sensatos para desenvolvimento
- Validação no carregamento
- Fail-fast se inválido

### 6.2 `internal/httpserver/router.go` - HTTP Server

```go
func NewServer(port int, todoRepo todo.Repository) *Server {
    router := chi.NewRouter()

    // Middleware stack
    router.Use(middleware.RequestID)     // Adiciona X-Request-ID único
    router.Use(middleware.RealIP)        // Extrai IP real do cliente
    router.Use(middleware.Logger)        // Loga todas as requests
    router.Use(middleware.Recoverer)     // Recupera de panics
    router.Use(middleware.Timeout(60 * time.Second)) // Timeout global

    // Registra rotas
    healthHandler := health.NewHandler()
    healthHandler.RegisterRoutes(router)

    todoHandler := todo.NewHandler(todoRepo)
    todoHandler.RegisterRoutes(router)

    // Cria HTTP server
    httpServer := &http.Server{
        Addr:         fmt.Sprintf(":%d", port),
        Handler:      router,
        ReadTimeout:  15 * time.Second,
        WriteTimeout: 15 * time.Second,
        IdleTimeout:  60 * time.Second,
    }

    return &Server{httpServer: httpServer}
}
```

**Middleware Chain:**
```
Request → RequestID → RealIP → Logger → Recoverer → Timeout → Handler
```

**Timeouts:**
- `ReadTimeout` (15s): Tempo máximo para ler request completa
- `WriteTimeout` (15s): Tempo máximo para escrever response
- `IdleTimeout` (60s): Tempo que conexão keep-alive fica aberta
- `Timeout` middleware (60s): Timeout da request completa

### 6.3 `internal/health/handler.go` - Health Checks

```go
type Handler struct {
    ready bool  // Flag de readiness
}

// Liveness: processo está vivo?
func (h *Handler) Liveness(w http.ResponseWriter, r *http.Request) {
    w.WriteHeader(http.StatusOK)
    w.Write([]byte("OK"))
}

// Readiness: pode receber tráfego?
func (h *Handler) Readiness(w http.ResponseWriter, r *http.Request) {
    if !h.ready {
        w.WriteHeader(http.StatusServiceUnavailable)
        w.Write([]byte("Not Ready"))
        return
    }
    w.WriteHeader(http.StatusOK)
    w.Write([]byte("OK"))
}
```

**Diferença entre Liveness e Readiness:**

| Probe | Propósito | Ação do K8s se falhar |
|-------|-----------|----------------------|
| **Liveness** | Processo está vivo? | **Reinicia o pod** |
| **Readiness** | Pode receber requisições? | **Remove do load balancer** |

**Exemplo de uso:**
- **Liveness**: Detecta deadlocks, processos travados
- **Readiness**: Detecta quando app está inicializando, conectando ao DB, etc.

---

## 7. Containerização com Docker

### 7.1 Dockerfile - Multi-Stage Build

```dockerfile
# ========================================
# STAGE 1: BUILD
# ========================================
FROM golang:1.22-alpine AS builder

# Instala dependências de build
RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /app

# Copia módulos primeiro (melhor cache)
COPY go.mod go.sum ./
RUN go mod download && go mod verify

# Copia código fonte
COPY . .

# Build do binário
# CGO_ENABLED=0: binário estático (sem dependências C)
# -ldflags="-w -s": remove debug info (reduz tamanho ~30%)
# -a: força rebuild de todos os pacotes
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s" \
    -a -installsuffix cgo \
    -o todo-service \
    ./cmd/server/main.go

# ========================================
# STAGE 2: RUNTIME
# ========================================
FROM alpine:latest

# Instala certificados CA e timezone data
RUN apk --no-cache add ca-certificates tzdata && \
    # Cria usuário não-root
    addgroup -g 1000 appuser && \
    adduser -D -u 1000 -G appuser appuser && \
    # Cria diretório para BoltDB
    mkdir -p /data && \
    chown -R appuser:appuser /data

WORKDIR /home/appuser

# Copia binário do stage de build
COPY --from=builder /app/todo-service .

# Muda ownership
RUN chown appuser:appuser todo-service

# Muda para usuário não-root (SEGURANÇA)
USER appuser

# Expõe porta
EXPOSE 8080

# Health check embutido
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health/live || exit 1

# Comando de execução
CMD ["./todo-service"]
```

#### **Por que Multi-Stage Build?**

**Sem Multi-Stage:**
```
Imagem final: golang:1.22-alpine (~800MB)
├── Compilador Go completo
├── Ferramentas de build
├── Headers C
├── Bibliotecas de desenvolvimento
└── Binário da aplicação (5MB)
```

**Com Multi-Stage:**
```
Imagem final: alpine:latest (~15MB)
├── Sistema operacional mínimo
├── Certificados CA
├── Timezone data
└── Binário da aplicação (5MB)
```

**Benefícios:**
- ✅ **94% menor** (~800MB → ~15MB)
- ✅ **Mais rápido** para fazer pull
- ✅ **Mais seguro** (menos superfície de ataque)
- ✅ **Mais barato** (menos armazenamento)

#### **Otimizações de Build**

**Layer Caching:**
```dockerfile
# Primeiro: dependências (muda raramente)
COPY go.mod go.sum ./
RUN go mod download

# Depois: código (muda frequentemente)
COPY . .
RUN go build ...
```

Se apenas código mudar, reutiliza cache das dependências.

**Binário Estático (CGO_ENABLED=0):**
- Sem dependências de bibliotecas C
- Pode rodar em qualquer imagem Linux (inclusive scratch)
- Facilita distribuição

**Strip Debug Info (-ldflags="-w -s"):**
- `-w`: Remove informações DWARF (debug)
- `-s`: Remove symbol table
- Resultado: binário ~30% menor

### 7.2 .dockerignore

```
# Git
.git/
.gitignore

# Documentação
*.md
docs/

# IDE
.vscode/
.idea/

# Build artifacts
bin/
*.exe
todo-service

# Testes
*_test.go
coverage.txt

# Temporários
*.tmp
*.log

# Database files
*.db
*.bolt
data/

# Kubernetes (não precisa no container)
deployments/
k8s/
*.yaml
*.yml

# Docker
Dockerfile*
.dockerignore
```

**Por que .dockerignore é importante?**
- ✅ **Build mais rápido**: Menos arquivos para copiar
- ✅ **Imagem menor**: Não inclui arquivos desnecessários
- ✅ **Mais seguro**: Não vaza .git ou arquivos sensíveis

---

## 8. Orquestração com Kubernetes

### 8.1 Conceitos Fundamentais

#### **Pod**
- Menor unidade deployável no Kubernetes
- Pode conter um ou mais containers
- Compartilha rede e volumes
- IP efêmero (muda a cada restart)

#### **Deployment**
- Gerencia ReplicaSets
- Define estado desejado (ex: 2 réplicas)
- Controla rollouts e rollbacks
- Self-healing (recria pods que falham)

#### **Service**
- IP estável para acessar pods
- Load balancer interno
- Service discovery (DNS)
- Tipos: ClusterIP, NodePort, LoadBalancer

#### **ConfigMap**
- Armazena configurações não-sensíveis
- Pode ser montado como volume ou env vars
- Desacoplamento de configuração e código

#### **PersistentVolumeClaim (PVC)**
- Requisição de armazenamento persistente
- Abstraí detalhes de implementação
- Dados sobrevivem a restarts de pods

### 8.2 Manifesto: Namespace

**Arquivo:** `k8s/namespace.yaml`

```yaml
apiVersion: v1
kind: Namespace
metadata:
  name: todo-app
  labels:
    app: todo-service
    environment: production
```

**Por que usar Namespace?**
- ✅ **Isolamento**: Separa recursos de diferentes apps/ambientes
- ✅ **Organização**: Facilita gerenciamento
- ✅ **RBAC**: Permite controle de acesso por namespace
- ✅ **Quotas**: Pode limitar recursos por namespace

### 8.3 Manifesto: ConfigMap

**Arquivo:** `k8s/configmap.yaml`

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: todo-config
  namespace: todo-app
  labels:
    app: todo-service
data:
  PORT: "8080"
  STORAGE_BACKEND: "memory"
  BOLTDB_PATH: "/data/todos.db"
```

**Como funciona:**
1. ConfigMap armazena pares chave-valor
2. Deployment referencia ConfigMap
3. Valores injetados como variáveis de ambiente nos pods

**Benefício:**
- Mudar ConfigMap não requer rebuild da imagem
- Mesmo container, configurações diferentes por ambiente

### 8.4 Manifesto: Deployment

**Arquivo:** `k8s/deployment.yaml`

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: todo-service
  namespace: todo-app
  labels:
    app: todo-service
    version: v1
spec:
  replicas: 2  # Alta disponibilidade
  
  selector:
    matchLabels:
      app: todo-service
  
  strategy:
    type: RollingUpdate
    rollingUpdate:
      maxSurge: 1        # +1 pod extra durante update
      maxUnavailable: 0  # Zero downtime
  
  template:
    metadata:
      labels:
        app: todo-service
        version: v1
    spec:
      securityContext:
        runAsNonRoot: true  # Segurança: não roda como root
        runAsUser: 1000
        fsGroup: 1000
      
      containers:
      - name: todo-service
        image: todo-service:latest
        imagePullPolicy: IfNotPresent
        
        ports:
        - containerPort: 8080
          name: http
          protocol: TCP
        
        env:
        - name: PORT
          valueFrom:
            configMapKeyRef:
              name: todo-config
              key: PORT
        - name: STORAGE_BACKEND
          valueFrom:
            configMapKeyRef:
              name: todo-config
              key: STORAGE_BACKEND
        
        resources:
          requests:
            cpu: "100m"      # 0.1 CPU core garantido
            memory: "128Mi"  # 128 MiB RAM garantido
          limits:
            cpu: "500m"      # Máximo 0.5 CPU core
            memory: "256Mi"  # Máximo 256 MiB RAM
        
        livenessProbe:
          httpGet:
            path: /health/live
            port: 8080
          initialDelaySeconds: 10
          periodSeconds: 10
          timeoutSeconds: 5
          failureThreshold: 3
        
        readinessProbe:
          httpGet:
            path: /health/ready
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 5
          timeoutSeconds: 3
          failureThreshold: 3
        
        securityContext:
          allowPrivilegeEscalation: false
          readOnlyRootFilesystem: false
          capabilities:
            drop:
            - ALL
```

#### **Explicação Detalhada**

**Replicas:**
```yaml
replicas: 2
```
- 2 pods rodando simultaneamente
- Load balancer distribui tráfego entre eles
- Se um falhar, outro continua respondendo

**Rolling Update Strategy:**
```yaml
strategy:
  type: RollingUpdate
  rollingUpdate:
    maxSurge: 1        # Pode criar 1 pod extra (total 3)
    maxUnavailable: 0  # Sempre mantém pelo menos 2 rodando
```

**Fluxo de Update:**
```
Estado inicial: Pod1, Pod2 (versão v1)
↓
Cria Pod3 (versão v2)
↓
Espera Pod3 ficar Ready
↓
Deleta Pod1
↓
Cria Pod4 (versão v2)
↓
Espera Pod4 ficar Ready
↓
Deleta Pod2
↓
Estado final: Pod3, Pod4 (versão v2)
```

**Resource Requests e Limits:**

| Tipo | Propósito | Exemplo |
|------|-----------|---------|
| **requests** | Recursos **garantidos** ao pod | cpu: 100m, memory: 128Mi |
| **limits** | Máximo de recursos que pod pode usar | cpu: 500m, memory: 256Mi |

**O que acontece se exceder?**
- **CPU limit**: Throttling (pod fica lento)
- **Memory limit**: Pod é **killed** (OOMKilled)

**Probes Explicadas:**

**Liveness Probe:**
```yaml
livenessProbe:
  httpGet:
    path: /health/live
    port: 8080
  initialDelaySeconds: 10  # Espera 10s após pod iniciar
  periodSeconds: 10        # Checa a cada 10s
  timeoutSeconds: 5        # Timeout de 5s
  failureThreshold: 3      # Falha após 3 tentativas consecutivas
```

Se falhar 3 vezes seguidas → Kubernetes **reinicia o pod**

**Readiness Probe:**
```yaml
readinessProbe:
  httpGet:
    path: /health/ready
    port: 8080
  initialDelaySeconds: 5   # Começa a checar após 5s
  periodSeconds: 5         # Checa a cada 5s
  timeoutSeconds: 3        # Timeout de 3s
  failureThreshold: 3      # Not ready após 3 falhas
```

Se falhar → Pod **removido do Service** (não recebe tráfego)

**Diferença crucial:**
- **Liveness falha** → Pod reinicia (problema grave)
- **Readiness falha** → Pod fica ativo mas sem tráfego (problema temporário)

### 8.5 Manifesto: Service

**Arquivo:** `k8s/service.yaml`

```yaml
apiVersion: v1
kind: Service
metadata:
  name: todo-service
  namespace: todo-app
  labels:
    app: todo-service
spec:
  type: ClusterIP  # Acessível apenas dentro do cluster
  
  selector:
    app: todo-service  # Seleciona pods com este label
  
  ports:
  - name: http
    protocol: TCP
    port: 80           # Porta do Service
    targetPort: 8080   # Porta do container
```

#### **Como o Service funciona?**

```
Cliente → Service (todo-service:80)
          ↓
          Endpoints (lista de IPs dos pods)
          ↓
    ┌─────┴─────┐
    ↓           ↓
  Pod1:8080  Pod2:8080
```

**Service Discovery (DNS):**
- Dentro do cluster: `http://todo-service.todo-app.svc.cluster.local`
- Mesmo namespace: `http://todo-service`
- Kubernetes DNS resolve para IP do Service

**Load Balancing:**
- Service distribui tráfego entre pods healthy
- Algoritmo: Round-robin por padrão
- Pods não-ready são excluídos automaticamente

**Tipos de Service:**

| Tipo | Uso | Acesso |
|------|-----|--------|
| **ClusterIP** | Interno | Apenas dentro do cluster |
| **NodePort** | Desenvolvimento | NodeIP:NodePort |
| **LoadBalancer** | Produção (cloud) | IP público externo |
| **ExternalName** | Proxy externo | DNS externo |

### 8.6 Manifesto: PVC (Persistência)

**Arquivo:** `k8s/pvc.yaml`

```yaml
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: todo-boltdb-pvc
  namespace: todo-app
  labels:
    app: todo-service
spec:
  accessModes:
  - ReadWriteOnce  # Apenas um node pode montar (requerido para BoltDB)
  
  resources:
    requests:
      storage: 1Gi  # Requisita 1 GiB de armazenamento
  
  volumeMode: Filesystem
```

**Access Modes:**

| Modo | Descrição | Compatibilidade |
|------|-----------|-----------------|
| **ReadWriteOnce (RWO)** | Um node pode montar para ler/escrever | Maioria dos storages |
| **ReadOnlyMany (ROX)** | Múltiplos nodes podem montar para ler | NFS, alguns cloud storages |
| **ReadWriteMany (RWX)** | Múltiplos nodes podem ler/escrever | NFS, CephFS, GlusterFS |

**Por que RWO para BoltDB?**
- BoltDB usa file locking
- Apenas um processo pode abrir o arquivo
- RWO garante isso no Kubernetes

**Deployment com PVC:**

```yaml
# Em deployment-boltdb.yaml
spec:
  template:
    spec:
      containers:
      - name: todo-service
        volumeMounts:
        - name: data
          mountPath: /data  # Onde BoltDB salva
      
      volumes:
      - name: data
        persistentVolumeClaim:
          claimName: todo-boltdb-pvc  # Referencia o PVC
```

**Fluxo:**
```
PVC (todo-boltdb-pvc) requisita 1Gi
↓
Storage Class provisiona PersistentVolume
↓
PVC faz bind com PV
↓
Pod monta PV em /data
↓
BoltDB salva em /data/todos.db
↓
Dados persistem mesmo se pod for deletado
```

---

## 9. Fluxo de Dados Completo

### 9.1 Request Flow: POST /todos

```
1. Cliente envia HTTP POST
   ↓
2. Kubernetes Service (LoadBalancer)
   ↓
3. Service seleciona Pod (round-robin)
   ↓
4. Chi Router recebe request
   ↓
5. Middleware chain processa:
   - RequestID: Adiciona X-Request-ID
   - Logger: Loga request
   - Timeout: Inicia timer de 60s
   ↓
6. Handler.Create() chamado
   ↓
7. Valida request (title não vazio)
   ↓
8. Cria entidade ToDo com UUID
   ↓
9. Chama repo.Create(todo)
   ↓
10a. MemoryRepository: Salva em map com mutex
10b. BoltDBRepository: Salva em transação ACID
   ↓
11. Retorna 201 Created com JSON
   ↓
12. Response volta pelo mesmo caminho
   ↓
13. Cliente recebe response
```

### 9.2 Deployment Flow: kubectl apply

```
1. kubectl apply -f deployment.yaml
   ↓
2. API Server valida manifesto
   ↓
3. Salva no etcd (state store)
   ↓
4. Controller Manager detecta mudança
   ↓
5. ReplicaSet Controller cria pods
   ↓
6. Scheduler escolhe nodes para pods
   ↓
7. Kubelet no node:
   - Pull da imagem Docker
   - Cria container
   - Inicia processo
   ↓
8. Container inicia:
   - Config.Load() lê env vars
   - Inicializa Repository
   - HTTP server escuta na porta 8080
   ↓
9. Probes começam:
   - Liveness: GET /health/live
   - Readiness: GET /health/ready
   ↓
10. Se readiness OK:
    - Pod adicionado ao Service
    - Começa a receber tráfego
```

### 9.3 Failure Scenario: Pod Crash

```
1. Pod crashea (panic, OOMKill, etc)
   ↓
2. Liveness probe falha 3x
   ↓
3. Kubelet reporta ao API Server
   ↓
4. Deployment Controller detecta
   ↓
5. ReplicaSet cria novo pod
   ↓
6. Novo pod inicia
   ↓
7. Readiness probe passa
   ↓
8. Service adiciona novo pod
   ↓
9. Tráfego roteado para pods healthy
   ↓
10. Pod antigo é terminado (graceful shutdown)
```

**Com BoltDB:**
- ✅ Dados persistem (PVC mantém arquivo)
- ✅ Novo pod monta mesmo PVC
- ✅ Recupera dados automaticamente

**Com In-Memory:**
- ❌ Dados perdidos no crash
- ✅ Pod volta rápido (sem PVC)

### 9.4 Update Flow: Rolling Update

```
Estado: 2 pods v1 rodando

1. kubectl set image deployment/todo-service todo-service=v2
   ↓
2. Deployment Controller inicia rolling update
   ↓
3. Cria 1 pod v2 (maxSurge=1, total 3 pods)
   ↓
4. Pod v2 inicia, passa readiness
   ↓
5. Service adiciona pod v2 ao pool
   ↓
6. Deleta 1 pod v1 (maxUnavailable=0, mantém 2 rodando)
   ↓
7. Cria outro pod v2
   ↓
8. Pod v2 #2 inicia, passa readiness
   ↓
9. Deleta último pod v1
   ↓
Estado final: 2 pods v2 rodando

Durante TODO o processo:
- ✅ Sempre 2 pods respondendo requests
- ✅ Zero downtime
- ✅ Rollback automático se v2 falhar readiness
```

---

## 10. Guia de Apresentação

### 10.1 Roteiro para Vídeo (5-7 minutos)

#### **Minuto 1: Introdução (0:00-1:00)**

**Script:**
> "Olá! Hoje vou apresentar um microservice de ToDo completo, desenvolvido em Go, containerizado com Docker e orquestrado com Kubernetes. Este projeto demonstra práticas modernas de desenvolvimento backend e DevOps."

**Mostrar na tela:**
- Estrutura do projeto
- Tecnologias usadas (badges)

---

#### **Minuto 2: Arquitetura Limpa (1:00-2:00)**

**Script:**
> "A arquitetura segue Clean Architecture com Repository Pattern. Veja como está organizado:
> - Camada de Domínio: Entidades e interfaces
> - Camada de Aplicação: Handlers HTTP
> - Camada de Infraestrutura: Implementações concretas
> 
> O Repository Pattern permite trocar o backend de armazenamento sem modificar nenhum handler. Posso usar in-memory para desenvolvimento ou BoltDB para persistência."

**Mostrar:**
```bash
# Estrutura do código
tree -L 3 internal/

# Apontar para:
# - internal/todo/repository.go (interface)
# - internal/todo/memory_repo.go (impl 1)
# - internal/todo/boltdb_repo.go (impl 2)
```

---

#### **Minuto 3: Código Go (2:00-3:00)**

**Script:**
> "Vamos ver um exemplo de handler. O POST /todos:
> 1. Valida o request
> 2. Gera UUID único
> 3. Persiste via Repository interface
> 4. Retorna 201 Created
>
> Note a injeção de dependência - o handler recebe o Repository no construtor, não conhece qual implementação está usando."

**Mostrar:**
```go
// internal/todo/handler.go
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
    // Parse e validação
    var req CreateToDoRequest
    json.NewDecoder(r.Body).Decode(&req)
    
    // Cria entidade
    todo := &ToDo{
        ID: uuid.New().String(),
        Title: req.Title,
        Completed: false,
    }
    
    // Persiste (não sabe se é memory ou boltdb!)
    h.repo.Create(todo)
    
    // Retorna
    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(todo)
}
```

---

#### **Minuto 4: Docker (3:00-4:00)**

**Script:**
> "O Dockerfile usa multi-stage build para otimização:
> - Stage 1: Compila o binário em uma imagem Go completa
> - Stage 2: Copia apenas o binário para Alpine Linux
>
> Resultado: imagem final de apenas 15MB, contra 800MB se não usássemos multi-stage. O binário é estático, sem dependências, e roda como usuário não-root para segurança."

**Demonstrar:**
```bash
# Build
docker build -t todo-service:demo .

# Tamanho
docker images todo-service:demo

# Run
docker run -d -p 8080:8080 -e STORAGE_BACKEND=memory todo-service:demo

# Testar
curl http://localhost:8080/health/live
curl -X POST http://localhost:8080/todos -H "Content-Type: application/json" -d '{"title":"Demo"}'
curl http://localhost:8080/todos
```

---

#### **Minuto 5: Kubernetes (4:00-5:30)**

**Script:**
> "Agora vamos deployar no Kubernetes. Os manifestos incluem:
> - Namespace para isolamento
> - ConfigMap para configuração
> - Deployment com 2 réplicas para alta disponibilidade
> - Service para load balancing
> - Probes de liveness e readiness
>
> O Deployment usa Rolling Update para zero downtime. Veja: sempre mantém 2 pods rodando durante updates."

**Demonstrar:**
```bash
# Load image
minikube image load todo-service:demo

# Deploy
kubectl apply -f k8s/namespace.yaml
kubectl apply -f k8s/configmap.yaml
kubectl apply -f k8s/deployment.yaml
kubectl apply -f k8s/service.yaml

# Status
kubectl get pods -n todo-app

# Logs
kubectl logs -f -n todo-app -l app=todo-service

# Port forward
kubectl port-forward -n todo-app svc/todo-service 8080:80 &

# Testar
curl http://localhost:8080/todos
```

---

#### **Minuto 6: Demonstração de Resiliência (5:30-6:30)**

**Script:**
> "Vamos testar a resiliência. Com 2 réplicas rodando, vou deletar um pod e mostrar que o serviço continua respondendo sem interrupção. O Kubernetes automaticamente recria o pod deletado."

**Demonstrar:**
```bash
# Estado inicial
kubectl get pods -n todo-app
# Mostra 2 pods rodando

# Criar um todo
curl -X POST http://localhost:8080/todos -H "Content-Type: application/json" -d '{"title":"Test"}'

# Deletar um pod (simular crash)
kubectl delete pod -n todo-app -l app=todo-service --field-selector=status.phase=Running | head -1

# Serviço continua respondendo (outro pod ativo)
curl http://localhost:8080/todos

# Pod sendo recriado
kubectl get pods -n todo-app -w
```

---

#### **Minuto 7: Conclusão (6:30-7:00)**

**Script:**
> "Resumindo:
> - Arquitetura limpa e testável com Repository Pattern
> - Containerização eficiente com multi-stage build
> - Deployment cloud-native com Kubernetes
> - Alta disponibilidade e resiliência
> - Zero downtime deployments
>
> O código está disponível no GitHub com documentação completa, incluindo Makefile para automação. Obrigado!"

---

### 10.2 Comandos Úteis para Apresentação

#### **Setup Rápido**
```bash
# Terminal 1: Build e Deploy
make docker-build
make k8s-load-image
make k8s-deploy
make k8s-port-forward

# Terminal 2: Testes
make test-api
```

#### **Demonstrações**

**1. Partial Update:**
```bash
# Criar
curl -X POST http://localhost:8080/todos \
  -H "Content-Type: application/json" \
  -d '{"title":"Original"}' | jq '.'

# Copiar ID da response

# Atualizar apenas completed
curl -X PUT http://localhost:8080/todos/{ID} \
  -H "Content-Type: application/json" \
  -d '{"completed":true}' | jq '.'

# Title permanece "Original", completed agora é true
```

**2. Health Checks:**
```bash
# Liveness
curl -i http://localhost:8080/health/live

# Readiness
curl -i http://localhost:8080/health/ready
```

**3. Trocar Storage Backend:**
```bash
# Parar versão in-memory
kubectl delete -f k8s/deployment.yaml

# Aplicar versão BoltDB
kubectl apply -f k8s/pvc.yaml
kubectl apply -f k8s/configmap-boltdb.yaml
kubectl apply -f k8s/deployment-boltdb.yaml

# Verificar PVC bound
kubectl get pvc -n todo-app
```

---

### 10.3 Pontos-Chave para Mencionar

#### **Decisões Técnicas**
1. **Por que Go?**
   - Performance, concorrência nativa, binário único
   - Baixo uso de memória (importante em containers)
   - Excelente para microservices

2. **Por que Repository Pattern?**
   - Testabilidade (mocks)
   - Flexibilidade (trocar storage)
   - Manutenibilidade (separação de concerns)

3. **Por que Multi-Stage Build?**
   - Imagem 94% menor
   - Mais rápido para deploy
   - Mais seguro (menos superfície de ataque)

4. **Por que 2 Réplicas?**
   - Alta disponibilidade
   - Rolling updates sem downtime
   - Load balancing

#### **Possíveis Perguntas**

**P: "Por que não usar um banco de dados externo?"**
**R:** "Para este desafio, BoltDB demonstra persistência sem complexidade adicional. Em produção com múltiplas réplicas escrevendo, sim, usaríamos PostgreSQL ou similar. A arquitetura permite essa mudança facilmente - só criar PostgreSQLRepository implementando a mesma interface."

**P: "Como você testaria isso?"**
**R:** "Várias camadas: unit tests com mocks do repository, integration tests com testcontainers, e2e tests no Kubernetes com kind. O Repository Pattern facilita muito os testes."

**P: "E autenticação?"**
**R:** "Fora do escopo deste desafio, mas adicionaria middleware JWT. Chi facilita isso - basta adicionar na cadeia de middleware antes dos handlers."

**P: "Como você monitoraria em produção?"**
**R:** "Adicionaria:
- Prometheus para métricas (endpoint /metrics)
- Grafana para visualização
- ELK/Loki para logs centralizados
- Jaeger para distributed tracing
- Já temos health checks para Kubernetes"

---

### 10.4 Checklist Antes da Apresentação

- [ ] Código commitado e no GitHub
- [ ] README.md completo e formatado
- [ ] Docker image buildando sem erros
- [ ] Kubernetes local (minikube/kind) funcionando
- [ ] Testou todos os comandos
- [ ] Preparou terminais lado a lado
- [ ] Fonte grande o suficiente para vídeo
- [ ] Microfone testado
- [ ] Screen recording software pronto
- [ ] Roteiro impresso/visível

---

## 🎯 Conclusão

Este documento técnico cobriu:

1. ✅ **Arquitetura** - Clean Architecture e Hexagonal Pattern
2. ✅ **Código Go** - Cada pacote explicado linha por linha
3. ✅ **Padrões de Design** - Repository, SOLID, DDD
4. ✅ **Docker** - Multi-stage build e otimizações
5. ✅ **Kubernetes** - Todos os manifestos explicados
6. ✅ **Fluxos** - Como dados e deploys funcionam
7. ✅ **Apresentação** - Roteiro completo para vídeo

Use este documento como referência durante sua apresentação e estudo do projeto!

**Boa sorte! 🚀**
