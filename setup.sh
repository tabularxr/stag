#!/bin/bash
set -e

echo "🚀 Setting up STAG development environment..."

# Check if Docker is running
if ! docker info > /dev/null 2>&1; then
    echo "❌ Docker is not running. Please start Docker Desktop first."
    echo "   - Launch Docker Desktop from Applications folder"
    echo "   - Wait for Docker icon to appear in menu bar"
    echo "   - Then run this script again"
    exit 1
fi

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo "❌ Go is not installed. Please install Go first:"
    echo "   macOS: brew install go"
    echo "   Other: https://golang.org/dl/"
    exit 1
fi

echo "✅ Docker is running"
echo "✅ Go is installed ($(go version))"

# Install dependencies
echo "📦 Installing Go dependencies..."
go mod download
go mod tidy

# Start ArangoDB
echo "🗄️  Starting ArangoDB..."
docker compose up -d arangodb

# Wait for ArangoDB to be ready
echo "⏳ Waiting for ArangoDB to be ready..."
sleep 10

# Check if ArangoDB is healthy
if docker compose ps arangodb | grep -q "healthy"; then
    echo "✅ ArangoDB is running and healthy"
else
    echo "⚠️  ArangoDB may still be starting up..."
fi

# Set environment variables
echo "🔧 Setting environment variables..."
export ARANGO_URL=http://localhost:8529
export ARANGO_DATABASE=stag
export ARANGO_USERNAME=root
export ARANGO_PASSWORD=stagpassword
export STAG_PORT=8080
export LOG_LEVEL=info

# Create .env file for convenience
cat > .env << EOF
ARANGO_URL=http://localhost:8529
ARANGO_DATABASE=stag
ARANGO_USERNAME=root
ARANGO_PASSWORD=stagpassword
STAG_PORT=8080
LOG_LEVEL=info
EOF

echo "✅ Created .env file with default configuration"

echo ""
echo "🎉 Setup complete! Next steps:"
echo "   1. Source the environment variables:"
echo "      source .env"
echo "   2. Run the application:"
echo "      make run"
echo "   3. Test the health endpoint:"
echo "      curl http://localhost:8080/api/v1/health"
echo ""
echo "📚 Full documentation: README.md"