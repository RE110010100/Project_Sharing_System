# Use a minimal base image
FROM golang:1.22.1 AS build

# Set the current working directory inside the container
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download all dependencies
RUN go mod download

# Copy the source code into the container
COPY . .

# Build the Go service
RUN go build -o projectService .

# Start a new stage from scratch
FROM alpine:latest  

WORKDIR /root/

# Copy the Pre-built binary file from the previous stage
COPY --from=build /app/projectService .

# Expose port 8080 to the outside world
EXPOSE 8082

# Command to run the executable
CMD ["./projectService"]
