# Deployment Strategy

**Version:** 1.0
**Date:** August 4, 2025
**Author:** Senior Staff Software Architect, Garnizeh
**Status:** In Progress

---

## 📋 Overview

This document outlines the comprehensive deployment strategy for the EngLog system, covering containerization, orchestration, cloud infrastructure, CI/CD pipelines, and environment management. The strategy emphasizes reliability, scalability, and maintainability across all deployment environments.

## 🏗️ Infrastructure Architecture

### Multi-Environment Setup

```mermaid
graph TD
    subgraph "Development Environment"
        A1[Developer Laptop]
        A2[Docker Compose]
        A3[Local Database]
        A4[Local Redis]
    end

    subgraph "Staging Environment"
        B1[Kubernetes Cluster]
        B2[Staging API Pods]
        B3[Staging Worker Pods]
        B4[Staging Database]
        B5[Staging Redis]
    end

    subgraph "Production Environment"
        C1[Kubernetes Cluster]
        C2[API Service Pods]
        C3[Worker Pool Pods]
        C4[Production Database]
        C5[Production Redis]
        C6[Load Balancer]
        C7[CDN]
    end

    A1 --> A2
    A2 --> A3
    A2 --> A4

    B1 --> B2
    B1 --> B3
    B2 --> B4
    B2 --> B5

    C1 --> C2
    C1 --> C3
    C2 --> C4
    C2 --> C5
    C6 --> C2
    C7 --> C6
```

### Cloud Provider Architecture (AWS)

```mermaid
graph TB
    subgraph "AWS Cloud Infrastructure"
        subgraph "Network Layer"
            VPC[VPC]
            IGW[Internet Gateway]
            NAT[NAT Gateway]
            ALB[Application Load Balancer]
        end

        subgraph "Compute Layer"
            EKS[EKS Cluster]
            EC2[EC2 Instances]
            FARGATE[Fargate Nodes]
        end

        subgraph "Storage Layer"
            RDS[RDS PostgreSQL]
            REDIS[ElastiCache Redis]
            S3[S3 Buckets]
            EBS[EBS Volumes]
        end

        subgraph "Security & Monitoring"
            IAM[IAM Roles]
            SECRETS[Secrets Manager]
            CW[CloudWatch]
            XRAY[X-Ray Tracing]
        end

        subgraph "CI/CD"
            CODECOMMIT[CodeCommit]
            CODEBUILD[CodeBuild]
            CODEPIPELINE[CodePipeline]
            ECR[ECR Registry]
        end
    end

    IGW --> ALB
    ALB --> EKS
    EKS --> EC2
    EKS --> FARGATE
    EC2 --> RDS
    EC2 --> REDIS
    EKS --> S3
    EKS --> EBS

    IAM --> EKS
    SECRETS --> EKS
    CW --> EKS
    XRAY --> EKS

    CODEPIPELINE --> CODEBUILD
    CODEBUILD --> ECR
    ECR --> EKS
```

## 🐳 Containerization

### Docker Configuration

```dockerfile
# Dockerfile for API Service
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Install dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./cmd/api

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata
WORKDIR /root/

# Copy the binary from builder stage
COPY --from=builder /app/main .

# Copy static files if any
COPY --from=builder /app/static ./static

# Create non-root user
RUN adduser -D -s /bin/sh appuser
USER appuser

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

EXPOSE 8080

CMD ["./main"]
```

```dockerfile
# Dockerfile for Worker Service
FROM golang:1.24-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git ca-certificates tzdata

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o worker ./cmd/worker

FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata
WORKDIR /root/

COPY --from=builder /app/worker .

RUN adduser -D -s /bin/sh appuser
USER appuser

# Health check for worker (checks if it can connect to queue)
HEALTHCHECK --interval=60s --timeout=10s --start-period=10s --retries=3 \
    CMD ./worker --health-check || exit 1

CMD ["./worker"]
```

### Multi-stage Optimization

```dockerfile
# Optimized Dockerfile with build cache
FROM golang:1.24-alpine AS base
RUN apk add --no-cache git ca-certificates tzdata

FROM base AS dependencies
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

FROM dependencies AS test
COPY . .
RUN go test ./...

FROM dependencies AS build
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -a -installsuffix cgo -o main ./cmd/api

FROM scratch AS runtime
COPY --from=alpine:latest /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=build /app/main /main
EXPOSE 8080
ENTRYPOINT ["/main"]
```

### Docker Compose for Development

```yaml
# docker-compose.yml
version: "3.8"

services:
  api:
    build:
      context: .
      dockerfile: Dockerfile.api
      target: development
    ports:
      - "8080:8080"
    environment:
      - DB_HOST=db
      - DB_PORT=5432
      - DB_NAME=englog_dev
      - DB_USER=postgres
      - DB_PASSWORD=devpass
      - REDIS_URL=redis://redis:6379
      - LOG_LEVEL=debug
    volumes:
      - .:/app
      - /app/vendor
    depends_on:
      db:
        condition: service_healthy
      redis:
        condition: service_healthy
    networks:
      - englog-network

  worker:
    build:
      context: .
      dockerfile: Dockerfile.worker
      target: development
    environment:
      - DB_HOST=db
      - DB_PORT=5432
      - DB_NAME=englog_dev
      - DB_USER=postgres
      - DB_PASSWORD=devpass
      - REDIS_URL=redis://redis:6379
      - OPENAI_API_KEY=${OPENAI_API_KEY}
      - LOG_LEVEL=debug
    volumes:
      - .:/app
    depends_on:
      db:
        condition: service_healthy
      redis:
        condition: service_healthy
    networks:
      - englog-network

  db:
    image: postgres:15-alpine
    environment:
      - POSTGRES_DB=englog_dev
      - POSTGRES_USER=postgres
      - POSTGRES_PASSWORD=devpass
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data
      - ./scripts/init.sql:/docker-entrypoint-initdb.d/init.sql
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U postgres"]
      interval: 10s
      timeout: 5s
      retries: 5
    networks:
      - englog-network

  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"
    volumes:
      - redis_data:/data
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 10s
      timeout: 5s
      retries: 5
    networks:
      - englog-network

  frontend:
    build:
      context: ./frontend
      dockerfile: Dockerfile
      target: development
    ports:
      - "3000:3000"
    volumes:
      - ./frontend:/app
      - /app/node_modules
    environment:
      - API_URL=http://api:8080
    depends_on:
      - api
    networks:
      - englog-network

volumes:
  postgres_data:
  redis_data:

networks:
  englog-network:
    driver: bridge
```

## ☸️ Kubernetes Deployment

### Kubernetes Manifests

```yaml
# k8s/namespace.yaml
apiVersion: v1
kind: Namespace
metadata:
  name: englog
  labels:
    name: englog
    environment: production
```

```yaml
# k8s/configmap.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: englog-config
  namespace: englog
data:
  DB_HOST: "postgres-service"
  DB_PORT: "5432"
  DB_NAME: "englog"
  REDIS_URL: "redis://redis-service:6379"
  LOG_LEVEL: "info"
  ENVIRONMENT: "production"
  API_PORT: "8080"
  WORKER_CONCURRENCY: "10"
  MAX_CONTENT_SIZE: "50000"
```

```yaml
# k8s/secrets.yaml
apiVersion: v1
kind: Secret
metadata:
  name: englog-secrets
  namespace: englog
type: Opaque
data:
  DB_PASSWORD: <base64-encoded-password>
  JWT_SECRET: <base64-encoded-jwt-secret>
  OPENAI_API_KEY: <base64-encoded-openai-key>
  AZURE_API_KEY: <base64-encoded-azure-key>
```

```yaml
# k8s/api-deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: englog-api
  namespace: englog
  labels:
    app: englog-api
    version: v1
spec:
  replicas: 3
  selector:
    matchLabels:
      app: englog-api
      version: v1
  template:
    metadata:
      labels:
        app: englog-api
        version: v1
    spec:
      containers:
        - name: api
          image: englog/api:latest
          ports:
            - containerPort: 8080
              name: http
          env:
            - name: DB_USER
              value: "postgres"
            - name: DB_PASSWORD
              valueFrom:
                secretKeyRef:
                  name: englog-secrets
                  key: DB_PASSWORD
            - name: JWT_SECRET
              valueFrom:
                secretKeyRef:
                  name: englog-secrets
                  key: JWT_SECRET
          envFrom:
            - configMapRef:
                name: englog-config
          resources:
            requests:
              memory: "128Mi"
              cpu: "100m"
            limits:
              memory: "512Mi"
              cpu: "500m"
          livenessProbe:
            httpGet:
              path: /health
              port: 8080
            initialDelaySeconds: 30
            periodSeconds: 10
            timeoutSeconds: 5
            failureThreshold: 3
          readinessProbe:
            httpGet:
              path: /ready
              port: 8080
            initialDelaySeconds: 5
            periodSeconds: 5
            timeoutSeconds: 3
            failureThreshold: 3
          securityContext:
            runAsNonRoot: true
            runAsUser: 1000
            allowPrivilegeEscalation: false
            readOnlyRootFilesystem: true
          volumeMounts:
            - name: tmp-volume
              mountPath: /tmp
      volumes:
        - name: tmp-volume
          emptyDir: {}
      serviceAccountName: englog-api-sa
      imagePullSecrets:
        - name: regcred
```

```yaml
# k8s/worker-deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: englog-worker
  namespace: englog
  labels:
    app: englog-worker
    version: v1
spec:
  replicas: 2
  selector:
    matchLabels:
      app: englog-worker
      version: v1
  template:
    metadata:
      labels:
        app: englog-worker
        version: v1
    spec:
      containers:
        - name: worker
          image: englog/worker:latest
          env:
            - name: DB_USER
              value: "postgres"
            - name: DB_PASSWORD
              valueFrom:
                secretKeyRef:
                  name: englog-secrets
                  key: DB_PASSWORD
            - name: OPENAI_API_KEY
              valueFrom:
                secretKeyRef:
                  name: englog-secrets
                  key: OPENAI_API_KEY
            - name: AZURE_API_KEY
              valueFrom:
                secretKeyRef:
                  name: englog-secrets
                  key: AZURE_API_KEY
          envFrom:
            - configMapRef:
                name: englog-config
          resources:
            requests:
              memory: "256Mi"
              cpu: "200m"
            limits:
              memory: "1Gi"
              cpu: "800m"
          securityContext:
            runAsNonRoot: true
            runAsUser: 1000
            allowPrivilegeEscalation: false
            readOnlyRootFilesystem: true
          volumeMounts:
            - name: tmp-volume
              mountPath: /tmp
      volumes:
        - name: tmp-volume
          emptyDir: {}
      serviceAccountName: englog-worker-sa
      imagePullSecrets:
        - name: regcred
```

```yaml
# k8s/service.yaml
apiVersion: v1
kind: Service
metadata:
  name: englog-api-service
  namespace: englog
  labels:
    app: englog-api
spec:
  selector:
    app: englog-api
  ports:
    - name: http
      port: 80
      targetPort: 8080
      protocol: TCP
  type: ClusterIP
```

```yaml
# k8s/ingress.yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: englog-ingress
  namespace: englog
  annotations:
    kubernetes.io/ingress.class: "nginx"
    cert-manager.io/cluster-issuer: "letsencrypt-prod"
    nginx.ingress.kubernetes.io/ssl-redirect: "true"
    nginx.ingress.kubernetes.io/force-ssl-redirect: "true"
    nginx.ingress.kubernetes.io/rate-limit: "100"
    nginx.ingress.kubernetes.io/rate-limit-window: "1m"
spec:
  tls:
    - hosts:
        - api.englog.io
        - englog.io
      secretName: englog-tls
  rules:
    - host: api.englog.io
      http:
        paths:
          - path: /
            pathType: Prefix
            backend:
              service:
                name: englog-api-service
                port:
                  number: 80
    - host: englog.io
      http:
        paths:
          - path: /
            pathType: Prefix
            backend:
              service:
                name: englog-frontend-service
                port:
                  number: 80
```

### Horizontal Pod Autoscaler

```yaml
# k8s/hpa.yaml
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: englog-api-hpa
  namespace: englog
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: englog-api
  minReplicas: 3
  maxReplicas: 20
  metrics:
    - type: Resource
      resource:
        name: cpu
        target:
          type: Utilization
          averageUtilization: 70
    - type: Resource
      resource:
        name: memory
        target:
          type: Utilization
          averageUtilization: 80
  behavior:
    scaleDown:
      stabilizationWindowSeconds: 300
      policies:
        - type: Percent
          value: 50
          periodSeconds: 60
    scaleUp:
      stabilizationWindowSeconds: 60
      policies:
        - type: Percent
          value: 100
          periodSeconds: 30
```

## 🚀 CI/CD Pipeline

### GitHub Actions Workflow

```yaml
# .github/workflows/deploy.yml
name: Build and Deploy

on:
  push:
    branches: [main, develop]
  pull_request:
    branches: [main]

env:
  REGISTRY: ghcr.io
  IMAGE_NAME: ${{ github.repository }}

jobs:
  test:
    runs-on: ubuntu-latest
    services:
      postgres:
        image: postgres:15
        env:
          POSTGRES_PASSWORD: testpass
          POSTGRES_USER: testuser
          POSTGRES_DB: englog_test
        options: >-
          --health-cmd pg_isready
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5

    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Set up Go
        uses: actions/setup-go@v4
        with:
          go-version: "1.24"

      - name: Cache Go modules
        uses: actions/cache@v3
        with:
          path: ~/go/pkg/mod
          key: ${{ runner.os }}-go-${{ hashFiles('**/go.sum') }}

      - name: Run tests
        run: go test -v -race -coverprofile=coverage.out ./...
        env:
          DB_URL: postgres://testuser:testpass@localhost:5432/englog_test?sslmode=disable

      - name: Upload coverage
        uses: codecov/codecov-action@v3

  build-and-push:
    runs-on: ubuntu-latest
    needs: test
    if: github.event_name == 'push'

    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Set up Docker Buildx
        uses: docker/setup-buildx-action@v3

      - name: Log in to Container Registry
        uses: docker/login-action@v3
        with:
          registry: ${{ env.REGISTRY }}
          username: ${{ github.actor }}
          password: ${{ secrets.GITHUB_TOKEN }}

      - name: Extract metadata
        id: meta
        uses: docker/metadata-action@v5
        with:
          images: ${{ env.REGISTRY }}/${{ env.IMAGE_NAME }}
          tags: |
            type=ref,event=branch
            type=ref,event=pr
            type=sha,prefix={{branch}}-
            type=raw,value=latest,enable={{is_default_branch}}

      - name: Build and push API image
        uses: docker/build-push-action@v5
        with:
          context: .
          file: ./Dockerfile.api
          push: true
          tags: ${{ steps.meta.outputs.tags }}
          labels: ${{ steps.meta.outputs.labels }}
          cache-from: type=gha
          cache-to: type=gha,mode=max

      - name: Build and push Worker image
        uses: docker/build-push-action@v5
        with:
          context: .
          file: ./Dockerfile.worker
          push: true
          tags: ${{ env.REGISTRY }}/${{ env.IMAGE_NAME }}/worker:${{ github.sha }}
          cache-from: type=gha
          cache-to: type=gha,mode=max

  deploy-staging:
    runs-on: ubuntu-latest
    needs: build-and-push
    if: github.ref == 'refs/heads/develop'
    environment: staging

    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Configure AWS credentials
        uses: aws-actions/configure-aws-credentials@v4
        with:
          aws-access-key-id: ${{ secrets.AWS_ACCESS_KEY_ID }}
          aws-secret-access-key: ${{ secrets.AWS_SECRET_ACCESS_KEY }}
          aws-region: us-west-2

      - name: Update kubeconfig
        run: aws eks update-kubeconfig --name englog-staging --region us-west-2

      - name: Deploy to staging
        run: |
          # Update image tags in staging manifests
          sed -i "s|image: englog/api:.*|image: ${{ env.REGISTRY }}/${{ env.IMAGE_NAME }}:${{ github.sha }}|" k8s/staging/api-deployment.yaml
          sed -i "s|image: englog/worker:.*|image: ${{ env.REGISTRY }}/${{ env.IMAGE_NAME }}/worker:${{ github.sha }}|" k8s/staging/worker-deployment.yaml

          # Apply manifests
          kubectl apply -f k8s/staging/

          # Wait for rollout
          kubectl rollout status deployment/englog-api -n englog-staging --timeout=600s
          kubectl rollout status deployment/englog-worker -n englog-staging --timeout=600s

      - name: Run smoke tests
        run: |
          # Wait for service to be ready
          kubectl wait --for=condition=ready pod -l app=englog-api -n englog-staging --timeout=300s

          # Get service URL
          SERVICE_URL=$(kubectl get service englog-api-service -n englog-staging -o jsonpath='{.status.loadBalancer.ingress[0].hostname}')

          # Run basic health check
          curl -f http://$SERVICE_URL/health || exit 1

  deploy-production:
    runs-on: ubuntu-latest
    needs: [build-and-push, deploy-staging]
    if: github.ref == 'refs/heads/main'
    environment: production

    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Configure AWS credentials
        uses: aws-actions/configure-aws-credentials@v4
        with:
          aws-access-key-id: ${{ secrets.AWS_ACCESS_KEY_ID }}
          aws-secret-access-key: ${{ secrets.AWS_SECRET_ACCESS_KEY }}
          aws-region: us-west-2

      - name: Update kubeconfig
        run: aws eks update-kubeconfig --name englog-production --region us-west-2

      - name: Deploy to production with blue-green
        run: |
          # Create blue-green deployment script
          ./scripts/blue-green-deploy.sh \
            --image "${{ env.REGISTRY }}/${{ env.IMAGE_NAME }}:${{ github.sha }}" \
            --worker-image "${{ env.REGISTRY }}/${{ env.IMAGE_NAME }}/worker:${{ github.sha }}" \
            --namespace englog \
            --timeout 600

      - name: Run production health checks
        run: |
          # Comprehensive health checks
          ./scripts/health-check.sh --environment production --timeout 300

      - name: Notify deployment
        uses: 8398a7/action-slack@v3
        with:
          status: ${{ job.status }}
          channel: "#deployments"
          webhook_url: ${{ secrets.SLACK_WEBHOOK }}
```

### Blue-Green Deployment Script

```bash
#!/bin/bash
# scripts/blue-green-deploy.sh

set -euo pipefail

# Configuration
NAMESPACE="englog"
TIMEOUT=600
IMAGE=""
WORKER_IMAGE=""

# Parse arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --image)
            IMAGE="$2"
            shift 2
            ;;
        --worker-image)
            WORKER_IMAGE="$2"
            shift 2
            ;;
        --namespace)
            NAMESPACE="$2"
            shift 2
            ;;
        --timeout)
            TIMEOUT="$2"
            shift 2
            ;;
        *)
            echo "Unknown option $1"
            exit 1
            ;;
    esac
done

# Validate inputs
if [[ -z "$IMAGE" || -z "$WORKER_IMAGE" ]]; then
    echo "Error: --image and --worker-image are required"
    exit 1
fi

echo "Starting blue-green deployment..."
echo "Namespace: $NAMESPACE"
echo "API Image: $IMAGE"
echo "Worker Image: $WORKER_IMAGE"
echo "Timeout: $TIMEOUT seconds"

# Get current deployment
CURRENT_DEPLOYMENT=$(kubectl get deployment englog-api -n $NAMESPACE -o jsonpath='{.metadata.labels.version}' || echo "v1")

# Determine next version
if [[ "$CURRENT_DEPLOYMENT" == "v1" ]]; then
    NEXT_VERSION="v2"
else
    NEXT_VERSION="v1"
fi

echo "Current version: $CURRENT_DEPLOYMENT"
echo "Deploying version: $NEXT_VERSION"

# Create new deployment manifests
envsubst < k8s/api-deployment-template.yaml > k8s/api-deployment-$NEXT_VERSION.yaml
envsubst < k8s/worker-deployment-template.yaml > k8s/worker-deployment-$NEXT_VERSION.yaml

# Update image tags and version labels
sed -i "s|{{IMAGE}}|$IMAGE|g" k8s/api-deployment-$NEXT_VERSION.yaml
sed -i "s|{{WORKER_IMAGE}}|$WORKER_IMAGE|g" k8s/worker-deployment-$NEXT_VERSION.yaml
sed -i "s|{{VERSION}}|$NEXT_VERSION|g" k8s/api-deployment-$NEXT_VERSION.yaml
sed -i "s|{{VERSION}}|$NEXT_VERSION|g" k8s/worker-deployment-$NEXT_VERSION.yaml

# Deploy new version
echo "Deploying new version..."
kubectl apply -f k8s/api-deployment-$NEXT_VERSION.yaml
kubectl apply -f k8s/worker-deployment-$NEXT_VERSION.yaml

# Wait for new deployment to be ready
echo "Waiting for new deployment to be ready..."
kubectl rollout status deployment/englog-api-$NEXT_VERSION -n $NAMESPACE --timeout=${TIMEOUT}s
kubectl rollout status deployment/englog-worker-$NEXT_VERSION -n $NAMESPACE --timeout=${TIMEOUT}s

# Run health checks on new version
echo "Running health checks on new version..."
NEW_PODS=$(kubectl get pods -l app=englog-api,version=$NEXT_VERSION -n $NAMESPACE -o jsonpath='{.items[*].metadata.name}')

for pod in $NEW_PODS; do
    echo "Checking health of pod: $pod"
    kubectl exec $pod -n $NAMESPACE -- wget --quiet --tries=1 --timeout=10 --spider http://localhost:8080/health
done

# Switch traffic to new version
echo "Switching traffic to new version..."
kubectl patch service englog-api-service -n $NAMESPACE -p '{"spec":{"selector":{"version":"'$NEXT_VERSION'"}}}'

# Wait a bit and verify traffic is flowing
sleep 30

# Final health check with external endpoint
echo "Running final health check..."
EXTERNAL_URL=$(kubectl get ingress englog-ingress -n $NAMESPACE -o jsonpath='{.spec.rules[0].host}')
curl -f https://$EXTERNAL_URL/health

# Clean up old deployment
echo "Cleaning up old deployment..."
kubectl delete deployment englog-api-$CURRENT_DEPLOYMENT -n $NAMESPACE --ignore-not-found=true
kubectl delete deployment englog-worker-$CURRENT_DEPLOYMENT -n $NAMESPACE --ignore-not-found=true

# Clean up temporary files
rm -f k8s/api-deployment-$NEXT_VERSION.yaml
rm -f k8s/worker-deployment-$NEXT_VERSION.yaml

echo "Blue-green deployment completed successfully!"
echo "New version $NEXT_VERSION is now live"
```

## 🌍 Environment Management

### Environment Configuration

```go
// config/environment.go
package config

import (
    "fmt"
    "os"
    "strconv"
    "time"
)

type Environment string

const (
    Development Environment = "development"
    Staging     Environment = "staging"
    Production  Environment = "production"
)

type Config struct {
    Environment Environment

    // Server Configuration
    Port         int
    ReadTimeout  time.Duration
    WriteTimeout time.Duration

    // Database Configuration
    DatabaseURL      string
    DatabaseMaxConns int
    DatabaseTimeout  time.Duration

    // Redis Configuration
    RedisURL        string
    RedisMaxRetries int

    // Security Configuration
    JWTSecret            string
    PasswordSaltRounds   int
    SessionTimeout       time.Duration
    RateLimitRPS         int

    // AI Service Configuration
    OpenAIAPIKey    string
    AzureAPIKey     string
    OllamaEndpoint  string
    AITimeout       time.Duration

    // Monitoring Configuration
    MetricsEnabled    bool
    TracingEnabled    bool
    LogLevel          string

    // Feature Flags
    Features FeatureFlags
}

type FeatureFlags struct {
    EnableAdvancedAI    bool
    EnableGroupSharing  bool
    EnableExportFeature bool
    EnableBetaFeatures  bool
}

func LoadConfig() (*Config, error) {
    env := Environment(getEnvString("ENVIRONMENT", string(Development)))

    config := &Config{
        Environment: env,

        // Server Configuration
        Port:         getEnvInt("PORT", 8080),
        ReadTimeout:  getEnvDuration("READ_TIMEOUT", 30*time.Second),
        WriteTimeout: getEnvDuration("WRITE_TIMEOUT", 30*time.Second),

        // Database Configuration
        DatabaseURL:      getEnvString("DATABASE_URL", ""),
        DatabaseMaxConns: getEnvInt("DB_MAX_CONNS", 25),
        DatabaseTimeout:  getEnvDuration("DB_TIMEOUT", 10*time.Second),

        // Redis Configuration
        RedisURL:        getEnvString("REDIS_URL", "redis://localhost:6379"),
        RedisMaxRetries: getEnvInt("REDIS_MAX_RETRIES", 3),

        // Security Configuration
        JWTSecret:            getEnvString("JWT_SECRET", ""),
        PasswordSaltRounds:   getEnvInt("PASSWORD_SALT_ROUNDS", 12),
        SessionTimeout:       getEnvDuration("SESSION_TIMEOUT", 24*time.Hour),
        RateLimitRPS:         getEnvInt("RATE_LIMIT_RPS", 100),

        // AI Configuration
        OpenAIAPIKey:   getEnvString("OPENAI_API_KEY", ""),
        AzureAPIKey:    getEnvString("AZURE_API_KEY", ""),
        OllamaEndpoint: getEnvString("OLLAMA_ENDPOINT", "http://localhost:11434"),
        AITimeout:      getEnvDuration("AI_TIMEOUT", 30*time.Second),

        // Monitoring Configuration
        MetricsEnabled: getEnvBool("METRICS_ENABLED", true),
        TracingEnabled: getEnvBool("TRACING_ENABLED", env == Production),
        LogLevel:       getEnvString("LOG_LEVEL", getDefaultLogLevel(env)),

        // Feature Flags
        Features: FeatureFlags{
            EnableAdvancedAI:    getEnvBool("FEATURE_ADVANCED_AI", env == Production),
            EnableGroupSharing:  getEnvBool("FEATURE_GROUP_SHARING", true),
            EnableExportFeature: getEnvBool("FEATURE_EXPORT", true),
            EnableBetaFeatures:  getEnvBool("FEATURE_BETA", env != Production),
        },
    }

    // Validate configuration
    if err := config.Validate(); err != nil {
        return nil, fmt.Errorf("configuration validation failed: %w", err)
    }

    return config, nil
}

func (c *Config) Validate() error {
    if c.DatabaseURL == "" {
        return fmt.Errorf("DATABASE_URL is required")
    }

    if c.JWTSecret == "" {
        return fmt.Errorf("JWT_SECRET is required")
    }

    if c.Environment == Production {
        if c.OpenAIAPIKey == "" && c.AzureAPIKey == "" {
            return fmt.Errorf("at least one AI API key is required in production")
        }
    }

    return nil
}

func getEnvString(key, defaultValue string) string {
    if value := os.Getenv(key); value != "" {
        return value
    }
    return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
    if value := os.Getenv(key); value != "" {
        if intValue, err := strconv.Atoi(value); err == nil {
            return intValue
        }
    }
    return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
    if value := os.Getenv(key); value != "" {
        if boolValue, err := strconv.ParseBool(value); err == nil {
            return boolValue
        }
    }
    return defaultValue
}

func getEnvDuration(key string, defaultValue time.Duration) time.Duration {
    if value := os.Getenv(key); value != "" {
        if duration, err := time.ParseDuration(value); err == nil {
            return duration
        }
    }
    return defaultValue
}

func getDefaultLogLevel(env Environment) string {
    switch env {
    case Development:
        return "debug"
    case Staging:
        return "info"
    case Production:
        return "warn"
    default:
        return "info"
    }
}
```

### Secret Management with AWS Secrets Manager

```go
// config/secrets.go
package config

import (
    "context"
    "encoding/json"
    "fmt"

    "github.com/aws/aws-sdk-go-v2/aws"
    "github.com/aws/aws-sdk-go-v2/config"
    "github.com/aws/aws-sdk-go-v2/service/secretsmanager"
)

type SecretsManager struct {
    client *secretsmanager.Client
    region string
}

func NewSecretsManager(region string) (*SecretsManager, error) {
    cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(region))
    if err != nil {
        return nil, fmt.Errorf("failed to load AWS config: %w", err)
    }

    return &SecretsManager{
        client: secretsmanager.NewFromConfig(cfg),
        region: region,
    }, nil
}

func (sm *SecretsManager) GetSecret(secretId string) (map[string]string, error) {
    input := &secretsmanager.GetSecretValueInput{
        SecretId: aws.String(secretId),
    }

    result, err := sm.client.GetSecretValue(context.TODO(), input)
    if err != nil {
        return nil, fmt.Errorf("failed to get secret %s: %w", secretId, err)
    }

    var secrets map[string]string
    err = json.Unmarshal([]byte(*result.SecretString), &secrets)
    if err != nil {
        return nil, fmt.Errorf("failed to parse secret %s: %w", secretId, err)
    }

    return secrets, nil
}

// Load secrets into environment
func LoadSecretsToEnv() error {
    env := os.Getenv("ENVIRONMENT")
    if env == "" {
        env = "development"
    }

    // Skip secret loading in development
    if env == "development" {
        return nil
    }

    region := os.Getenv("AWS_REGION")
    if region == "" {
        region = "us-west-2"
    }

    sm, err := NewSecretsManager(region)
    if err != nil {
        return err
    }

    // Load application secrets
    secretId := fmt.Sprintf("englog/%s/app", env)
    secrets, err := sm.GetSecret(secretId)
    if err != nil {
        return err
    }

    // Set environment variables
    for key, value := range secrets {
        os.Setenv(key, value)
    }

    return nil
}
```

## 🏥 Health Checks and Monitoring

### Health Check Implementation

```go
// health/checker.go
package health

import (
    "context"
    "database/sql"
    "fmt"
    "net/http"
    "time"

    "github.com/gin-gonic/gin"
    "github.com/redis/go-redis/v9"
)

type HealthChecker struct {
    db    *sql.DB
    redis *redis.Client
}

type HealthStatus struct {
    Status      string                 `json:"status"`
    Timestamp   time.Time              `json:"timestamp"`
    Version     string                 `json:"version"`
    Environment string                 `json:"environment"`
    Checks      map[string]CheckResult `json:"checks"`
}

type CheckResult struct {
    Status    string        `json:"status"`
    Message   string        `json:"message,omitempty"`
    Duration  time.Duration `json:"duration"`
    Timestamp time.Time     `json:"timestamp"`
}

func NewHealthChecker(db *sql.DB, redis *redis.Client) *HealthChecker {
    return &HealthChecker{
        db:    db,
        redis: redis,
    }
}

func (hc *HealthChecker) HealthHandler() gin.HandlerFunc {
    return func(c *gin.Context) {
        status := hc.CheckHealth()

        httpStatus := http.StatusOK
        if status.Status != "healthy" {
            httpStatus = http.StatusServiceUnavailable
        }

        c.JSON(httpStatus, status)
    }
}

func (hc *HealthChecker) ReadinessHandler() gin.HandlerFunc {
    return func(c *gin.Context) {
        status := hc.CheckReadiness()

        httpStatus := http.StatusOK
        if status.Status != "ready" {
            httpStatus = http.StatusServiceUnavailable
        }

        c.JSON(httpStatus, status)
    }
}

func (hc *HealthChecker) CheckHealth() HealthStatus {
    checks := make(map[string]CheckResult)
    overall := "healthy"

    // Check database
    dbResult := hc.checkDatabase()
    checks["database"] = dbResult
    if dbResult.Status != "healthy" {
        overall = "unhealthy"
    }

    // Check Redis
    redisResult := hc.checkRedis()
    checks["redis"] = redisResult
    if redisResult.Status != "healthy" {
        overall = "unhealthy"
    }

    // Check external dependencies (optional)
    aiResult := hc.checkAIServices()
    checks["ai_services"] = aiResult
    if aiResult.Status != "healthy" {
        // AI services are not critical for basic operation
        if overall == "healthy" {
            overall = "degraded"
        }
    }

    return HealthStatus{
        Status:      overall,
        Timestamp:   time.Now(),
        Version:     getVersion(),
        Environment: getEnvironment(),
        Checks:      checks,
    }
}

func (hc *HealthChecker) CheckReadiness() HealthStatus {
    // Readiness is more strict than health
    checks := make(map[string]CheckResult)
    overall := "ready"

    // All critical services must be ready
    dbResult := hc.checkDatabase()
    checks["database"] = dbResult
    if dbResult.Status != "healthy" {
        overall = "not_ready"
    }

    redisResult := hc.checkRedis()
    checks["redis"] = redisResult
    if redisResult.Status != "healthy" {
        overall = "not_ready"
    }

    return HealthStatus{
        Status:      overall,
        Timestamp:   time.Now(),
        Version:     getVersion(),
        Environment: getEnvironment(),
        Checks:      checks,
    }
}

func (hc *HealthChecker) checkDatabase() CheckResult {
    start := time.Now()
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    if err := hc.db.PingContext(ctx); err != nil {
        return CheckResult{
            Status:    "unhealthy",
            Message:   fmt.Sprintf("Database connection failed: %v", err),
            Duration:  time.Since(start),
            Timestamp: time.Now(),
        }
    }

    // Test a simple query
    var result int
    err := hc.db.QueryRowContext(ctx, "SELECT 1").Scan(&result)
    if err != nil {
        return CheckResult{
            Status:    "unhealthy",
            Message:   fmt.Sprintf("Database query failed: %v", err),
            Duration:  time.Since(start),
            Timestamp: time.Now(),
        }
    }

    return CheckResult{
        Status:    "healthy",
        Duration:  time.Since(start),
        Timestamp: time.Now(),
    }
}

func (hc *HealthChecker) checkRedis() CheckResult {
    start := time.Now()
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    pong, err := hc.redis.Ping(ctx).Result()
    if err != nil || pong != "PONG" {
        return CheckResult{
            Status:    "unhealthy",
            Message:   fmt.Sprintf("Redis ping failed: %v", err),
            Duration:  time.Since(start),
            Timestamp: time.Now(),
        }
    }

    return CheckResult{
        Status:    "healthy",
        Duration:  time.Since(start),
        Timestamp: time.Now(),
    }
}

func (hc *HealthChecker) checkAIServices() CheckResult {
    start := time.Now()

    // This is a simplified check - in practice, you might check
    // if API keys are valid or test a lightweight endpoint
    apiKey := os.Getenv("OPENAI_API_KEY")
    if apiKey == "" {
        return CheckResult{
            Status:    "unhealthy",
            Message:   "OpenAI API key not configured",
            Duration:  time.Since(start),
            Timestamp: time.Now(),
        }
    }

    return CheckResult{
        Status:    "healthy",
        Duration:  time.Since(start),
        Timestamp: time.Now(),
    }
}

func getVersion() string {
    // This would typically be injected at build time
    return os.Getenv("APP_VERSION")
}

func getEnvironment() string {
    return os.Getenv("ENVIRONMENT")
}
```

---

## 🔗 Related Documents

- **[Security](./SECURITY.md)** - Security considerations for deployment
- **[Monitoring](./MONITORING.md)** - Production monitoring and alerting
- **[Testing](./TESTING.md)** - Testing in deployment pipeline
- **[API Service Design](../components/API_SERVICE.md)** - Service architecture
- **[Database Design](../components/DATABASE.md)** - Database deployment considerations

---

**Document Status:** 🚧 In Progress
**Next Review:** 2025-09-04
**Last Updated:** 2025-08-04

---

_This document outlines the comprehensive deployment strategy for the EngLog system. It serves as the technical specification for implementing reliable, scalable, and maintainable deployment processes across all environments._
