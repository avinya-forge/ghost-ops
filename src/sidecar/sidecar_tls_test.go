package sidecar

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"io"
	"math/big"
	"net"
	"testing"
	"time"
)

func generateTestCert() (tls.Certificate, *x509.CertPool, error) {
	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return tls.Certificate{}, nil, err
	}

	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization: []string{"Acme Co"},
		},
		NotBefore: time.Now(),
		NotAfter:  time.Now().Add(time.Hour * 24 * 180),

		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
		BasicConstraintsValid: true,
	}
	template.IPAddresses = []net.IP{net.ParseIP("127.0.0.1")}

	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	if err != nil {
		return tls.Certificate{}, nil, err
	}

	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: derBytes})

	privBytes, err := x509.MarshalECPrivateKey(priv)
	if err != nil {
		return tls.Certificate{}, nil, err
	}
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: privBytes})

	cert, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		return tls.Certificate{}, nil, err
	}

	pool := x509.NewCertPool()
	pool.AppendCertsFromPEM(certPEM)

	return cert, pool, nil
}

func TestProxy_mTLS(t *testing.T) {
	cert, pool, err := generateTestCert()
	if err != nil {
		t.Fatalf("Failed to generate test cert: %v", err)
	}

	serverTLSConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
		ClientAuth:   tls.RequireAndVerifyClientCert,
		ClientCAs:    pool,
	}

	clientTLSConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
		RootCAs:      pool,
		ServerName:   "127.0.0.1",
	}

	// 1. Start a dummy target server (plain TCP)
	targetListener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Failed to start target server: %v", err)
	}
	defer targetListener.Close()

	targetAddr := targetListener.Addr().String()
	go func() {
		for {
			conn, err := targetListener.Accept()
			if err != nil {
				return
			}
			go func(c net.Conn) {
				defer c.Close()
				io.Copy(c, c) // Echo server
			}(conn)
		}
	}()

	// 2. Start server proxy (mTLS listener -> plain target)
	serverProxy := NewProxy(targetAddr)
	serverProxy.ServerTLSConfig = serverTLSConfig

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	serverProxyListener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Failed to listen for server proxy: %v", err)
	}
	serverProxyAddr := serverProxyListener.Addr().String()
	serverProxyListener.Close()

	errCh1 := make(chan error, 1)
	go func() {
		errCh1 <- serverProxy.Start(ctx, serverProxyAddr)
	}()

	// 3. Start client proxy (plain listener -> mTLS target(server proxy))
	clientProxy := NewProxy(serverProxyAddr)
	clientProxy.ClientTLSConfig = clientTLSConfig

	clientProxyListener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Failed to listen for client proxy: %v", err)
	}
	clientProxyAddr := clientProxyListener.Addr().String()
	clientProxyListener.Close()

	errCh2 := make(chan error, 1)
	go func() {
		errCh2 <- clientProxy.Start(ctx, clientProxyAddr)
	}()

	// Wait for proxies to start
	time.Sleep(200 * time.Millisecond)

	// 4. Test connection via client proxy
	conn, err := net.Dial("tcp", clientProxyAddr)
	if err != nil {
		t.Fatalf("Failed to connect to client proxy: %v", err)
	}
	defer conn.Close()

	msg := "hello mTLS proxy"
	fmt.Fprintf(conn, "%s", msg)

	buf := make([]byte, len(msg))
	_, err = io.ReadFull(conn, buf)
	if err != nil {
		t.Fatalf("Failed to read from client proxy: %v", err)
	}

	if string(buf) != msg {
		t.Errorf("Expected %q, got %q", msg, string(buf))
	}

	clientProxy.Stop()
	serverProxy.Stop()
	cancel()
	<-errCh1
	<-errCh2
}
