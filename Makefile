# Установить переменные окружения
GOBIN=$(PWD)/bin
export GOBIN

# Путь к Go, если не установлен по умолчанию
GO ?= go

# Директория для генерации кода
GEN_DIR=$(PWD)/internal/pb

# Зависимости для установки
DEPENDENCIES = google.golang.org/protobuf/cmd/protoc-gen-go \
               google.golang.org/grpc/cmd/protoc-gen-go-grpc \
               github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2 \
               github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway \
               github.com/bufbuild/buf/cmd/buf

.PHONY: .install-dependencies
.install-dependencies:
	@mkdir -p $(GOBIN)
	@for dep in $(DEPENDENCIES) ; do \
		echo "Устанавливаю $$dep" ; \
		$(GO) install "$$dep@latest" ; \
	done


.PHONY: generate
generate: .install-dependencies
	@mkdir -p $(GEN_DIR)
	@$(GOBIN)/buf mod update
	@$(GOBIN)/buf build
	@$(GOBIN)/buf generate

.PHONY: run
run:
	go run ./cmd/bff/main.go --dotenv=true

.PHONY: build-image
build-image:
	docker build -t app -f build/Dockerfile .

run-in-container:
	 docker compose up

run-services:
	docker compose up centrifugo db

PG_PASSWORD=qwerty
.PHONY: migrate
migrate:
	goose -dir migrations postgres "user=postgres password=${PG_PASSWORD} dbname=postgres host=localhost port=5432 sslmode=disable" up

migrate-down:
	goose -dir migrations postgres "user=postgres password=${PG_PASSWORD} dbname=postgres host=localhost port=5432 sslmode=disable" down

.PHONY: templ
templ:
	templ generate ./internal/...