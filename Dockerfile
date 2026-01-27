# BeeCount Cloud 微服务架构 Dockerfile
FROM golang:1.25.6-alpine3.22 AS builder

# 安装依赖
RUN apk add --no-cache git gcc musl-dev

# 设置工作目录
WORKDIR /app

# 设置Go代理，确保依赖下载稳定
ENV GOPROXY=https://goproxy.cn,https://proxy.golang.org,direct
ENV GOPRIVATE=
# 设置编译参数，启用大文件支持
ENV CGO_CFLAGS="-D_LARGEFILE64_SOURCE -D_FILE_OFFSET_BITS=64"

# 复制go.mod文件
COPY go.mod go.sum ./

# 安装依赖
RUN go mod tidy

# 复制代码
COPY . .

# 构建所有服务
RUN mkdir -p bin

# 定义服务到可执行文件的映射
RUN echo "gateway=gateway" > service_mapping.txt && \
    echo "config=config" >> service_mapping.txt && \
    echo "auth=auth" >> service_mapping.txt && \
    echo "business=business" >> service_mapping.txt && \
    echo "storage=storage" >> service_mapping.txt && \
    echo "log=log" >> service_mapping.txt && \
    echo "firewall=firewall" >> service_mapping.txt && \
    echo "beecount=BeeCount-Cloud" >> service_mapping.txt

# 构建所有服务
RUN while IFS="=" read -r service executable; do \
    echo "Building $service -> $executable"; \
    cd services/$service && go build -ldflags="-s -w" -o /app/bin/$executable ./cmd; \
    cd /app; \
done < service_mapping.txt

# 最终镜像
FROM alpine:3.22.2

# 安装必要的依赖
RUN apk add --no-cache ca-certificates tzdata

# 设置时区和服务环境变量
ENV TZ=Asia/Shanghai
ENV MAIN_SERVICE=BeeCount-Cloud
ENV START_COMMAND="--all"

# 设置工作目录
WORKDIR /app

# 复制构建产物
COPY --from=builder /app/bin /app/bin
# 保留web目录（包含静态资源）
COPY --from=builder /app/web /app/web
# 复制i18n目录（包含国际化文件）
COPY --from=builder /app/services/beecount/i18n /app/i18n

# 创建必要的目录
RUN mkdir -p /app/config /app/data /app/logs

# 暴露端口
EXPOSE 8080

# 设置启动命令
CMD ["/app/bin/BeeCount-Cloud", "--all"]
