# Build stage
FROM golang:1.22.2-alpine AS build-base

WORKDIR /app

# Copy source code and dependencies
COPY go.mod go.sum ./
RUN go mod download
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o ./assessment-tax main.go


# Runtime stage
FROM alpine:3.16.2

WORKDIR /app

# Copy built binaries from build stage
COPY --from=build-base /app/assessment-tax ./assessment-tax

# Start the application
CMD ["/app/assessment-tax"]
