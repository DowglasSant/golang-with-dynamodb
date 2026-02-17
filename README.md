# CRUD Go + DynamoDB — Local e AWS (ECS Fargate)

Projeto educativo para aprender os conceitos fundamentais do **Amazon DynamoDB** e **deploy na AWS** usando **Go** com containers.

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

Para alternar entre local e AWS, use a variavel de ambiente `ENV`:
- `ENV=local` (padrao) — usa DynamoDB Local
- `ENV=aws` — usa DynamoDB real na AWS com credenciais do ambiente (IAM Role)

---

## Deploy na AWS — Guia Passo a Passo

### Arquitetura alvo

```
Internet
   │
   ▼
[ ALB - Application Load Balancer ]     porta 80
   │        (subnets publicas)
   ▼
[ ECS Fargate - Task/Container ]        porta 8080
   │        (subnets privadas)
   ▼
[ DynamoDB - Tabela Users ]             serverless
```

Todos os componentes ficam dentro de uma **VPC** (Virtual Private Cloud).

---

### Passo 1 — Entender a VPC e Networking

**O que e uma VPC?**

VPC (Virtual Private Cloud) e sua rede privada isolada dentro da AWS. Pense nela como o "predio" onde seus recursos ficam. Nada entra ou sai sem regras explicitas.

**Conceitos-chave:**

| Conceito | O que e | Analogia |
|----------|---------|----------|
| **VPC** | Rede privada isolada | O predio inteiro |
| **Subnet publica** | Sub-rede com acesso a internet | Andar com porta pra rua |
| **Subnet privada** | Sub-rede SEM acesso direto a internet | Andar interno, sem janela |
| **Internet Gateway (IGW)** | Porta de entrada da internet para a VPC | Portaria do predio |
| **NAT Gateway** | Permite que recursos privados acessem a internet (mas ninguem de fora acessa eles) | Porteiro que faz entregas para andares internos |
| **Security Group** | Firewall por recurso (quais portas aceitar) | Fechadura de cada sala |

**Por que subnets publicas E privadas?**

- O **ALB** fica em subnets **publicas** — ele precisa receber requests da internet.
- O **Fargate (container)** fica em subnets **privadas** — ninguem acessa ele diretamente, so via ALB. Mais seguro.
- O container precisa de um **NAT Gateway** para acessar a internet (baixar imagens, falar com DynamoDB).

**O que fazer no Console AWS:**

1. Va em **VPC** > **Your VPCs**
2. Voce pode usar a **VPC default** (ja vem com subnets publicas) ou criar uma nova
3. Se criar nova, use o wizard **"VPC and more"** que cria tudo automaticamente:
   - 1 VPC
   - 2 subnets publicas (em AZs diferentes — o ALB exige pelo menos 2)
   - 2 subnets privadas
   - 1 Internet Gateway
   - 1 NAT Gateway
4. Anote os IDs das subnets publicas e privadas — voce vai precisar

---

### Passo 2 — Criar a tabela no DynamoDB

Diferente do ambiente local, aqui a tabela pode ser criada manualmente pelo console (ou o proprio app cria na primeira execucao).

1. Va em **DynamoDB** > **Create table**
2. Configure:
   - Table name: `Users`
   - Partition key: `id` (String)
   - Table settings: **Default settings** (On-Demand)
3. Clique **Create table**

Pronto. Sem servidor, sem cluster, sem disco. O DynamoDB e **serverless** — a AWS cuida de tudo.

---

### Passo 3 — ECR (Elastic Container Registry)

**O que e o ECR?**

ECR e o "Docker Hub da AWS" — um repositorio privado para armazenar suas imagens Docker. O ECS puxa as imagens de la para rodar os containers.

**Criar o repositorio:**

1. Va em **ECR** > **Create repository**
2. Nome: `golang-dynamodb`
3. Mantenha as configs padrao e crie

**Build e push da imagem:**

```bash
# 1. Autenticar Docker no ECR (substitua ACCOUNT_ID e REGION)
aws ecr get-login-password --region us-east-1 | docker login --username AWS --password-stdin ACCOUNT_ID.dkr.ecr.us-east-1.amazonaws.com

# 2. Build da imagem
docker build -t golang-dynamodb .

# 3. Tag da imagem com a URL do ECR
docker tag golang-dynamodb:latest ACCOUNT_ID.dkr.ecr.us-east-1.amazonaws.com/golang-dynamodb:latest

# 4. Push para o ECR
docker push ACCOUNT_ID.dkr.ecr.us-east-1.amazonaws.com/golang-dynamodb:latest
```

Substitua `ACCOUNT_ID` pelo seu ID de conta AWS (12 digitos, encontra no canto superior direito do console).

---

### Passo 4 — IAM Roles (Permissoes)

**Por que IAM Roles?**

No ECS Fargate, o container nao tem credenciais AWS por padrao. Voce precisa criar **roles** (papeis) que dao permissoes especificas ao container. Isso e muito mais seguro que colocar access keys no codigo.

**Existem 2 roles diferentes:**

| Role | Para que serve | Quem usa |
|------|---------------|----------|
| **Task Execution Role** | Permissao para o ECS puxar a imagem do ECR e enviar logs pro CloudWatch | O **agente ECS** (infraestrutura) |
| **Task Role** | Permissao para o **seu codigo** acessar servicos AWS (DynamoDB, S3, etc) | O **seu container** (aplicacao) |

**Criar a Task Execution Role:**

1. Va em **IAM** > **Roles** > **Create role**
2. Trusted entity: **AWS service** > **Elastic Container Service** > **Elastic Container Service Task**
3. Adicione a policy: `AmazonECSTaskExecutionRolePolicy`
4. Nome: `ecsTaskExecutionRole`

**Criar a Task Role (acesso ao DynamoDB):**

1. Va em **IAM** > **Roles** > **Create role**
2. Trusted entity: **AWS service** > **Elastic Container Service** > **Elastic Container Service Task**
3. Adicione a policy: `AmazonDynamoDBFullAccess`
   - (Em producao, voce criaria uma policy restrita so para a tabela Users)
4. Nome: `ecsTaskRole-dynamodb`

---

### Passo 5 — ECS Cluster + Task Definition

**O que e o ECS?**

ECS (Elastic Container Service) e o orquestrador de containers da AWS. Ele decide **onde** e **como** rodar seus containers.

**O que e Fargate?**

Fargate e o modo **serverless** do ECS — voce nao gerencia servidores (EC2). So diz quanto de CPU e memoria precisa, e a AWS cuida do resto.

**Criar o Cluster:**

1. Va em **ECS** > **Clusters** > **Create cluster**
2. Nome: `golang-dynamodb-cluster`
3. Infrastructure: selecione apenas **AWS Fargate**
4. Crie o cluster

**Criar a Task Definition:**

1. Va em **ECS** > **Task definitions** > **Create new task definition**
2. Configure:
   - Task definition family: `golang-dynamodb-task`
   - Launch type: **AWS Fargate**
   - OS: **Linux/X86_64**
   - CPU: `0.25 vCPU` (256)
   - Memory: `0.5 GB` (512)
   - Task role: `ecsTaskRole-dynamodb`
   - Task execution role: `ecsTaskExecutionRole`
3. Na secao **Container**:
   - Name: `golang-dynamodb`
   - Image URI: `ACCOUNT_ID.dkr.ecr.us-east-1.amazonaws.com/golang-dynamodb:latest`
   - Container port: `8080`
   - Protocol: **TCP**
4. Na secao **Environment variables**:
   - Key: `ENV` → Value: `aws`
   - Key: `AWS_REGION` → Value: `us-east-1`
5. Crie a task definition

**O que significa CPU 0.25 vCPU e 0.5 GB?**

Fargate cobra por CPU e memoria alocados. Para uma API simples, os valores minimos sao mais que suficientes. Em producao voce ajustaria conforme a carga.

---

### Passo 6 — ALB (Application Load Balancer)

**O que e um Load Balancer?**

O ALB distribui requests entre seus containers. Mesmo com 1 so container, ele e util porque:
- Da um **DNS publico** estavel (voce nao acessa o container diretamente)
- Faz **health check** — se o container morrer, o ECS sobe outro
- Permite escalar para N containers sem mudar a URL

**Conceitos do ALB:**

```
ALB (porta 80)
 └── Listener (regra: porta 80 → encaminhar para Target Group)
      └── Target Group (grupo de containers na porta 8080)
           └── Target 1: container Fargate :8080
           └── Target 2: container Fargate :8080  (se escalar)
```

**Criar o Target Group:**

1. Va em **EC2** > **Target Groups** > **Create target group**
2. Target type: **IP addresses** (Fargate usa IPs, nao EC2 instances)
3. Nome: `golang-dynamodb-tg`
4. Protocol: **HTTP**, Port: **8080**
5. VPC: selecione a sua VPC
6. Health check path: `/users` (um GET que retorna 200)
7. Crie sem registrar targets (o ECS faz isso automaticamente)

**Criar o ALB:**

1. Va em **EC2** > **Load Balancers** > **Create Load Balancer** > **Application Load Balancer**
2. Nome: `golang-dynamodb-alb`
3. Scheme: **Internet-facing** (acessivel pela internet)
4. Subnets: selecione as **2 subnets publicas**
5. Security group: crie ou use um que permita **porta 80 de qualquer lugar** (0.0.0.0/0)
6. Listener: **HTTP : 80** → Forward to → `golang-dynamodb-tg`
7. Crie o ALB

---

### Passo 7 — ECS Service

**O que e um Service?**

O Service e o "gerente" que mantem seus containers rodando. Ele:
- Garante que o numero desejado de tasks esteja sempre rodando
- Reconecta ao ALB quando um container reinicia
- Faz rolling deploy quando voce atualiza a imagem

**Criar o Service:**

1. Va em **ECS** > **Clusters** > `golang-dynamodb-cluster` > **Create service**
2. Configure:
   - Launch type: **Fargate**
   - Task definition: `golang-dynamodb-task` (latest revision)
   - Service name: `golang-dynamodb-service`
   - Desired tasks: `1`
3. Networking:
   - VPC: selecione a sua VPC
   - Subnets: selecione as **subnets privadas**
   - Security group: crie um que permita **porta 8080 vindo do Security Group do ALB**
   - Auto-assign public IP: **DISABLED** (esta em subnet privada)
4. Load balancing:
   - Tipo: **Application Load Balancer**
   - Selecione o ALB: `golang-dynamodb-alb`
   - Container: `golang-dynamodb : 8080`
   - Target group: `golang-dynamodb-tg`
5. Crie o service

**Sobre o Security Group do Fargate:**

O container so deve aceitar trafego vindo do ALB. Entao a regra de entrada (inbound) deve ser:
- Porta: `8080`
- Source: **Security Group do ALB** (nao 0.0.0.0/0!)

Isso garante que ninguem acessa o container diretamente, so via Load Balancer.

---

### Passo 8 — Testar

1. Va em **EC2** > **Load Balancers** > `golang-dynamodb-alb`
2. Copie o **DNS name** (algo como `golang-dynamodb-alb-123456.us-east-1.elb.amazonaws.com`)
3. Teste:

```bash
# Criar usuario
curl -s -X POST http://SEU-ALB-DNS/users \
  -H "Content-Type: application/json" \
  -d '{"name":"Joao","email":"joao@email.com"}' | jq

# Listar todos
curl -s http://SEU-ALB-DNS/users | jq
```

Se o service nao subir, verifique:
- **ECS** > **Cluster** > **Service** > **Tasks** > clique na task > **Logs** (CloudWatch)
- O container consegue acessar o DynamoDB? (NAT Gateway configurado?)
- O Security Group permite trafego do ALB na porta 8080?
- A Task Role tem permissao de DynamoDB?

---

### Resumo do fluxo completo

```
1. Voce faz push da imagem Docker → ECR
2. ECS Fargate puxa a imagem do ECR e roda o container
3. O container roda com ENV=aws e usa a Task Role para acessar DynamoDB
4. O ALB recebe requests da internet na porta 80
5. O ALB encaminha para o container na porta 8080 (via Target Group)
6. O container processa e responde
```

### Custos estimados (para estudo)

| Servico | Custo aproximado |
|---------|-----------------|
| DynamoDB On-Demand | ~$0 (free tier: 25 WCU + 25 RCU) |
| Fargate (0.25 vCPU, 0.5 GB) | ~$0.01/hora (~$8/mes) |
| ALB | ~$0.02/hora (~$16/mes) |
| NAT Gateway | ~$0.045/hora (~$32/mes) |
| ECR | ~$0 (500MB free tier) |

**Dica:** destrua tudo depois de testar para nao gerar custos. O NAT Gateway e o mais caro.
