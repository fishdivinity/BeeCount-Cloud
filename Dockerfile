# 多阶段构建Dockerfile - 支持CGO和SQLite

# 第一阶段：构建阶段
FROM golang:1.25.5-alpine3.21 AS builder

# 安装构建依赖
RUN apk add --no-cache \
    git \
    gcc \
    musl-dev \
    sqlite-dev \
    sqlite-libs

WORKDIR /app

# 复制依赖文件
COPY go.mod go.sum ./
RUN go mod download

# 复制源代码
COPY . .

# 构建应用程序（启用CGO以支持SQLite）
RUN CGO_ENABLED=1 GOOS=linux go build \
    -a \
    -installsuffix cgo \
    -ldflags="-w -s" \
    -o server ./cmd/server

# 第二阶段：运行阶段
FROM alpine:3.21.5

# 安装运行时依赖
RUN apk --no-cache add \
    ca-certificates \
    tzdata \
    sqlite-libs \
    wget

WORKDIR /app

# 从构建阶段复制二进制文件
COPY --from=builder /app/server .

# 复制 ReDoc 静态文件
COPY --from=builder /app/web/redoc ./web/redoc

# 创建必要的目录
RUN mkdir -p /app/data /app/logs /app/uploads

# 设置权限
RUN chmod +x /app/server

# 暴露端口
EXPOSE 8080

# 健康检查
HEALTHCHECK --interval=30s --timeout=10s --start-period=40s --retries=3 \
    CMD wget --quiet --tries=1 --spider http://localhost:8080/health || exit 1

# 运行应用程序
CMD ["./server"]
