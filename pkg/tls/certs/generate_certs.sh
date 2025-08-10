#!/bin/bash

# 清理旧文件
rm -f *.pem *.key *.crt *.csr *.srl openssl.cnf

# 创建 openssl 配置文件
cat > openssl.cnf <<EOF
[req]
distinguished_name = req_distinguished_name
req_extensions = v3_req

[req_distinguished_name]
# 留空以在命令行中提供

[v3_req]
subjectAltName = @alt_names

[alt_names]
DNS.1 = localhost
IP.1 = 127.0.0.1

[v3_ca]
subjectKeyIdentifier = hash
authorityKeyIdentifier = keyid:always,issuer
basicConstraints = critical, CA:TRUE
EOF

# 1. 创建CA私钥和证书
openssl genrsa -out ca.key 2048
openssl req -new -x509 -days 365 -key ca.key -out ca.crt -subj "/CN=MyCA" -extensions v3_ca -config openssl.cnf

# 2. 生成服务器证书
openssl genrsa -out server.key 2048
openssl req -new -key server.key -out server.csr -subj "/CN=server" -config openssl.cnf
openssl x509 -req -days 365 -in server.csr -CA ca.crt -CAkey ca.key -CAcreateserial -out server.crt -extfile openssl.cnf -extensions v3_req

# 3. 生成客户端证书 (客户端证书通常不需要SAN)
openssl genrsa -out client.key 2048
openssl req -new -key client.key -out client.csr -subj "/CN=client"
openssl x509 -req -days 365 -in client.csr -CA ca.crt -CAkey ca.key -CAcreateserial -out client.crt

# 4. 生成网关证书
openssl genrsa -out gateway.key 2048
openssl req -new -key gateway.key -out gateway.csr -subj "/CN=gateway" -config openssl.cnf
openssl x509 -req -days 365 -in gateway.csr -CA ca.crt -CAkey ca.key -CAcreateserial -out gateway.crt -extfile openssl.cnf -extensions v3_req

# 设置适当的权限
chmod 600 *.key
chmod 644 *.crt *.csr *.srl *.cnf

echo "证书生成完毕!" 