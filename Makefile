.PHONY: run run-server run-web build test clean install

run:
	@echo "Starting MES System..."
	@make -C server run &
	@sleep 2
	@cd web && pnpm dev

run-server:
	@make -C server run

run-web:
	@cd web && pnpm dev

build:
	@echo "Building server..."
	@make -C server build
	@echo "Building web..."
	@cd web && pnpm build

test:
	@echo "Testing server..."
	@make -C server test
	@echo "Testing web..."
	@cd web && pnpm typecheck

install:
	@echo "Installing server dependencies..."
	@cd server && go mod tidy
	@echo "Installing web dependencies..."
	@cd web && pnpm install

clean:
	@make -C server clean
	@cd web && rm -rf dist node_modules/.vite
