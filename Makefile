CMD := docker-compose
SRCS := main.go $(wildcard ./**/*.go)

all:
	@echo "make up   :: run docker containers"
	@echo "make run  :: run the application"
	@echo "make down :: stop docker containers"
	@echo "make fmt  :: format source files;"
	@echo ""
	@echo "managed source files;"
	@echo $(SRCS)

.PHONY: up down format clean
up:
	@$(CMD) up -d

run:
	@$(CMD) exec app go run main.go

down:
	@$(CMD) down

fmt:
	@$(CMD) exec app go fmt ./...

clean:
	@$(CMD) down --rmi all --volumes --remove-orphans
