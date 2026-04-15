#!/bin/bash
# 根据 POLARIS_EXTRA_DBS 环境变量动态创建额外数据库
# 由 PostgreSQL 镜像 /docker-entrypoint-initdb.d 机制在首次启动时执行
set -e
for db in ${POLARIS_EXTRA_DBS}; do
  psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" <<-EOSQL
    SELECT 'CREATE DATABASE "$db" OWNER $POSTGRES_USER'
    WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = '$db')\gexec
EOSQL
done
