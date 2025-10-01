# ---- Config -----------------------------------------------
APP      := bbtelgo
PKG      := ./...
MAIN     := ./cmd/$(APP)/main.go
BIN_DIR  := .dist
BIN      := $(BIN_DIR)/$(APP)

# Force use vendor with: make build USE_VENDOR=1
USE_VENDOR ?= 0
ifeq ($(USE_VENDOR),1)
  GOFLAGS := -mod=vendor
endif

# list src Go (exclude vendor/)
SRC := $(shell find . -type f -name '*.go' -not -path './vendor/*')

# ---- Tool local (pin + install in ./.bin) ----------------------------
BIN_LOCAL := .bin
REFLEX    := $(BIN_LOCAL)/reflex
TOOLS     := github.com/cespare/reflex@v0.3.1

# Pattern file for watch in dev
DEV_PAT := (\.go$$|go\.mod|go\.sum)

# ---- Main Target ------------------------------------------------
.PHONY: help
help: ## Show list of target available
	@echo "Target available:"
	@grep -E '^[a-zA-Z0-9_-]+:.*?## ' $(lastword $(MAKEFILE_LIST)) | sed -E 's/:.*## /\t- /'

.PHONY: all
all: build ## Compile bin (default)

.PHONY: build
build: $(BIN) ## Compile bin in $(BIN_DIR)

# Rule of build (ricompile only if edit src)
$(BIN): $(SRC) go.mod | $(BIN_DIR)
	@echo ">> building $(APP)"
	go build $(GOFLAGS) -o $(BIN) $(MAIN)

$(BIN_DIR):
	@mkdir -p $(BIN_DIR)

.PHONY: run
run: build ## Compile and run bin (use ARGS="..." for arguments)
	@echo ">> running $(APP) $(ARGS)"
	./$(BIN) $(ARGS)

.PHONY: test
test: ## Run test (verbose)
	@echo ">> go test"
	go test -v $(PKG)

.PHONY: fmt
fmt: ## Format code
	@echo ">> go fmt"
	go fmt $(PKG)

.PHONY: tidy
tidy: ## Align modules (go.mod/go.sum)
	@echo ">> go mod tidy"
	go mod tidy

.PHONY: vet
vet: ## Static analysis (vet)
	@echo ">> go vet"
	go vet $(PKG)

.PHONY: clean
clean: ## Clean bin
	@echo ">> cleaning $(BIN_DIR)"
	rm -rf $(BIN_DIR)

# ---- Dev (watch & restart) -------------------------------------------
.PHONY: dev
dev: ## Watch & restart (polling)
	@bash -c 'set -euo pipefail; \
	  pid=""; last=""; \
	  watch() { \
	    find . -path "./vendor" -prune -o -type f \( -name "*.go" -o -name "go.mod" -o -name "go.sum" \) -print0 \
	      | xargs -0 stat -c "%n %Y" 2>/dev/null \
	      | md5sum | awk "{print \$$1}"; \
	  }; \
	  while true; do \
	    cur="$$(watch)"; \
	    if [ "$$cur" != "$$last" ]; then \
	      echo ">> change detected"; \
	      last="$$cur"; \
	      if [ -n "$$pid" ] && kill -0 "$$pid" 2>/dev/null; then \
	        kill "$$pid" || true; \
	        wait "$$pid" 2>/dev/null || true; \
	      fi; \
	      $(MAKE) -s build; \
	      ./$(BIN) $(ARGS) & pid="$$!"; \
	    fi; \
	    sleep 1; \
	  done'

# ---- Vendor ------------------------
.PHONY: vendor
vendor: ## Populate ./vendor with the module's dependencies
	@echo ">> go mod vendor"
	go mod vendor

