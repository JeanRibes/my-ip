#!/bin/bash
# Script to generate a self-signed certificate for local development

set -e

echo "Generating self-signed certificate for local development..."

# Generate a self-signed certificate valid for 365 days
# -x509: generate a self-signed certificate
# -newkey rsa:4096: generate a 4096-bit RSA key
# -keyout key.pem: output the private key
# -out cert.pem: output the certificate
# -days 365: valid for 365 days
# -nodes: do not encrypt the private key (no passphrase)
# -subj: set the subject of the certificate

openssl req -x509 -newkey rsa:4096 \
    -keyout key.pem \
    -out cert.pem \
    -days 365 \
    -nodes \
    -subj "/CN=localhost"

echo "Certificate generated successfully!"
echo "  Certificate: cert.pem"
echo "  Private Key: key.pem"
echo ""
echo "To start the server with HTTPS:"
echo "  ./main.go --tls --cert-file cert.pem --key-file key.pem"
