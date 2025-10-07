# 🛠️ bbtelgo – Telegram Bot Starter in Go

Questo progetto è una base solida per sviluppare **bot Telegram in Go** con integrazione a **MongoDB**.  
Offre un’architettura pulita, logging avanzato e repository già pronti per la gestione utenti e messaggi da usare come esempio.



## ✨ Funzionalità principali
- Gestione **polling** e **webhook** per ricevere gli update da Telegram.  
- Integrazione con **MongoDB** tramite driver ufficiale e repository già pronti.  
- Sistema di **logging personalizzato** con rotazione dei file e livelli (debug, info, warn, error).  
- Configurazione completa tramite file `.env`.  
- Struttura modulare facilmente estendibile.  

## 📂 Struttura del progetto
```
bbtelgo/
├── cmd/
│   └── bbtelgo/
│       └── main.go          # entrypoint dell'applicazione
├── internal/
│   ├── app/                 # inizializzazione bot, avvio, webhook/polling
│   │   └── app.go
│   ├── config/              # gestione configurazioni da env
│   │   └── config.go
│   ├── db/                  # connessione e setup MongoDB
│   │   └── mongo.go
│   ├── entities/            # modelli (User, Message, ecc.)
│   │   ├── user.go
│   │   └── message.go
│   ├── handlers/            # gestione update Telegram
│   │   └── handler.go
│   ├── logx/                # logger personalizzato
│   │   └── logx.go
│   ├── repo/                # repository pattern per Mongo
│   │   ├── user-repo.go
│   │   └── message-repo.go
│   └── utils/               # utility generiche (es. JSON pretty)
│       └── utils.go
├── logs/                    # directory di output dei log
│   └── bot.log
├── .env                     # variabili d'ambiente locali
├── .env.example             # esempio precompilato per altri dev
├── docker-compose.yml       # avvio rapido MongoDB
├── go.mod / go.sum          # moduli Go
├── makefile                 # comandi di build/run
└── README.md
```

## ⚙️ Requisiti

- [Go >= 1.21](https://go.dev/dl/)
- [Docker](https://docs.docker.com/get-docker/) (per MongoDB e avvio rapido)
- Token Bot Telegram da [@BotFather](https://t.me/BotFather)

## 🚀 Installazione

Clona il repository e scarica le dipendenze:

```bash
git clone https://github.com/tuo-username/bbtelgo.git
cd bbtelgo
make tidy
```


## 🐳 Avvio con Docker (MongoDB)

Il servizio MongoDB sarà disponibile su `mongodb://localhost:27017`

```bash
docker-compose up -d
```

## 🔧 Configurazione

Crea un file .env partendo da questo esempio:

```env
# Modalità di esecuzione (polling o webhook)
APP_MODE=polling

# Token Telegram
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

# Webhook (solo se APP_MODE=webhook)
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

## ▶️ Utilizzo

### Compilazione

```bash
make build
```
Il binario viene creato in ` .dist/bbtelgo`.

### Esecuzione

```bash
make run
```
Puoi passare parametri aggiuntivi al binario con la variabile `ARGS`.
### Sviluppo (hot reload con polling)
Assicurati di avere un certificato TLS valido:

```bash
make dev
```
Il processo ricompila e riavvia automaticamente quando modifichi file .go, go.mod o go.sum.

### Pulizia

```bash
make clean
```
Rimuove i binari compilati da `.dist/`

### Formattazione

```bash
make fmt
```


## 📚 Estensioni
- Aggiungi nuovi handler in internal/handlers/handler.go
- Usa i repository in internal/db/ per salvare o leggere dati da MongoDB.
- Modifica internal/entities/ per aggiungere nuove entità.