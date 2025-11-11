package main

import (
	"crypto/tls"
	"crypto/x509"
	"flowedge-client/service"
	"io/ioutil"
	"log"
)

func main() {
	// Load client certificate and key
	clientCert, err := tls.LoadX509KeyPair("./certs/client.crt", "./certs/client.key")
	if err != nil {
		log.Fatalf("failed to load client certificates: %v", err)
	}
	// Load CA certificate
	caCert, err := ioutil.ReadFile("./certs/ca.crt")
	if err != nil {
		log.Fatalf("failed to read CA certificate: %v", err)
	}
	certPool := x509.NewCertPool()
	if ok := certPool.AppendCertsFromPEM(caCert); !ok {
		log.Fatalf("failed to append CA certificates to pool")
	}
	// Set up the TLS configuration
	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{clientCert},
		RootCAs:      certPool,
	}
	//service.StartAgent("192.168.1.112:50051", tlsConfig)
	//service.StartAgent("10.11.11.56:50051", tlsConfig)
	service.StartAgent("222.73.137.153:50051", tlsConfig)
}
