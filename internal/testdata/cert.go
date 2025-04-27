package testdata

import (
	"crypto/tls"
<<<<<<< HEAD
	"crypto/x509"
	"os"
=======
>>>>>>> project-faster/main
	"path"
	"runtime"
)

var certPath string

func init() {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		panic("Failed to get current frame")
	}

<<<<<<< HEAD
	certPath = path.Dir(filename)
}

// GetCertificatePaths returns the paths to certificate and key
func GetCertificatePaths() (string, string) {
	return path.Join(certPath, "cert.pem"), path.Join(certPath, "priv.key")
=======
	certPath = path.Join(path.Dir(path.Dir(path.Dir(filename))), "example")
}

// GetCertificatePaths returns the paths to 'fullchain.pem' and 'privkey.pem' for the
// quic.clemente.io cert.
func GetCertificatePaths() (string, string) {
	return path.Join(certPath, "fullchain.pem"), path.Join(certPath, "privkey.pem")
>>>>>>> project-faster/main
}

// GetTLSConfig returns a tls config for quic.clemente.io
func GetTLSConfig() *tls.Config {
	cert, err := tls.LoadX509KeyPair(GetCertificatePaths())
	if err != nil {
		panic(err)
	}
	return &tls.Config{
<<<<<<< HEAD
		MinVersion:   tls.VersionTLS13,
=======
>>>>>>> project-faster/main
		Certificates: []tls.Certificate{cert},
	}
}

<<<<<<< HEAD
// AddRootCA adds the root CA certificate to a cert pool
func AddRootCA(certPool *x509.CertPool) {
	caCertPath := path.Join(certPath, "ca.pem")
	caCertRaw, err := os.ReadFile(caCertPath)
	if err != nil {
		panic(err)
	}
	if ok := certPool.AppendCertsFromPEM(caCertRaw); !ok {
		panic("Could not add root ceritificate to pool.")
	}
}

// GetRootCA returns an x509.CertPool containing (only) the CA certificate
func GetRootCA() *x509.CertPool {
	pool := x509.NewCertPool()
	AddRootCA(pool)
	return pool
=======
// GetCertificate returns a certificate for quic.clemente.io
func GetCertificate() tls.Certificate {
	cert, err := tls.LoadX509KeyPair(GetCertificatePaths())
	if err != nil {
		panic(err)
	}
	return cert
>>>>>>> project-faster/main
}
