# =============================================================================
# ZITADEL ALL-IN-ONE - For HuggingFace Spaces (Development/Testing Only)
# Combines: PostgreSQL + Zitadel + Login UI in a single container
#
# Reference: docker-compose.yaml (official Zitadel setup)
# Constraint: HF Spaces only exposes port 7860, no docker-compose support
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
  git \
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
# Zitadel Configuration (matching docker-compose environment variables)
# -----------------------------------------------------------------------------
RUN mkdir -p /data/zitadel /current-dir /var/log/supervisor

# Write config via RUN to guarantee LF line endings
RUN printf '%s\n' \
  'Log:' \
  '  Level: info' \
  '' \
  'Port: 7860' \
  'ExternalPort: 7860' \
  'ExternalSecure: false' \
  '' \
  'TLS:' \
  '  Enabled: false' \
  '' \
  'Database:' \
  '  postgres:' \
  '    Host: localhost' \
  '    Port: 5432' \
  '    Database: zitadel' \
  '    User:' \
  '      Username: zitadel' \
  '      Password: zitadel' \
  '      SSL:' \
  '        Mode: disable' \
  '    Admin:' \
  '      Username: postgres' \
  '      Password: postgres' \
  '      SSL:' \
  '        Mode: disable' \
  '' \
  'FirstInstance:' \
  '  Org:' \
  '    Human:' \
  '      UserName: admin' \
  '      Password: Admin123!' \
  '      PasswordChangeRequired: false' \
  '      Email:' \
  '        Address: admin@nucleus.dev' \
  '        Verified: true' \
  > /app/zitadel/config.yaml

# -----------------------------------------------------------------------------
# Supervisord Configuration (written via RUN for guaranteed LF endings)
# -----------------------------------------------------------------------------
RUN printf '%s\n' \
  '[supervisord]' \
  'nodaemon=true' \
  'logfile=/var/log/supervisor/supervisord.log' \
  'pidfile=/var/run/supervisord.pid' \
  '' \
  '[program:postgresql]' \
  'command=/usr/lib/postgresql/14/bin/postgres -D /var/lib/postgresql/14/main -c config_file=/etc/postgresql/14/main/postgresql.conf' \
  'user=postgres' \
  'autostart=true' \
  'autorestart=true' \
  'priority=10' \
  'stdout_logfile=/dev/stdout' \
  'stdout_logfile_maxbytes=0' \
  'stderr_logfile=/dev/stderr' \
  'stderr_logfile_maxbytes=0' \
  '' \
  '[program:zitadel]' \
  'command=/app/zitadel-start.sh' \
  'directory=/app/zitadel' \
  'autostart=true' \
  'autorestart=true' \
  'priority=20' \
  'startsecs=30' \
  'startretries=3' \
  'stdout_logfile=/dev/stdout' \
  'stdout_logfile_maxbytes=0' \
  'stderr_logfile=/dev/stderr' \
  'stderr_logfile_maxbytes=0' \
  > /etc/supervisor/conf.d/services.conf

# -----------------------------------------------------------------------------
# Zitadel Startup Wrapper - waits for PostgreSQL (written via RUN for LF)
# Uses --masterkey inline (matching docker-compose approach)
# -----------------------------------------------------------------------------
RUN printf '#!/bin/bash\n\
  set -e\n\
  echo "Waiting for PostgreSQL to accept connections..."\n\
  for i in $(seq 1 30); do\n\
  if pg_isready -h localhost -U postgres -q 2>/dev/null; then\n\
  echo "PostgreSQL is ready!"\n\
  break\n\
  fi\n\
  echo "  Attempt $i/30 - PostgreSQL not ready yet..."\n\
  sleep 2\n\
  done\n\
  if ! pg_isready -h localhost -U postgres -q 2>/dev/null; then\n\
  echo "ERROR: PostgreSQL failed to start after 60 seconds"\n\
  exit 1\n\
  fi\n\
  echo "Starting Zitadel (start-from-init)..."\n\
  exec /app/zitadel/zitadel start-from-init \\\n\
  --masterkey "MasterkeyNeedsToHave32Characters" \\\n\
  --config /app/zitadel/config.yaml\n' > /app/zitadel-start.sh \
  && chmod +x /app/zitadel-start.sh

# -----------------------------------------------------------------------------
# Main Startup Script (written via RUN for guaranteed LF)
# -----------------------------------------------------------------------------
RUN printf '#!/bin/bash\n\
  set -e\n\
  # Fix PostgreSQL permissions\n\
  chown -R postgres:postgres /var/lib/postgresql\n\
  chown -R postgres:postgres /var/run/postgresql || true\n\
  mkdir -p /var/log/supervisor\n\
  mkdir -p /current-dir\n\
  echo "Starting services via supervisord..."\n\
  exec /usr/bin/supervisord -c /etc/supervisor/supervisord.conf\n' > /app/start.sh \
  && chmod +x /app/start.sh

# -----------------------------------------------------------------------------
# Expose & Run
# -----------------------------------------------------------------------------
EXPOSE 7860

CMD ["/app/start.sh"]
