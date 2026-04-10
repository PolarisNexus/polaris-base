#!/bin/bash
# 为需要独立数据库的组件创建数据库（PG 容器启动时自动执行）
set -e

psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" <<-EOSQL
    SELECT 'CREATE DATABASE casdoor OWNER $POSTGRES_USER'
    WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'casdoor')\gexec
EOSQL
