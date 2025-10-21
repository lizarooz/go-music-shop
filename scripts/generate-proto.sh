#!/bin/bash

# Создаем папку для сгенерированного кода
mkdir -p pkg/gen/catalog

# Генерируем Go код из proto файлов
protoc --go_out=./pkg/gen/catalog \
       --go-grpc_out=./pkg/gen/catalog \
       --go_opt=paths=source_relative \
       --go-grpc_opt=paths=source_relative \
       --proto_path=./cmd/api/proto \
       ./cmd/api/proto/catalog.proto

echo "✅ Go код из Protobuf сгенерирован успешно!"