package certificates

import (
	"crypto/x509"
	"gitlab.lrz.de/orderless/orderlesschain/internal/logger"
	"io/ioutil"
	"path/filepath"
)

const certsDir = "./configs/certs"

var CAs *x509.CertPool

func init() {
	tempCAs := x509.NewCertPool()
	certs, err := ioutil.ReadDir(certsDir)
	if err != nil {
		logger.FatalLogger.Fatalln("Could not find certificates")
	}
	for _, cert := range certs {
		b, _ := ioutil.ReadFile(filepath.Join(certsDir, cert.Name()))
		if !tempCAs.AppendCertsFromPEM(b) {
			logger.FatalLogger.Fatalln("Failed to import the certificate")
		}
	}
	CAs = tempCAs
}
