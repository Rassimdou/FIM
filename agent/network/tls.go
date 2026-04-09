package network

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"

	"google.golang.org/grpc/credentials"
)

func LoadClientTLSConfig(caCertPath, clientCertPath, clientKeyPath string) (credentials.TransportCredentials, error) {
	// Load CA certificate
	cert, err := tls.LoadX509KeyPair(clientCertPath, clientKeyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load client certificate and key: %v", err)
	}

	//read CA cert
	caCert, err := os.ReadFile(caCertPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read CA certificate: %v", err)
	}

	// Create a certificate pool and add the CA certificate
	certPool := x509.NewCertPool()
	if !certPool.AppendCertsFromPEM(caCert) {
		return nil, fmt.Errorf("failed to append CA certificate to pool")
	}

	// Create TLS configuration
	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
		RootCAs:      certPool,
	}

	return credentials.NewTLS(tlsConfig), nil
}
