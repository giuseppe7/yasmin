.PHONY: build

build:
	@echo
	@echo "⋮⋮ Building..."
	go build -ldflags "-X main.version=`cat build_number``date -u +.%Y%m%d%H%M%S`"

test: build
	@echo
	@echo "⋮⋮ Testing..."
	go test -count=1 ./... -coverprofile cover.out -v

review: test
	@echo
	@echo "⋮⋮ Reviewing tests..."
	go tool cover -html cover.out

container: test
	@echo
	@echo "⋮⋮ Creating container..."
	docker build -f ./build/package/Dockerfile -t yasmin .

local: container
	@echo
	@echo "⋮⋮ Creating local environment..."
	docker compose -f ./deployments/docker-compose.yaml --project-name yasmin up -d --force-recreate
	docker ps | grep yasmin

clean-local:
	docker compose -f ./deployments/docker-compose.yaml --project-name yasmin down

all: build test review
	@echo