# Use a base Go image for building
FROM golang:alpine AS builder

# Install git which may be required by some Go packages
RUN apk update && apk add --no-cache git

# Set the working directory
WORKDIR /build

# Copy the application source code
COPY . .
# Download dependencies
RUN go mod download
# Install swag using go get
RUN go install github.com/swaggo/swag/cmd/swag@latest
# Generate Swagger documentation using swag
RUN swag init
# Build your application
RUN go build -o main .
# Set the working directory for the final image
WORKDIR /dist
# Copy the built binary to the final directory
RUN cp /build/main .

# Create the final image
FROM scratch
# Copy the binary from the builder stage
COPY --from=builder /dist/main /main
# Set the entry point
ENTRYPOINT ["/main"]
