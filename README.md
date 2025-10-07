# 🛠️ bbtelgo – Telegram Bot Starter in Go

This project is a solid foundation for building Telegram bots in Go with MongoDB integration.
It provides a clean architecture, advanced logging, and ready-to-use repositories for managing users and messages as examples.



## ✨ Key Features
- Support for both **polling** and **webhook** modes to receive updates from Telegram.
- **MongoDB** integration using the official driver and ready-to-use repositories.
- **Custom logging system** with file rotation and log levels (debug, info, warn, error).
- Full configuration via `.env` file.
- Modular structure, easy to extend.

## 📂 Project Structure
```
bbtelgo/
├── cmd/
│   └── bbtelgo/
│       └── main.go          # application entrypoint
├── internal/
│   ├── app/                 # bot initialization, startup, webhook/polling
│   │   └── app.go
│   ├── config/              # configuration from environment variables
│   │   └── config.go
│   ├── db/                  # MongoDB connection and setup
│   │   └── mongo.go
│   ├── entities/            # models (User, Message, etc.)
│   │   ├── user.go
│   │   └── message.go
│   ├── handlers/            # update handlers for Telegram
│   │   └── handler.go
│   ├── logx/                # custom logger
│   │   └── logx.go
│   ├── repo/                # repository pattern for Mongo
│   │   ├── user-repo.go
│   │   └── message-repo.go
│   └── utils/               # generic utilities (e.g., JSON pretty-print)
│       └── utils.go
├── logs/                    # log output directory
│   └── bot.log
├── .env                     # local environment variables
├── .env.example             # prefilled example for other developers
├── docker-compose.yml       # quick MongoDB startup
├── go.mod / go.sum          # Go modules
├── makefile                 # build/run commands
└── README.md
```

## ⚙️ Requirements

- [Go >= 1.21](https://go.dev/dl/)
- [Docker](https://docs.docker.com/get-docker/) (for MongoDB and quick startup)
- Token Bot Telegram from [@BotFather](https://t.me/BotFather)

## 🚀 Installation

Clone the repository and download dependencies:

```bash
git clone https://github.com/tuo-username/bbtelgo.git
cd bbtelgo
make tidy
```


## 🐳 Run with Docker (MongoDB)

MongoDB will be available at `mongodb://localhost:27017`

```bash
docker-compose up -d
```

## 🔧 Configuration

Create a `.env` file based on the example:

```env
# Execution mode (polling or webhook)
APP_MODE=polling

# Telegram token
APP_TELEGRAM_TOKEN=123456:ABC-DEF1234ghIkl-zyx57W2v1u123ew11

# Logging
APP_LOG_LEVEL=debug
APP_LOG_FILE=true
APP_LOG_FILE_PATH=logs/bot.log
APP_LOG_FILE_MAX_MB=10
APP_LOG_INCLUDE_SRC=true
APP_LOG_TIMEFORMAT=2006-01-02T15:04:05Z07:00

# HTTP client
APP_HTTPCLIENT_TIMEOUT=10
APP_HTTPCLIENT_TRANSPORT_MAXIDLECONNS=100
APP_HTTPCLIENT_TRANSPORT_IDLECONNTIMEOUT=90

# Webhook (only if APP_MODE=webhook)
APP_WEBHOOK_SECRET=supersecret
APP_WEBHOOK_PORT=8443
APP_WEBHOOK_PUBLIC_URL=https://tuo-dominio.com/bot
APP_WEBHOOK_TLS_CERT_FILE=cert.pem
APP_WEBHOOK_TLS_KEY_FILE=key.pem
APP_RESET_WEBHOOK=true

# MongoDB
MONGO_URI=mongodb://root:example@localhost:27017
MONGO_DB=telegrambot
MONGO_APPNAME=bbtelgo
MONGO_MIN_POOL_SIZE=1
MONGO_MAX_POOL_SIZE=20
MONGO_CONNECTION_TIMEOUT=10
MONGO_CMD_TIMEOUT=5
MONGO_MAX_CONNECTING_LIMIT=5

```

## ▶️ Usage

### Build

```bash
make build
```
The binary is created at ` .dist/bbtelgo`

### Run

```bash
make run
```
You can pass additional arguments to the binary with the `ARGS` variable.

### Development (hot reload with polling)
```bash
make dev
```
The process will automatically recompile and restart whenever you modify `.go`, `go.mod`, or `go.sum` files.
__(Make sure you have a valid TLS certificate if running in webhook mode).__

### Clean

```bash
make clean
```
Removes the binaries from `.dist/`

### Format

```bash
make fmt
```


## 📚 Extending
- Add new handlers in `internal/handlers/handler.go`
- Use repositories in `internal/db/` to persist or retrieve data from MongoDB.
- Modify `internal/entities/` to add new entities.