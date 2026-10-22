# Money Savior - Telegram Bot

<div align="center">

![Go](https://img.shields.io/badge/Go-00ADD8?style=for-the-badge&logo=go&logoColor=white)
![AWS](https://img.shields.io/badge/AWS-232F3E?style=for-the-badge&logo=amazon-aws&logoColor=FF9900)
![DynamoDB](https://img.shields.io/badge/DynamoDB-527FFF?style=for-the-badge&logo=amazon-dynamodb&logoColor=white)
![Telegram](https://img.shields.io/badge/Telegram-26A5E4?style=for-the-badge&logo=telegram&logoColor=white)
![Groq](https://img.shields.io/badge/Groq-F55036?style=for-the-badge)

Um **assistente pessoal de controle de gastos** no Telegram: registre gastos digitando ou falando e veja o resumo em um gráfico.

[Quick Start](#quick-start) • [Comandos](#comandos-disponíveis) • [Gastos por Áudio](#gastos-por-áudio) • [Arquitetura](#arquitetura)

</div>

---

## Features

[+] **Registro por texto** - `/gastei 45,50 supermercado débito`, com o valor guardado em centavos  
[+] **Registro por áudio** - mande um áudio ("comprei uma água por 3 reais e paguei no crédito") e o bot transcreve, interpreta e registra  
[+] **Consulta com gráfico** - resumo por categoria em imagem, com total, percentuais e o dia mais pesado do período  
[+] **Períodos flexíveis** - hoje, ontem, semana, mês, mês passado ou um intervalo de datas  
[+] **Clean Architecture** - handler, controller, usecase e repository separados, testáveis sem Telegram e sem AWS  
[+] **Logs estruturados** - uma linha por mensagem, sem dados pessoais e com segredos mascarados  
[+] **Banco de dados cloud** - DynamoDB da AWS, com consulta por período através de um índice secundário  

---

## Arquitetura

As dependências apontam sempre para dentro: `infra -> adapter -> core`. O `core` não conhece Telegram nem AWS.

```
Telegram
   |  (polling no modo local / webhook na Lambda)
   v
handler ........ recebe o update, traduz para uma mensagem neutra e roteia pelo comando
   |
   v
controller ..... lê os argumentos, valida o FORMATO e chama o usecase
   |
   v
usecase ........ orquestra as regras (o domain valida o CONTEÚDO)
   |
   v
port ........... contrato que o core espera do mundo externo
   |
   v
repository ..... grava e consulta no DynamoDB
   |
   v  (a resposta volta pelo controller)
presenter ...... monta o texto ou o gráfico PNG
   |
   v
sender ......... envia ao Telegram (texto ou foto com legenda)
```

Uma regra que organiza tudo: **formato é do controller, regra é do domain**. "Isso é um número?" é uma pergunta do controller; "o valor é maior que zero?" é do domain.

---

## Estrutura do Projeto

```
money-savior-telegram-bot/
├── app/
│   ├── local/
│   │   └── main.go                  # Entrypoint local (polling) e montagem das peças
│   └── src/
│       ├── adapter/
│       │   ├── handler/             # Roteamento e uma linha de log por mensagem
│       │   ├── controller/          # Leitura de argumentos e validação de formato
│       │   └── presenter/           # Textos, gráfico PNG e fontes embutidas (fonts/)
│       ├── core/
│       │   ├── domain/              # Expense, período, ID e exceções
│       │   │   └── exception/
│       │   ├── port/                # Contratos (repositório, leitor, gerador de ID, voz)
│       │   └── usecase/             # Registrar gasto e consultar gastos
│       │       └── input/           # Pedidos de cada caso de uso
│       └── infra/
│           ├── aws/                 # Client do DynamoDB
│           │   └── repository/      # Persistência e consulta por período
│           ├── config/              # Variáveis de ambiente
│           ├── http/                # Envio de mensagens ao Telegram
│           ├── id/                  # Gerador de ID dos gastos
│           └── logger/              # slog com mascaramento de segredos
├── go.mod
├── go.sum
└── README.md
```

---

## Quick Start

### Pré-requisitos

- **Go** >= 1.24
- **Conta AWS** com credenciais configuradas (`aws configure`)
- **Token do bot do Telegram**, criado no [@BotFather](https://t.me/botfather)
- **Chave da API da Groq** (para os áudios), gerada em [console.groq.com](https://console.groq.com)
- **Tabela no DynamoDB** (veja abaixo)

### Tabela do DynamoDB

| Item | Valor |
|------|-------|
| Partition key | `user_id` (Number) |
| Sort key | `expense_id` (String) |
| Capacidade | On-demand |

E um **índice secundário global**, necessário para consultar por período:

| Item | Valor |
|------|-------|
| Nome | `user_id-created_at-index` |
| Partition key | `user_id` (Number) |
| Sort key | `created_at` (String) |
| Projeção | All |

O nome do índice precisa ser exatamente esse. Enquanto ele estiver em `Creating` (preenchimento), as consultas por período falham.

### Permissões IAM

O bot só precisa gravar e consultar. Política mínima para o usuário (ou a role da Lambda):

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": ["dynamodb:PutItem", "dynamodb:Query"],
      "Resource": [
        "arn:aws:dynamodb:<REGIAO>:<ID_DA_CONTA>:table/<NOME_DA_TABELA>",
        "arn:aws:dynamodb:<REGIAO>:<ID_DA_CONTA>:table/<NOME_DA_TABELA>/index/*"
      ]
    }
  ]
}
```

O `/index/*` é obrigatório: consultar um índice conta como outro recurso.

### Instalação Local

```bash
# Clone o repositório
git clone https://github.com/AmrmDev/the-money-savior-telegram-bot.git
cd the-money-savior-telegram-bot

# Instale as dependências
go mod download

# Configure as variáveis de ambiente (uma por linha)
export TELEGRAM_BOT_TOKEN="seu_token_aqui"
export TABLE_NAME="nome_da_tabela"
export AWS_REGION="us-east-1"

# Chaves digitadas sem ficar no histórico do terminal
read -s -p "GROQ_API_KEY: " GROQ_API_KEY; echo
export GROQ_API_KEY

# Execute o bot
go run ./app/local
```

O modo local usa **polling**. Um bot que já tem webhook ativo (como o de produção na Lambda) recusa o polling, então use um **bot de desenvolvimento**, criado à parte no @BotFather, e uma tabela de desenvolvimento.

---

## Comandos Disponíveis

### Registrar Gasto
```
/gastei <valor> <categoria> [método]
```
**Exemplos:** `/gastei 45,50 supermercado débito` • `/gastei 21.90 uber`

- O valor aceita vírgula ou ponto, com até 2 casas decimais (`50`, `50,9`, `50.90`).
- A categoria e o método são normalizados (`mercado` vira `Mercado`, `pix` vira `Pix`).
- O método é opcional.

### Consultar Gastos
```
/consulta [período]
```

| Comando | Período |
|---------|---------|
| `/consulta` ou `/consulta mes` | Do dia 1 do mês atual até hoje |
| `/consulta hoje` | Hoje |
| `/consulta ontem` | Ontem |
| `/consulta semana` | Últimos 7 dias, contando hoje |
| `/consulta mes passado` | O mês anterior, inteiro |
| `/consulta 01/10 15/10` | Do dia 1 ao dia 15 (ano atual), os dois dias incluídos |

A resposta é uma **imagem** com o total, as barras por categoria (as 5 maiores, o resto agrupado em "Outros"), o percentual de cada uma e, quando o período tem mais de um dia, o **dia mais pesado**. Com uma categoria só, ou sem gastos, a resposta é em texto. As datas seguem o horário de Brasília.

### Ajuda
```
/help                      # Exibe os comandos
/start                     # Mensagem de boas-vindas
```

---

## Gastos por Áudio

Mande uma mensagem de voz dizendo o valor, o que foi e, se quiser, como pagou:

> "Comprei uma água por 3 reais e paguei no crédito"

O bot transcreve o áudio, interpreta a frase e registra o gasto, igual ao `/gastei`.

```
1. Usuário: envia o áudio
   |
2. Bot: baixa o áudio pelo Telegram
   |
3. Groq: transcreve o áudio em texto
   |
4. Groq: interpreta o texto e devolve DADOS estruturados (valor, categoria, método)
   |
5. Bot: valida com as mesmas regras do /gastei
   |
6. Bot: salva no DynamoDB e responde
```

Dois princípios guiam essa parte:

- **A IA não gera comandos em texto.** Ela devolve campos (o valor vem como texto, como `"3,00"`), que entram direto no mesmo caso de uso do `/gastei`.
- **A saída da IA é tratada como dado não confiável.** O valor é convertido para centavos pelo código, sem aritmética da IA, e passa pelas regras do domain como qualquer entrada digitada.

<!-- Se o gasto por áudio tiver confirmação por botão, ou uma resposta específica para quando o bot não entender o áudio, descrever aqui. -->

---

## Variáveis de Ambiente

| Variável | Descrição | Obrigatória |
|----------|-----------|-------------|
| `TELEGRAM_BOT_TOKEN` | Token do bot do Telegram | Sim |
| `TABLE_NAME` | Nome da tabela do DynamoDB | Sim |
| `AWS_REGION` | Região da AWS (ex.: `us-east-1`) | Sim |
| `GROQ_API_KEY` | Chave da API da Groq (transcrição e interpretação dos áudios) | Para áudio |
| `AWS_PROFILE` | Perfil das credenciais da AWS (só no modo local) | Não |
| `LOG_LEVEL` | `debug`, `info`, `warn` ou `error` (padrão: `info`) | Não |
| `LOG_FORMAT` | `text` ou `json` (padrão: `json` na Lambda, `text` no resto) | Não |

Nunca coloque tokens ou chaves em arquivos do repositório.

---

## Modelo de Dados

### Expense (domain)
```go
type Expense struct {
    ID          string    // Ex.: 123MSB456PA
    UserID      int64     // ID do usuário no Telegram
    AmountCents int64     // Valor em centavos (5090 = R$ 50,90)
    Category    string    // Normalizada (Mercado)
    Method      string    // Normalizado (Pix), opcional
    CreatedAt   time.Time // Instante do registro
}
```

A entidade não tem tags de banco e não conhece o Telegram. O modelo de persistência fica separado, em `infra/aws/repository`.

### Item no DynamoDB
```
{
  "user_id":      { "N": "123456789" },
  "expense_id":   { "S": "123MSB456PA" },
  "amount_cents": { "N": "5090" },
  "category":     { "S": "Mercado" },
  "method":       { "S": "Pix" },
  "created_at":   { "S": "2026-10-13T10:39:25.489Z" }
}
```

Decisões do modelo:

- **Dinheiro em centavos inteiros**, para somas exatas, sem erro de ponto flutuante.
- **ID no formato `123MSB456PA`** (3 dígitos, `MSB`, 3 dígitos, `PA`), gerado pelo sistema. Se um ID colidir para o mesmo usuário, o caso de uso sorteia outro e tenta de novo (até 3 vezes).
- **`created_at` em UTC, com tamanho fixo.** O índice ordena como texto, então `...25.000Z` e `...25.489Z` precisam ter o mesmo comprimento para a ordem alfabética coincidir com a cronológica.
- **O dia calendário é sempre o de Brasília** (UTC-3, fixo), tanto para agrupar quanto para os períodos.

---

## Fluxo de Operações

### Registrar Gasto
```
1. Usuário: /gastei 50,90 mercado pix
   |
2. Controller: confere os argumentos e converte "50,90" em 5090 centavos
   |
3. Usecase: gera o ID e cria o Expense (o domain valida valor e categoria)
   |
4. Repository: grava no DynamoDB (recusa se o ID já existir)
   |
5. Presenter: "Opa, gasto registrado!" com categoria e valor
```

### Consultar Gastos
```
1. Usuário: /consulta semana
   |
2. Controller: transforma "semana" em um período
   |
3. Domain: calcula o intervalo exato (meia-noite de Brasília, em UTC)
   |
4. Repository: consulta o índice user_id + created_at
   |
5. Usecase: soma o total, agrupa por categoria e acha o dia mais pesado
   |
6. Presenter: desenha o gráfico (ou monta o texto) e o Sender envia
```

---

## Logs

O bot escreve **uma linha por mensagem** com `log/slog`:

```
level=INFO  msg="update handled"  update_id=1001 user_id=123456789 command=gastei   duration_ms=420 outcome=ok
level=WARN  msg="update rejected" update_id=1002 user_id=123456789 command=gastei   duration_ms=3   outcome=invalid_input reason="invalid input: invalid amount format"
level=ERROR msg="update failed"   update_id=1003 user_id=123456789 command=consulta duration_ms=95  outcome=error error="..."
```

| Resultado | Nível | Significado |
|-----------|-------|-------------|
| `ok` | INFO | A mensagem foi atendida |
| `invalid_input` | WARN | Erro de quem digitou (formato, valor, período) |
| `error` | ERROR | Falha do bot (banco, Telegram, etc.) |

Regras de privacidade:

- **Nunca vão para o log:** texto da mensagem, valores, categorias, nomes de usuário, token, chaves de API e transcrições de áudio.
- O `command` registrado é o **nome da rota** (`start`, `gastei`, `consulta`, `help`), e não o que o usuário digitou.
- Os segredos informados ao logger são mascarados em qualquer lugar da linha, inclusive dentro de mensagens de erro e URLs.

Na Lambda, a saída em JSON vai para o CloudWatch e permite filtrar por campo (`command`, `outcome`, `duration_ms`).

---

## Testes

```bash
# Testes unitários (não usam Telegram, AWS nem rede)
go test ./app/src/...

# Compilação de tudo
go build ./...

# Gera as imagens do gráfico para conferir o visual
CHART_DIR=$PWD go test ./app/src/adapter/presenter -run Preview
```

Os casos de uso rodam com repositórios falsos, e o gráfico é validado decodificando o PNG gerado.

---

## Tecnologias

- **Linguagem:** Go 1.24
- **API Telegram:** [go-telegram-bot-api v5](https://github.com/go-telegram-bot-api/telegram-bot-api)
- **Banco de dados:** AWS DynamoDB ([aws-sdk-go-v2](https://github.com/aws/aws-sdk-go-v2))
- **Gráficos:** imagens PNG desenhadas em Go com `golang.org/x/image` (sem navegador e sem serviço externo)
- **Áudio:** [Groq](https://groq.com) (transcrição e interpretação)
- **Logs:** `log/slog`
- **Fontes embutidas:** Space Grotesk e JetBrains Mono, sob a [SIL Open Font License](https://openfontlicense.org) (os textos das licenças ficam em `app/src/adapter/presenter/fonts/`)

As versões das dependências estão no `go.mod`.

---

## Segurança

- [x] **Validação de entrada** em todos os comandos, em duas camadas (formato e regra de negócio)
- [x] **Proteção por ID de usuário:** cada usuário só enxerga os próprios gastos
- [x] **Segredos por variável de ambiente** (nunca no código) e **mascarados nos logs**
- [x] **Logs sem dados pessoais:** só identificadores e dados técnicos
- [x] **Permissões mínimas no IAM:** `PutItem` e `Query`, só na tabela e no índice
- [x] **Saída da IA tratada como dado não confiável** e validada pelo domain

---

## Roadmap Futuro

- [ ] Deploy em AWS Lambda (webhook via API Gateway) com o novo código
- [ ] Consulta por áudio ("quanto gastei ontem?")
- [ ] Filtro por categoria ("quanto gastei com uber no mês")
- [ ] Mais formatos de período (semana passada, intervalos que cruzam o ano)
- [ ] Registro de entradas (income) e um dashboard web
- [ ] Aviso no log quando o desenho do gráfico falhar
- [ ] Exportar dados (CSV/PDF)
- [ ] Limite de gasto diário e alertas

---

## Contribuições

Contribuições são bem-vindas! Por favor:

1. Fork o projeto
2. Crie uma branch (`git checkout -b feature/AmazingFeature`)
3. Commit suas mudanças (`git commit -m 'Add some AmazingFeature'`)
4. Push para a branch (`git push origin feature/AmazingFeature`)
5. Abra uma Pull Request

---

## Licença

Distribuído sob a licença MIT. Veja `LICENSE` para mais detalhes.

---

## Autores

**Thyago Toledo**
- Email: thyago10a2007@gmail.com
- GitHub: [@ThyagoToledo](https://github.com/ThyagoToledo)

**Amorim**
- GitHub: [@AmrmDev](https://github.com/AmrmDev)

---

<div align="center">

**[⬆ Voltar ao topo](#money-savior---telegram-bot)**

Made with effort using Go + AWS

</div>
