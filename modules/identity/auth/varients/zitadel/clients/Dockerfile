# =============================================================================
# ZITADEL ALL-IN-ONE - For HuggingFace Spaces (Development/Testing Only)
# Combines: PostgreSQL + Zitadel + Login UI in a single container
# =============================================================================

FROM ubuntu:22.04

ENV DEBIAN_FRONTEND=noninteractive

# -----------------------------------------------------------------------------
# Install dependencies
# -----------------------------------------------------------------------------
RUN apt-get update && apt-get install -y \
    curl \
    wget \
    gnupg \
    lsb-release \
    supervisor \
    postgresql \
    postgresql-contrib \
    nodejs \
    npm \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/*

# -----------------------------------------------------------------------------
# Install Zitadel
# -----------------------------------------------------------------------------
ARG ZITADEL_VERSION=2.67.2
RUN mkdir -p /app/zitadel \
    && curl -L "https://github.com/zitadel/zitadel/releases/download/v${ZITADEL_VERSION}/zitadel-linux-amd64.tar.gz" \
    | tar -xz -C /app/zitadel --strip-components=1

# -----------------------------------------------------------------------------
# Download Login UI
# -----------------------------------------------------------------------------
ARG LOGIN_VERSION=latest
RUN mkdir -p /app/login \
    && curl -L "https://github.com/zitadel/zitadel-login/releases/latest/download/zitadel-login-linux-amd64.tar.gz" \
    | tar -xz -C /app/login --strip-components=1 2>/dev/null || echo "Login binary not available, will use container"

# -----------------------------------------------------------------------------
# PostgreSQL Configuration
# -----------------------------------------------------------------------------
USER postgres
RUN /etc/init.d/postgresql start && \
    psql --command "ALTER USER postgres PASSWORD 'postgres';" && \
    createdb zitadel && \
    psql --command "CREATE USER zitadel WITH PASSWORD 'zitadel';" && \
    psql --command "GRANT ALL PRIVILEGES ON DATABASE zitadel TO zitadel;" && \
    /etc/init.d/postgresql stop

USER root

# -----------------------------------------------------------------------------
# Zitadel Configuration
# -----------------------------------------------------------------------------
RUN mkdir -p /data/zitadel
COPY <<EOF /app/zitadel/config.yaml
Log:
  Level: info

ExternalDomain: localhost
ExternalPort: 7860
ExternalSecure: false

TLS:
  Enabled: false

Database:
  postgres:
    Host: localhost
    Port: 5432
    Database: zitadel
    User:
      Username: zitadel
      Password: zitadel
      SSL:
        Mode: disable
    Admin:
      Username: postgres
      Password: postgres
      SSL:
        Mode: disable

FirstInstance:
  Org:
    Human:
      UserName: admin
      Password: Admin123!
      PasswordChangeRequired: false
EOF

# -----------------------------------------------------------------------------
# Supervisord Configuration
# -----------------------------------------------------------------------------
COPY <<EOF /etc/supervisor/conf.d/services.conf
[supervisord]
nodaemon=true
logfile=/var/log/supervisor/supervisord.log
pidfile=/var/run/supervisord.pid

[program:postgresql]
command=/usr/lib/postgresql/14/bin/postgres -D /var/lib/postgresql/14/main -c config_file=/etc/postgresql/14/main/postgresql.conf
user=postgres
autostart=true
autorestart=true
priority=10
stdout_logfile=/var/log/supervisor/postgresql.log
stderr_logfile=/var/log/supervisor/postgresql_err.log

[program:zitadel]
command=/app/zitadel/zitadel start-from-init --masterkeyFile /data/zitadel/masterkey --config /app/zitadel/config.yaml
directory=/app/zitadel
autostart=true
autorestart=true
priority=20
startsecs=10
startretries=5
stdout_logfile=/var/log/supervisor/zitadel.log
stderr_logfile=/var/log/supervisor/zitadel_err.log
EOF

# -----------------------------------------------------------------------------
# Startup Script
# -----------------------------------------------------------------------------
COPY <<EOF /app/start.sh
#!/bin/bash
set -e

# Generate masterkey if not exists
if [ ! -f /data/zitadel/masterkey ]; then
    echo "MasterkeyNeedsToHave32Characters" > /data/zitadel/masterkey
fi

# Fix PostgreSQL permissions
chown -R postgres:postgres /var/lib/postgresql
chown -R postgres:postgres /var/run/postgresql || true

# Create log directory
mkdir -p /var/log/supervisor

# Start supervisord
exec /usr/bin/supervisord -c /etc/supervisor/supervisord.conf
EOF

RUN chmod +x /app/start.sh

# -----------------------------------------------------------------------------
# Expose & Run
# -----------------------------------------------------------------------------
EXPOSE 7860

VOLUME ["/data/zitadel", "/var/lib/postgresql"]

CMD ["/app/start.sh"]
