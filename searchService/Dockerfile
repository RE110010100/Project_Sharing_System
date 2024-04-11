# Use a minimal base image
FROM golang:alpine3.19 AS build

# Set the working directory
WORKDIR /app

# Copy the Go module files
COPY go.mod go.sum ./

# Download and install dependencies
RUN go mod download

# Copy the rest of the application code
COPY . .

# Build the Go application
RUN go build -o searchService .

# Use a lightweight base image
FROM alpine:latest

# Set the working directory
WORKDIR /app

# Copy the binary from the build stage
COPY --from=build /app/searchService .

# Expose the port your application listens on
EXPOSE 8083

# Command to run the executable
CMD ["./searchService"]
