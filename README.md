# 💰 Money Savior Telegram Bot

Bot do Telegram escrito em **Go** para ajudar no controle de finanças pessoais, com persistência de dados no **Amazon DynamoDB**.

<!-- TODO: descreva em 1-2 frases o que o bot faz de fato (ex.: registrar gastos, definir metas, relatórios). -->

---

## ✨ Funcionalidades

<!-- TODO: ajuste esta lista de acordo com os comandos reais implementados em `app/`. -->

- Registro de receitas e despesas direto pelo chat do Telegram
- Armazenamento persistente no DynamoDB
- Consulta de saldo / resumo de gastos
- Interação por comandos (`/start`, `/help`, ...)

## 🛠️ Stack

| Camada | Tecnologia |
| --- | --- |
| Linguagem | Go 1.24 |
| Telegram | [go-telegram-bot-api v5](https://github.com/go-telegram-bot-api/telegram-bot-api) |
| Banco de dados | Amazon DynamoDB (AWS SDK for Go v2) |

## 📁 Estrutura do projeto

```
.
├── app/        # Código-fonte da aplicação
├── go.mod      # Dependências do módulo Go
└── go.sum
```

## ✅ Pré-requisitos

- [Go 1.24+](https://go.dev/dl/)
- Um bot criado via [@BotFather](https://t.me/BotFather) (para obter o token)
- Conta AWS com acesso ao DynamoDB e credenciais configuradas (`aws configure` ou variáveis de ambiente)

## 🚀 Como executar

1. **Clone o repositório**

   ```bash
   git clone https://github.com/AmrmDev/money-savior-telegram-bot.git
   cd money-savior-telegram-bot
   ```

2. **Instale as dependências**

   ```bash
   go mod download
   ```

3. **Configure as variáveis de ambiente**

   <!-- TODO: confirme os nomes reais das variáveis lidas no código. -->

   ```bash
   export TELEGRAM_BOT_TOKEN="seu-token-do-botfather"
   export AWS_REGION="us-east-1"
   export AWS_ACCESS_KEY_ID="..."
   export AWS_SECRET_ACCESS_KEY="..."
   ```

4. **Rode o bot**

   ```bash
   go run ./app
   ```

## 🗄️ Banco de dados

O bot utiliza o DynamoDB para guardar os dados dos usuários.

<!-- TODO: informe o nome da tabela, a chave de partição/ordenação e se a tabela é criada automaticamente. -->

## 💬 Comandos

<!-- TODO: preencha com os comandos reais. -->

| Comando | Descrição |
| --- | --- |
| `/start` | Inicia a conversa com o bot |
| `/help` | Lista os comandos disponíveis |

## 🤝 Contribuindo

Contribuições são bem-vindas!

1. Faça um fork do projeto
2. Crie uma branch: `git checkout -b feature/minha-feature`
3. Commit: `git commit -m "feat: minha feature"`
4. Push: `git push origin feature/minha-feature`
5. Abra um Pull Request

## 📄 Licença

<!-- TODO: adicione um arquivo LICENSE (ex.: MIT) e referencie aqui. -->

## 👤 Autor

Feito por [AmrmDev](https://github.com/AmrmDev).
