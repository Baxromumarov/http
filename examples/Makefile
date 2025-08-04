.PHONY: help backend client test clean

# Default target
help:
	@echo "🚀 HTTP Library Examples"
	@echo ""
	@echo "Available commands:"
	@echo "  make backend    - Start the backend server"
	@echo "  make client     - Run the client demo"
	@echo "  make test       - Test the API with curl"
	@echo "  make clean      - Clean build artifacts"
	@echo "  make help       - Show this help message"

# Start the backend server
backend:
	@echo "🚀 Starting backend server..."
	@cd backend && go run main.go

# Run the client demo
client:
	@echo "📡 Running client demo..."
	@cd client && go run main.go

# Test the API with curl
test:
	@echo "🧪 Testing API endpoints..."
	@echo ""
	@echo "1. Health check:"
	@curl -s http://localhost:8080/api/health | jq .
	@echo ""
	@echo "2. Get all users:"
	@curl -s http://localhost:8080/api/users | jq .
	@echo ""
	@echo "3. Get user by ID:"
	@curl -s http://localhost:8080/api/users/1 | jq .
	@echo ""
	@echo "4. Create new user:"
	@curl -s -X POST http://localhost:8080/api/users \
		-H "Content-Type: application/json" \
		-d '{"name":"Test User","email":"test@example.com","age":25}' | jq .
	@echo ""
	@echo "5. Get server stats:"
	@curl -s http://localhost:8080/api/stats | jq .

# Clean build artifacts
clean:
	@echo "🧹 Cleaning build artifacts..."
	@rm -f backend/backend
	@rm -f client/client
	@echo "✅ Cleaned!"

# Build both backend and client
build:
	@echo "🔨 Building backend..."
	@cd backend && go build -o backend main.go
	@echo "🔨 Building client..."
	@cd client && go build -o client main.go
	@echo "✅ Built successfully!"

# Run backend in background and test
demo: build
	@echo "🎬 Starting demo..."
	@echo "Starting backend server in background..."
	@cd backend && ./backend &
	@sleep 2
	@echo "Running client demo..."
	@cd client && ./client
	@echo "Stopping backend server..."
	@pkill -f "./backend" || true
	@echo "✅ Demo completed!" 