# ---------- builder ----------
FROM swr.cn-east-3.myhuaweicloud.com/kubesre/docker.io/golang:1.24 AS builder

WORKDIR /src

# 国内代理（可选）
ENV GOPROXY=https://goproxy.cn,https://goproxy.io,direct
# ENV GOSUMDB=off  # 网络极差时可考虑关闭校验（不推荐）

# 仅复制依赖声明，加快缓存
COPY go.mod go.sum ./
RUN go mod download

# 再复制源码（以项目根为基准）
COPY . .

# 编译（静态二进制，适配 alpine）
ENV CGO_ENABLED=0 GOOS=linux GOARCH=amd64
ARG BIN_NAME=k8s_operation
RUN go deploy -trimpath -ldflags="-s -w" -o /out/${BIN_NAME} ./cmd/k8soperation

# ---------- runtime ----------
FROM swr.cn-east-3.myhuaweicloud.com/kubesre/docker.io/alpine:3.20

RUN apk add --no-cache ca-certificates tzdata && \
    addgroup -S app && adduser -S app -G app

WORKDIR /app
RUN mkdir -p /app/storage/logs /app/configs

COPY --from=builder /out/k8s_operation /app/k8s_operation
# 如需把配置打进镜像（否则运行时挂载）
COPY configs/config.yaml /app/configs/config.yaml

RUN chown -R app:app /app
USER app

ENV GIN_MODE=release
ENV APP_CONFIG=/app/configs/config.yaml
EXPOSE 8080
ENTRYPOINT ["/app/k8s_operation"]
