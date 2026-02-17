# CRUD Go + DynamoDB Local

Projeto educativo para aprender os conceitos fundamentais do **Amazon DynamoDB** usando **Go** com uma instancia local via Docker.

---

## O que e o DynamoDB?

O DynamoDB e um banco de dados NoSQL gerenciado pela AWS, projetado para alta performance em qualquer escala. Diferente de bancos relacionais (PostgreSQL, MySQL), ele **nao usa SQL, nao tem tabelas com schema fixo e nao faz JOINs**.

### Modelo de dados

```
Tabela
 └── Item (equivalente a uma "linha")
      └── Atributos (equivalente a "colunas", mas flexiveis por item)
```

- **Tabela**: agrupamento de itens (ex: Users, Products)
- **Item**: um registro individual (ex: um usuario)
- **Atributo**: um campo do item (ex: name, email). Cada item pode ter atributos diferentes — nao ha schema rigido

### Chaves

Toda tabela precisa de uma **chave primaria**, que pode ser:

| Tipo | Composicao | Quando usar |
|------|-----------|-------------|
| **Simple** | Partition Key (PK) | Quando a PK sozinha identifica o item unicamente |
| **Composite** | Partition Key + Sort Key (SK) | Quando precisa agrupar itens sob a mesma PK |

**Partition Key (PK):** determina em qual particao interna o item sera armazenado. O DynamoDB usa um hash da PK para distribuir os dados. Escolher uma boa PK e fundamental para distribuicao uniforme.

**Sort Key (SK):** permite ordenar e filtrar itens dentro da mesma particao. Exemplo: PK = `user_id`, SK = `order_date` para buscar pedidos de um usuario em ordem.

Neste projeto usamos uma **Simple Primary Key** com `id` (UUID) como Partition Key.

### Tipos de atributos

```
S  = String        "Joao"
N  = Number        42
B  = Binary        dados binarios
BOOL = Boolean     true/false
L  = List          ["a", "b", "c"]
M  = Map           {"endereco": {"rua": "..."}}
SS = String Set    {"azul", "verde"}
```

### Operacoes principais

| Operacao | O que faz | Custo |
|----------|-----------|-------|
| **PutItem** | Insere ou substitui um item inteiro | Baixo (acesso direto) |
| **GetItem** | Busca um item pela chave primaria | Muito baixo (O(1)) |
| **UpdateItem** | Modifica atributos especificos de um item | Baixo |
| **DeleteItem** | Remove um item pela chave primaria | Baixo |
| **Query** | Busca itens por PK (e opcionalmente filtra por SK) | Medio |
| **Scan** | Varre TODOS os itens da tabela | Alto (evitar em producao) |

### Query vs Scan

```
Query  → "Me de todos os pedidos do usuario X"       → eficiente, usa indice
Scan   → "Me de todos os itens da tabela inteira"     → caro, le tudo
```

**Regra de ouro:** sempre prefira **Query** sobre **Scan** em producao. Scan existe para casos simples ou migracoes.

### Indices secundarios (GSI e LSI)

Quando voce precisa buscar por um campo que nao e a chave primaria:

- **GSI (Global Secondary Index):** cria uma "visao" da tabela com outra PK/SK. Exemplo: buscar usuarios por email.
- **LSI (Local Secondary Index):** mesma PK da tabela, mas com outra SK. Deve ser criado junto com a tabela.

Neste projeto nao usamos indices secundarios (buscamos apenas por `id`), mas em um cenario real voce criaria um GSI para buscar por email, por exemplo.

### Capacidade e cobranca

| Modo | Como funciona |
|------|--------------|
| **On-Demand** (PAY_PER_REQUEST) | Paga por requisicao. Sem provisionar. Ideal para cargas imprevisiveis |
| **Provisioned** | Define RCU/WCU fixos. Mais barato para cargas previsiveis |

- **RCU** (Read Capacity Unit): 1 leitura consistente de ate 4KB/s
- **WCU** (Write Capacity Unit): 1 escrita de ate 1KB/s

Neste projeto usamos **On-Demand** por simplicidade.

### Expressions

O DynamoDB usa expressions (semelhante a prepared statements) para operacoes seguras:

```
UpdateExpression:          "SET #name = :name, #email = :email"
ExpressionAttributeNames:  {"#name": "name"}       ← aliases para palavras reservadas
ExpressionAttributeValues: {":name": "Joao"}       ← valores parametrizados
ConditionExpression:       "attribute_exists(id)"   ← so executa se a condicao for verdadeira
```

**Por que `#name`?** Porque `name` e uma palavra reservada do DynamoDB (existem mais de 500 palavras reservadas).

**Por que `:name`?** Para separar valores da logica, evitando injecao de expressoes.

---

## Arquitetura do projeto

```
cmd/api/main.go              → Bootstrap: cria client, tabela, inicia servidor
  │
  ├── handler/               → Recebe HTTP, valida input, retorna JSON
  │     └── user_handler.go
  │
  ├── service/               → Regras de negocio (gera UUID, valida existencia)
  │     └── user_service.go
  │
  ├── repository/            → Operacoes no DynamoDB (PutItem, GetItem, etc)
  │     └── user_repository.go   ← comentado com explicacoes detalhadas
  │
  ├── entity/                → Structs de dominio e DTOs
  │     └── user.go
  │
  └── pkg/dynamo/            → Client de conexao com DynamoDB Local
        └── client.go
```

O fluxo de uma requisicao:

```
HTTP Request → Handler → Service → Repository → DynamoDB
HTTP Response ← Handler ← Service ← Repository ←
```

---

## Como rodar

### Pre-requisitos
- Go 1.22+
- Docker

### 1. Subir DynamoDB Local
```bash
docker-compose up -d
```

### 2. Iniciar a API
```bash
go run cmd/api/main.go
```

O servidor sobe em `http://localhost:8080`.

### 3. Testar os endpoints

**Criar usuario:**
```bash
curl -s -X POST localhost:8080/users \
  -H "Content-Type: application/json" \
  -d '{"name":"Joao Silva","email":"joao@email.com"}' | jq
```

**Listar todos:**
```bash
curl -s localhost:8080/users | jq
```

**Buscar por ID:**
```bash
curl -s localhost:8080/users/{id} | jq
```

**Atualizar:**
```bash
curl -s -X PUT localhost:8080/users/{id} \
  -H "Content-Type: application/json" \
  -d '{"name":"Joao Atualizado","email":"joao.novo@email.com"}' | jq
```

**Deletar:**
```bash
curl -s -X DELETE localhost:8080/users/{id}
```

---

## Endpoints

| Metodo | Rota | Descricao |
|--------|------|-----------|
| POST | `/users` | Criar usuario |
| GET | `/users` | Listar todos |
| GET | `/users/{id}` | Buscar por ID |
| PUT | `/users/{id}` | Atualizar usuario |
| DELETE | `/users/{id}` | Deletar usuario |

---

## DynamoDB Local vs AWS

| Aspecto | Local (Docker) | AWS |
|---------|---------------|-----|
| Custo | Gratuito | Pago (On-Demand ou Provisioned) |
| Persistencia | Em memoria (perde ao reiniciar) | Duravel e replicado |
| Credenciais | Qualquer valor funciona | IAM real necessario |
| Endpoint | `http://localhost:8000` | `https://dynamodb.{region}.amazonaws.com` |
| Uso | Desenvolvimento e testes | Producao |

Para migrar para AWS, basta remover o `BaseEndpoint` do client e configurar credenciais reais via IAM.
