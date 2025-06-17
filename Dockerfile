# ---------- build ----------
FROM golang:1.22.6 AS builder
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
# полностью статический бинарь
RUN CGO_ENABLED=0 go build -trimpath -ldflags "-s -w" -o bot ./cmd/bot

# ---------- runtime ----------
FROM gcr.io/distroless/static AS final
COPY --from=builder /src/bot /bot
ENTRYPOINT ["/bot"]
