package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"log"
	"math/big"
	"net"
	"net/http"
	"os"
	"rag-chatbot/handlers"
	"time"
)

func main() {
	mux := http.NewServeMux()

	// CORS middleware
	corsHandler := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
			
			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}
			
			next.ServeHTTP(w, r)
		})
	}

	// Routes
	mux.HandleFunc("/api/chat", handlers.ChatHandler)
	mux.HandleFunc("/api/chat/stream", handlers.ChatStreamHandler)
	mux.HandleFunc("/api/health", handlers.HealthHandler)
	
	// OAuth routes
	mux.HandleFunc("/api/auth/confluence", handlers.ConfluenceAuthHandler)
	mux.HandleFunc("/api/auth/confluence/callback", handlers.ConfluenceCallbackHandler)
	mux.HandleFunc("/api/auth/gmail", handlers.GmailAuthHandler)
	mux.HandleFunc("/api/auth/google/callback", handlers.GmailCallbackHandler)
	mux.HandleFunc("/api/auth/slack", handlers.SlackAuthHandler)
	mux.HandleFunc("/api/auth/slack/callback", handlers.SlackCallbackHandler)

	server := &http.Server{
		Addr:    ":8085",
		Handler: corsHandler(mux),
	}

	// Check if we should run with HTTPS (for Slack OAuth)
	if os.Getenv("USE_HTTPS") == "true" {
		log.Println("Server starting with HTTPS on :8085")
		
		// Generate self-signed certificate for development
		cert, err := generateSelfSignedCert()
		if err != nil {
			log.Fatal("Failed to generate certificate:", err)
		}
		
		server.TLSConfig = &tls.Config{
			Certificates: []tls.Certificate{cert},
		}
		
		log.Fatal(server.ListenAndServeTLS("", ""))
	} else {
		log.Println("Server starting on :8085")
		log.Fatal(server.ListenAndServe())
	}
}

func generateSelfSignedCert() (tls.Certificate, error) {
	// Generate a private key
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return tls.Certificate{}, err
	}

	// Create certificate template
	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization:  []string{"RAG Chatbot Dev"},
			Country:       []string{"US"},
			Province:      []string{""},
			Locality:      []string{""},
			StreetAddress: []string{""},
			PostalCode:    []string{""},
		},
		NotBefore:    time.Now(),
		NotAfter:     time.Now().Add(365 * 24 * time.Hour), // Valid for 1 year
		KeyUsage:     x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		IPAddresses:  []net.IP{net.IPv4(127, 0, 0, 1), net.IPv6loopback},
		DNSNames:     []string{"localhost"},
	}

	// Create the certificate
	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	if err != nil {
		return tls.Certificate{}, err
	}

	// Create tls.Certificate
	cert := tls.Certificate{
		Certificate: [][]byte{certDER},
		PrivateKey:  priv,
	}

	return cert, nil
}