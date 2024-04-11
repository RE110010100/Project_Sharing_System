# Use a minimal base image
FROM golang:alpine3.19 AS build

# Set the current working directory inside the container
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download all dependencies
RUN go mod download

# Copy the source code into the container
COPY . .

# Build the Go service
RUN go build -o notificationService .

# Start a new stage from scratch
FROM alpine:latest  

WORKDIR /root/

# Copy the Pre-built binary file from the previous stage
COPY --from=build /app/notificationService .

# Expose port 8085 to the outside world
EXPOSE 8085

# Command to run the executable
CMD ["./notificationService"]
