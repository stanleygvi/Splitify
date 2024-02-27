# Use an official Go runtime as the base image
FROM golang:1.20 as builder

# Set the working directory inside the container
WORKDIR /app

# Copy the go.mod and go.sum files and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the entire application source inside the container
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -v -o myapp

# Use a minimal image for the runtime
FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/

# Copy the binary from the builder stage
COPY --from=builder /app/myapp .

# Expose the application on port 8080 (or whatever port your app uses)
EXPOSE 8080

# Run the application
CMD ["./myapp"]
