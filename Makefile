# Установить переменные окружения
GOBIN=$(PWD)/bin
export GOBIN

# Путь к Go, если не установлен по умолчанию
GO ?= go

# Директория для генерации кода
GEN_DIR=$(PWD)/internal/pb

# Директория с внешними контрактами
VENDOR_PROTO_FILES_DIR=$(PWD)/vendor.protogen

# Директория с внутренними контрактами
PROTO_FILES_DIR=$(PWD)/api

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
	@$(GOBIN)/buf generate