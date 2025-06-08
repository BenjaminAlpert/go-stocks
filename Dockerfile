FROM --platform=linux/amd64 golang:1.23 AS builder

WORKDIR /app

COPY . .

RUN go mod download

RUN CGO_ENABLED=0 GOOS=linux go build -o /stocks

FROM builder

COPY --from=builder /stocks /stocks

EXPOSE 8080

CMD ["/stocks"]