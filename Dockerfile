# ============================================================
# 数据中台一体镜像：单镜像打包 12 个二进制（11 个服务 + 调度引擎）+ 容器化配置
#   网关:     datacenter-api（对外 8887）
#   RPC 服务: datasource/metadata/datasvc/catalog/quality/metric/tag/job/ide/ai
#   调度引擎: scheduler（内置轻量调度，无 yaml 配置，环境变量见部署手册 §8）
# 注意：本机无 docker 构建型 CI，首次使用 `docker compose build` 验证。
# 用法：docker compose build && docker compose up -d
# ============================================================

# ---------- 阶段一：builder，编译全部 12 个二进制 ----------
FROM golang:1.26-alpine AS builder

# git 用于 go mod 拉取依赖；ca-certificates 用于 https 模块下载
RUN apk add --no-cache git ca-certificates

WORKDIR /src

# 先拷贝依赖清单并下载，最大化利用 docker 层缓存（源码改动不重复拉依赖）
COPY go.mod go.sum ./
RUN go mod download

# 再拷贝全部源码
COPY . .

# CGO_ENABLED=0 纯静态编译，运行层无需 glibc
ENV CGO_ENABLED=0 GOOS=linux
RUN mkdir -p /app/bin \
 && go build -o /app/bin/datacenter-api    . \
 && go build -o /app/bin/datasource-server ./datasource/rpc \
 && go build -o /app/bin/metadata-server   ./metadata/rpc \
 && go build -o /app/bin/datasvc-server    ./datasvc/rpc \
 && go build -o /app/bin/catalog-server    ./catalog/rpc \
 && go build -o /app/bin/quality-server    ./quality/rpc \
 && go build -o /app/bin/metric-server     ./metric/rpc \
 && go build -o /app/bin/tag-server        ./tag/rpc \
 && go build -o /app/bin/job-server        ./job/rpc \
 && go build -o /app/bin/ide-server        ./ide/rpc \
 && go build -o /app/bin/ai-server         ./ai/rpc \
 && go build -o /app/bin/scheduler         ./cmd/scheduler

# ---------- 阶段二：运行层，轻量 alpine ----------
FROM alpine:3.20

# ca-certificates：访问外部 https（如 LLM 接口）
# tzdata + Asia/Shanghai：容器时区与宿主一致
RUN apk add --no-cache ca-certificates tzdata \
 && cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime \
 && echo "Asia/Shanghai" > /etc/timezone

# 拷贝 12 个二进制（11 个服务 + 调度引擎 scheduler）
COPY --from=builder /app/bin /app/bin
# 拷贝容器化配置（主机名均为 compose 服务名：mysql/redis/etcd）
COPY deploy/docker /app/etc
# 网关静态资源目录（/static/** 路由使用）
COPY assets /app/assets

# 工作目录设为 /app/bin，compose 中各服务以不同 command 启动：
#   ./datasource-server -f /app/etc/datasource.yaml
#   ./datacenter-api    -f /app/etc/datacenter-api.docker.yaml
#   ./scheduler         （调度引擎，无 -f 配置，读 SCHED_MYSQL_DSN/SCHED_ETCD_HOSTS）
WORKDIR /app/bin
