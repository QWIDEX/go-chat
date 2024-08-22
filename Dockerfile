FROM golang

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . ./

RUN CGO_ENABLED=0 GOOS=linux go build -o /main cmd/api/main.go

EXPOSE 8080

CMD [ "/main" ]