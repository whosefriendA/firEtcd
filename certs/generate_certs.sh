#!/bin/bash

# 创建CA私钥和证书
openssl genrsa -out ca.key 2048
openssl req -new -x509 -days 365 -key ca.key -out ca.crt -subj "/CN=CA"

# 生成服务器证书
openssl genrsa -out server.key 2048
openssl req -new -key server.key -out server.csr -subj "/CN=server"
openssl x509 -req -days 365 -in server.csr -CA ca.crt -CAkey ca.key -CAcreateserial -out server.crt

# 生成客户端证书
openssl genrsa -out client.key 2048
openssl req -new -key client.key -out client.csr -subj "/CN=client"
openssl x509 -req -days 365 -in client.csr -CA ca.crt -CAkey ca.key -CAcreateserial -out client.crt

# 生成网关证书
openssl genrsa -out gateway.key 2048
openssl req -new -key gateway.key -out gateway.csr -subj "/CN=gateway"
openssl x509 -req -days 365 -in gateway.csr -CA ca.crt -CAkey ca.key -CAcreateserial -out gateway.crt

# 设置适当的权限
chmod 600 *.key
chmod 644 *.crt
chmod 644 *.csr
chmod 644 *.srl 