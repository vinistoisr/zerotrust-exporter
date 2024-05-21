# Use an official Golang runtime as a parent image
FROM golang:1.18-alpine

# Install git for go module fetching
RUN apk add --no-cache git

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy go.mod and go.sum files
COPY go.mod go.sum ./

# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download

# Copy the source from the current directory to the Working Directory inside the container
COPY . .

# Build the Go app
RUN go build -o zerotrust-exporter .

# Expose port 9184 to the outside world
EXPOSE 9184

# Command to run the executable with environment variables
CMD ["sh", "-c", "./zerotrust-exporter -apikey=${API_KEY} -accountid=${ACCOUNT_ID} -debug=${DEBUG} -devices=${DEVICES} -users=${USERS} -tunnels=${TUNNELS} -interface=${INTERFACE} -port=${PORT}"]
