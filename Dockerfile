FROM golang:1.20-alpine3.17 as builder
LABEL stage=go-builder
WORKDIR /app/
COPY ./ ./
RUN go build -ldflags="-s -w" -o ./bin/typer main.go

FROM alpine:edge
LABEL AUTHOR="youtiaoguagua"
WORKDIR /typer
VOLUME /typer/data
ENV TERM=xterm-256color
COPY --from=builder /app/bin/typer ./
EXPOSE 7788
ENTRYPOINT ["./typer"]