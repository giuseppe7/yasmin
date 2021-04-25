.PHONY: build

build:
	@echo
	@echo "⋮⋮ Building..."
	go build -ldflags "-X main.version=`cat build_number``date -u +.%Y%m%d%H%M%S`"

test: build
	@echo
	@echo "⋮⋮ Testing..."
	go test -count=1 ./... -coverprofile cover.out

review: test
	@echo
	@echo "⋮⋮ Reviewing tests..."
	go tool cover -html cover.out

all: build test review
	@echo