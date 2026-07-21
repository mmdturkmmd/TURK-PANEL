FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY . .

# فعال کردن CGO برای sqlite
ENV CGO_ENABLED=1

RUN apk add --no-cache gcc musl-dev sqlite-dev

RUN go mod download

RUN go build -o x-ui main.go


FROM alpine:3.19

RUN apk add --no-cache \
    bash \
    curl \
    ca-certificates \
    unzip \
    tzdata \
    sqlite \
    libc6-compat \
    && ln -sf /usr/share/zoneinfo/Asia/Tehran /etc/localtime


WORKDIR /app

COPY --from=builder /app/x-ui /app/x-ui
COPY --from=builder /app/web /app/web
COPY --from=builder /app/config /app/config
COPY start.sh /start.sh


RUN mkdir -p /app/bin \
    && curl -L https://github.com/XTLS/Xray-core/releases/download/v26.7.11/Xray-linux-64.zip \
    -o /tmp/xray.zip \
    && unzip /tmp/xray.zip -d /tmp/xray \
    && mv /tmp/xray/xray /app/bin/xray-linux-amd64 \
    && chmod +x /app/bin/xray-linux-amd64 \
    && rm -rf /tmp/xray /tmp/xray.zip

# دانلود GeoIP و GeoSite
RUN curl -L https://github.com/Loyalsoldier/v2ray-rules-dat/releases/latest/download/geoip.dat \
    -o /app/bin/geoip.dat \
    && curl -L https://github.com/Loyalsoldier/v2ray-rules-dat/releases/latest/download/geosite.dat \
    -o /app/bin/geosite.dat

RUN chmod +x /app/x-ui /start.sh

EXPOSE 8080

CMD ["./x-ui"]
