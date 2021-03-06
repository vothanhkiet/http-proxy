############################
# STEP 1 build executable binary
############################
FROM golang:1.13.5-alpine3.10 as builder

# Install git + SSL ca certificates.
# Git is required for fetching the dependencies.
# Ca-certificates is required to call HTTPS endpoints.
RUN apk update && apk add --no-cache git ca-certificates tzdata && update-ca-certificates

# Create appuser
RUN adduser -D -g '' appuser

WORKDIR $GOPATH/src/github.com/vothanhkiet/http-proxy
COPY ./src .

# Fetch dependencies.
RUN go get -d -v

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -a -installsuffix cgo -o /go/bin/proxy .
RUN apk add upx --no-cache --repository http://dl-cdn.alpinelinux.org/alpine/edge/community
RUN upx /go/bin/proxy
RUN ls /go/bin

############################
# STEP 2 build a small image
############################
FROM scratch
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /etc/passwd /etc/passwd
USER appuser
# Copy our static executable.
COPY --from=builder /go/bin/proxy proxy
EXPOSE 8080
# Run the hello binary.
ENTRYPOINT ["./proxy"]