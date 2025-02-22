#!/bin/bash
set -e

# Если MYSQL_HOST не передан, определяем его автоматически
if [ -z "$MYSQL_HOST" ]; then
  echo "Определяем IP MySQL..."

  if getent hosts host.docker.internal >/dev/null 2>&1; then
    MYSQL_HOST="host.docker.internal"
#   elif grep -qEi "(microsoft|wsl)" /proc/version &> /dev/null; then
#     MYSQL_HOST=$(ip route | awk '/default/ {print $3}')
  elif ip route | grep -q "192.168."; then
    MYSQL_HOST=$(ip route | awk '/192.168./ {print $3; exit}')
  elif ip route | grep -q "10.8."; then
    MYSQL_HOST=$(ip route | awk '/10.8./ {print $3; exit}')
  else
    MYSQL_HOST="127.0.0.1"
  fi

  echo "MySQL найден на $MYSQL_HOST"
  export MYSQL_HOST
fi

echo "Запускаю команду: $@"
exec "$@"