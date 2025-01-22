FROM golang:1.23.4-alpine3.21 AS builder

# `boilerplate` should be replaced with your project name
WORKDIR /go/src/code

# Copy all the Code and stuff to compile everything
COPY go.* ./
# Downloads all the dependencies in advance (could be left out, but it's more clear this way)
RUN go mod download
COPY . .
# Builds the application as a staticly linked one, to allow it to run on alpine
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -o app .


FROM alpine:3.21.2

# Create and enable user
ARG USER_NAME=alpine
ARG USER_ID=1000
RUN addgroup ${USER_NAME} && adduser --system -u ${USER_ID} -g ${USER_NAME} --no-create-home ${USER_NAME}
USER ${USER_NAME}

# Moving the binary to the 'final Image' to make it smaller
WORKDIR /app
COPY --from=builder /go/src/code/app .
CMD ["sh", "-c", "./app"]