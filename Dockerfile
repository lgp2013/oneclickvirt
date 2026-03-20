# OneClickVirt All-in-One Container

FROM node:22-slim AS frontend-builder
ARG TARGETARCH
WORKDIR /app/web

# Copy package files first
COPY web/package.json ./
COPY web/package-lock.json ./

# Remove any existing node_modules and reinstall from scratch
# This fixes the npm rollup optional dependency bug
RUN rm -rf node_modules package-lock.json && \
    npm install

# Install rollup for the correct architecture
RUN if [ "$TARGETARCH" = "amd64" ]; then \
        npm install @rollup/rollup-linux-x64-gnu; \
    elif [ "$TARGETARCH" = "arm64" ]; then \
        npm install @rollup/rollup-linux-arm64-gnu; \
    fi

# Copy the rest of the web files
COPY web/ ./

# Build the frontend
RUN npm run build


FROM golang:1.24-alpine AS backend-builder
ARG TARGETARCH
WORKDIR /app/server
RUN apk add --no-cache git ca-certificates
COPY server/ ./
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux GOARCH=${TARGETARCH} go build -a -installsuffix cgo -ldflags "-w -s" -o main .

FROM debian:12-slim
ARG TARGETARCH

# Install database and other services based on architecture
RUN apt-get update && \
    DEBIAN_FRONTEND=noninteractive apt-get install -y \
        gnupg2 wget lsb-release procps nginx supervisor ca-certificates && \
    if [ "$TARGETARCH" = "amd64" ]; then \
        echo "Installing MySQL for AMD64..." && \
        gpg --keyserver keyserver.ubuntu.com --recv-keys B7B3B788A8D3785C && \
        gpg --export B7B3B788A8D3785C > /usr/share/keyrings/mysql.gpg && \
        echo "deb [signed-by=/usr/share/keyrings/mysql.gpg] http://repo.mysql.com/apt/debian bookworm mysql-8.0" > /etc/apt/sources.list.d/mysql.list && \
        apt-get update && \
        DEBIAN_FRONTEND=noninteractive apt-get install -y mysql-server mysql-client; \
    else \
        echo "Installing MariaDB for ARM64..." && \
        DEBIAN_FRONTEND=noninteractive apt-get install -y mariadb-server mariadb-client; \
    fi && \
    apt-get clean

ENV TZ=Asia/Shanghai
WORKDIR /app
RUN mkdir -p /var/lib/mysql /var/log/mysql /var/run/mysqld /var/log/supervisor \
    && mkdir -p /app/storage/{cache,certs,configs,exports,logs,temp,uploads} \
    && mkdir -p /etc/mysql/conf.d

# Copy frontend build files
COPY --from=frontend-builder /app/web/dist /app/web/dist
COPY --from=frontend-builder /app/web/index.html /app/web/index.html
COPY --from=frontend-builder /app/web/logo.png /app/web/logo.png
COPY --from=frontend-builder /app/web/logo.ico /app/web/logo.ico
COPY --from=frontend-builder /app/web/package.json /app/web/package.json

# Copy backend binary and source
COPY --from=backend-builder /app/server/main /app/main
COPY --from=backend-builder /app/server/config.yaml /app/config.yaml
COPY --from=backend-builder /app/server/go.mod /app/go.mod
COPY --from=backend-builder /app/server/go.sum /app/go.sum
COPY --from=backend-builder /app/server/docs /app/docs
COPY --from=backend-builder /app/server/source /app/source

# Copy scripts and configs
COPY install.sh /app/install.sh
COPY deploy/my.cnf /etc/mysql/conf.d/my.cnf
COPY deploy/default.conf /etc/nginx/sites-available/default

# Copy router files
COPY server/router/dist /app/router/dist

# Create necessary symlinks and permissions
RUN ln -sf /dev/stdout /var/log/nginx/access.log && \
    ln -sf /dev/stderr /var/log/nginx/error.log && \
    chmod +x /app/install.sh

# Expose ports
EXPOSE 80 8888 3306 22

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=60s --retries=3 \
    CMD curl -f http://localhost:80/ || exit 1

# Start supervisor and services
CMD ["/bin/bash", "-c", "supervisord -n -c /etc/supervisor/supervisord.conf"]
