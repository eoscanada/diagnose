FROM alpine:latest
RUN apk update && apk add ca-certificates curl && rm -rf /var/cache/apk/*
COPY /diagnose /app/diagnose
COPY *html /app/
WORKDIR /app
