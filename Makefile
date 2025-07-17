OUTPUT_DIR=bin
APP_NAME=app

OAS3_GENERATOR_DOCKER_IMAGE = openapitools/openapi-generator-cli

dev:
	gow run cmd/server/main.go

docs:
	rm -rf internal/docs/swagger.*
	swag init --dir . \
	  --generalInfo ./cmd/server/main.go \
	  --output ./internal/docs \
	  --outputTypes json,yaml \
	  --parseDependency false \
	  --parseInternal false \
	  --parseVendor false \
	  --exclude ./data,./bin,sync/atomic

docs-v3:
	# First generate v2 docs
	make docs
	# Then convert to v3 using openapi-generator in Docker
	docker run -v ${PWD}:/local $(OAS3_GENERATOR_DOCKER_IMAGE) \
		generate \
		-i /local/internal/docs/swagger.json \
		-g openapi \
		-o /local/internal/docs/v3 \
		--skip-validate-spec \
		--additional-properties=outputFileName=openapi.json

build:
# 	make docs
	go mod tidy
	#go build -o bin/app ./cmd/server
	go build -o $(OUTPUT_DIR)/$(APP_NAME) -ldflags "$(LDFLAGS)" ./cmd/server

tidy:
	go mod tidy

migrate-up:
	go run cmd/server/main.go migrate


