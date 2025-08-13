#!/bin/bash

# Weather Service Deployment Script
# This script demonstrates proper deployment practices for a production service

set -e  # Exit on any error

# Configuration
SERVICE_NAME="weather-service"
VERSION=${1:-latest}
DOCKER_IMAGE="${SERVICE_NAME}:${VERSION}"
CONTAINER_NAME="${SERVICE_NAME}-${VERSION}"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Logging function
log() {
    echo -e "${GREEN}[$(date +'%Y-%m-%d %H:%M:%S')] $1${NC}"
}

warn() {
    echo -e "${YELLOW}[$(date +'%Y-%m-%d %H:%M:%S')] WARNING: $1${NC}"
}

error() {
    echo -e "${RED}[$(date +'%Y-%m-%d %H:%M:%S')] ERROR: $1${NC}"
}

# Check prerequisites
check_prerequisites() {
    log "Checking prerequisites..."
    
    if ! command -v docker &> /dev/null; then
        error "Docker is not installed"
        exit 1
    fi
    
    if ! command -v curl &> /dev/null; then
        error "curl is not installed"
        exit 1
    fi
    
    log "Prerequisites check passed"
}

# Build the application
build() {
    log "Building application..."
    
    # Clean previous builds
    make clean
    
    # Build the application
    make build
    
    if [ $? -eq 0 ]; then
        log "Application built successfully"
    else
        error "Build failed"
        exit 1
    fi
}

# Build Docker image
build_docker() {
    log "Building Docker image: ${DOCKER_IMAGE}"
    
    docker build -t ${DOCKER_IMAGE} .
    
    if [ $? -eq 0 ]; then
        log "Docker image built successfully"
    else
        error "Docker build failed"
        exit 1
    fi
}

# Stop existing container
stop_existing() {
    log "Checking for existing containers..."
    
    if docker ps -q -f name=${CONTAINER_NAME} | grep -q .; then
        log "Stopping existing container: ${CONTAINER_NAME}"
        docker stop ${CONTAINER_NAME}
        docker rm ${CONTAINER_NAME}
        log "Existing container stopped and removed"
    else
        log "No existing container found"
    fi
}

# Deploy the service
deploy() {
    log "Deploying service..."
    
    # Run the container
    docker run -d \
        --name ${CONTAINER_NAME} \
        -p 8080:8080 \
        -e LOGGING_LEVEL=info \
        -e METRICS_ENABLED=true \
        --restart unless-stopped \
        ${DOCKER_IMAGE}
    
    if [ $? -eq 0 ]; then
        log "Service deployed successfully"
    else
        error "Deployment failed"
        exit 1
    fi
}

# Health check
health_check() {
    log "Performing health check..."
    
    local max_attempts=30
    local attempt=1
    
    while [ $attempt -le $max_attempts ]; do
        if curl -f -s http://localhost:8080/health > /dev/null; then
            log "Health check passed"
            return 0
        fi
        
        warn "Health check attempt ${attempt}/${max_attempts} failed, retrying in 2 seconds..."
        sleep 2
        attempt=$((attempt + 1))
    done
    
    error "Health check failed after ${max_attempts} attempts"
    return 1
}

# Test the service
test_service() {
    log "Testing the service..."
    
    # Test weather endpoint
    local test_lat=40.7128
    local test_lon=-74.0060
    
    log "Testing weather endpoint with coordinates: ${test_lat}, ${test_lon}"
    
    local response=$(curl -s "http://localhost:8080/weather?lat=${test_lat}&lon=${test_lon}")
    
    if echo "$response" | grep -q "temperature_type"; then
        log "Weather endpoint test passed"
        log "Response preview: $(echo "$response" | head -c 200)..."
    else
        error "Weather endpoint test failed"
        log "Response: $response"
        return 1
    fi
}

# Show service status
show_status() {
    log "Service status:"
    echo "=================="
    echo "Container: ${CONTAINER_NAME}"
    echo "Image: ${DOCKER_IMAGE}"
    echo "Port: 8080"
    echo "Health: http://localhost:8080/health"
    echo "Weather: http://localhost:8080/weather?lat=40.7128&lon=-74.0060"
    echo "Metrics: http://localhost:8080/metrics"
    echo ""
    
    docker ps --filter name=${CONTAINER_NAME} --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}"
}

# Main deployment flow
main() {
    log "Starting Weather Service deployment..."
    
    check_prerequisites
    build
    build_docker
    stop_existing
    deploy
    
    # Wait a moment for container to start
    sleep 5
    
    if health_check; then
        test_service
        show_status
        log "Deployment completed successfully! 🎉"
    else
        error "Deployment failed during health check"
        exit 1
    fi
}

# Handle script arguments
case "${1:-deploy}" in
    "deploy")
        main
        ;;
    "build")
        build
        ;;
    "docker")
        build_docker
        ;;
    "test")
        test_service
        ;;
    "status")
        show_status
        ;;
    "stop")
        stop_existing
        ;;
    "help"|"-h"|"--help")
        echo "Weather Service Deployment Script"
        echo "Usage: $0 [command]"
        echo ""
        echo "Commands:"
        echo "  deploy  - Full deployment (default)"
        echo "  build   - Build application only"
        echo "  docker  - Build Docker image only"
        echo "  test    - Test the service"
        echo "  status  - Show service status"
        echo "  stop    - Stop existing container"
        echo "  help    - Show this help message"
        ;;
    *)
        error "Unknown command: $1"
        echo "Use '$0 help' for usage information"
        exit 1
        ;;
esac
