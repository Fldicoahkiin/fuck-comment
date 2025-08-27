# 多阶段构建 Dockerfile for fuck-comment

# 构建阶段
FROM golang:1.21-alpine AS builder

# 设置工作目录
WORKDIR /app

# 安装必要的包
RUN apk add --no-cache git ca-certificates tzdata

# 复制 go mod 文件
COPY go.mod go.sum ./

# 下载依赖
RUN go mod download

# 复制源代码
COPY . .

# 构建应用
ARG VERSION=docker
ARG BUILD_TIME
ARG GIT_COMMIT

RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags "-w -s -X main.Version=${VERSION} -X main.BuildTime=${BUILD_TIME} -X main.GitCommit=${GIT_COMMIT}" \
    -a -installsuffix cgo \
    -o fuck-comment .

# 运行阶段
FROM alpine:latest

# 安装ca-certificates用于HTTPS请求
RUN apk --no-cache add ca-certificates

# 创建非root用户
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

# 设置工作目录
WORKDIR /workspace

# 从构建阶段复制二进制文件
COPY --from=builder /app/fuck-comment /usr/local/bin/fuck-comment

# 设置文件权限
RUN chmod +x /usr/local/bin/fuck-comment

# 切换到非root用户
USER appuser

# 设置入口点
ENTRYPOINT ["fuck-comment"]

# 默认命令
CMD ["--help"]

# 元数据标签
LABEL maintainer="your-email@example.com" \
      description="一键删注释 - 跨平台代码注释删除工具" \
      version="${VERSION}" \
      org.opencontainers.image.title="fuck-comment" \
      org.opencontainers.image.description="一键删注释 - 跨平台代码注释删除工具" \
      org.opencontainers.image.vendor="Your Name" \
      org.opencontainers.image.version="${VERSION}" \
      org.opencontainers.image.created="${BUILD_TIME}" \
      org.opencontainers.image.revision="${GIT_COMMIT}" \
      org.opencontainers.image.licenses="MIT"
