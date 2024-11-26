# Use the specific Go version 1.22.5 image as the base
FROM golang:1.22.5-alpine

# Set the current working directory inside the container
WORKDIR /app

# Copy the Go module files and download dependencies
COPY go.mod go.sum ./

# Fetch the Go modules dependencies
RUN go mod tidy

# Copy the entire Go application into the container
COPY . .

# Build the Go application
RUN go build -o myapp .

# Expose port 5001 (adjust as necessary for your app)
EXPOSE 5001

# Command to run the application with the desired parameters
CMD ["go", "run", "main.go", "node", "-n", "5001"]
