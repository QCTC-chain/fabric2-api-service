/*
Copyright IBM Corp. 2016 All Rights Reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

                 http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package tlsgm

import (
	// "crypto/tls"

	"io/ioutil"
	"time"

	"github.com/pkg/errors"

	"github.com/cloudflare/cfssl/log"

	// "github.com/hyperledger/fabric/bccsp/factory"
	// "github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/bccsp/factory"
	factory "github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric-ca/sdkpatch/cryptosuitebridge"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/core"

	// "github.com/tjfoc/fabric-ca-gm/util"
	// "github.com/hyperledger/fabric-ca/util"
	// "github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric-ca/util"
	gmtls "gitee.com/china_uni/tjfoc-gm/tls"
	tls "gitee.com/china_uni/tjfoc-gm/tls"
	gmx509 "gitee.com/china_uni/tjfoc-gm/x509"
	"github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric-ca/sdkinternal/pkg/util"
)

// DefaultCipherSuites is a set of strong TLS cipher suites
var DefaultCipherSuites = []uint16{
	tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
	tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
	tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
	tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
	tls.TLS_RSA_WITH_AES_128_GCM_SHA256,
	tls.TLS_RSA_WITH_AES_256_GCM_SHA384,
	//add by liuhy
	tls.GMTLS_SM2_WITH_SM4_SM3,
	tls.GMTLS_ECDHE_SM2_WITH_SM4_SM3,
}

// ServerTLSConfig defines key material for a TLS server
type ServerTLSConfig struct {
	Enabled    bool   `help:"Enable TLS on the listening port"`
	CertFile   string `def:"tls-cert.pem" help:"PEM-encoded TLS certificate file for server's listening port"`
	KeyFile    string `help:"PEM-encoded TLS key for server's listening port"`
	ClientAuth ClientAuth
}

// ClientAuth defines the key material needed to verify client certificates
type ClientAuth struct {
	Type      string   `def:"noclientcert" help:"Policy the server will follow for TLS Client Authentication."`
	CertFiles []string `help:"A list of comma-separated PEM-encoded trusted certificate files (e.g. root1.pem,root2.pem)"`
}

// ClientTLSConfig defines the key material for a TLS client
type ClientTLSConfig struct {
	Enabled     bool     `skip:"true"`
	CertFiles   [][]byte `help:"A list of comma-separated PEM-encoded trusted certificate files (e.g. root1.pem,root2.pem)"`
	Client      KeyCertFiles
	TlsCertPool *gmx509.CertPool
	IsGMTLS     bool //add by liuhy for gm tls
}

// KeyCertFiles defines the files need for client on TLS
type KeyCertFiles struct {
	KeyFile  []byte `help:"PEM-encoded key file when mutual authentication is enabled"`
	CertFile []byte `help:"PEM-encoded certificate file when mutual authenticate is enabled"`
}

// GetClientTLSConfig creates a tls.Config object from certs and roots
func GetClientTLSConfig(cfg *ClientTLSConfig, csp core.CryptoSuite) (*gmtls.Config, error) {
	var (
		certs  []gmtls.Certificate
		config *gmtls.Config
	)

	if csp == nil {
		csp = factory.GetDefault()
	}

	log.Debugf("CA Files: %+v\n", cfg.CertFiles)
	log.Debugf("Client Cert File: %s\n", cfg.Client.CertFile)
	log.Debugf("Client Key File: %s\n", cfg.Client.KeyFile)

	if cfg.Client.CertFile != nil {
		err := checkCertDates("", cfg.Client.CertFile)
		if err != nil {
			return nil, err
		}

		clientCert, err := util.LoadX509KeyPairGM("", "", cfg.Client.CertFile, cfg.Client.KeyFile, csp)
		if err != nil {
			return nil, err
		}

		certs = append(certs, *clientCert)
	} else {
		log.Debug("Client TLS certificate and/or key file not provided")
	}
	rootCAPool := gmx509.NewCertPool()
	if len(cfg.CertFiles) == 0 {
		return nil, errors.New("No TLS certificate files were provided")
	}

	for _, cacert := range cfg.CertFiles {
		//delete by liuhy
		// caCert, err := ioutil.ReadFile(cacert)
		// if err != nil {
		// 	return nil, errors.Wrapf(err, "Failed to read '%s'", cacert)
		// }
		ok := rootCAPool.AppendCertsFromPEM(cacert)
		if !ok {
			return nil, errors.Errorf("Failed to process certificate from file %s", cacert)
		}
	}

	//add by liu for tls gm
	if cfg.IsGMTLS {
		config = &gmtls.Config{
			GMSupport:    &gmtls.GMSupport{},
			Certificates: certs,
			RootCAs:      rootCAPool,
		}
	} else {
		config = &gmtls.Config{
			Certificates: certs,
			RootCAs:      rootCAPool,
		}
	}
	return config, nil
}

// AbsTLSClient makes TLS client files absolute
func AbsTLSClient(cfg *ClientTLSConfig, configDir string) error {
	var err error

	//delete by liuhy
	// for i := 0; i < len(cfg.CertFiles); i++ {
	// 	cfg.CertFiles[i], err = util.MakeFileAbs(cfg.CertFiles[i], configDir)
	// 	if err != nil {
	// 		return err
	// 	}
	// 	//cfg.CertFiles[i] = []byte(certBytes)
	// }

	certFile, err := util.MakeFileAbs(string(cfg.Client.CertFile), configDir)
	if err != nil {
		return err
	}
	cfg.Client.CertFile = []byte(certFile)

	keyFile, err := util.MakeFileAbs(string(cfg.Client.KeyFile), configDir)
	if err != nil {
		return err
	}
	cfg.Client.KeyFile = []byte(keyFile)

	return nil
}

// AbsTLSServer makes TLS client files absolute
func AbsTLSServer(cfg *ServerTLSConfig, configDir string) error {
	var err error

	for i := 0; i < len(cfg.ClientAuth.CertFiles); i++ {
		cfg.ClientAuth.CertFiles[i], err = util.MakeFileAbs(cfg.ClientAuth.CertFiles[i], configDir)
		if err != nil {
			return err
		}

	}

	cfg.CertFile, err = util.MakeFileAbs(cfg.CertFile, configDir)
	if err != nil {
		return err
	}

	cfg.KeyFile, err = util.MakeFileAbs(cfg.KeyFile, configDir)
	if err != nil {
		return err
	}

	return nil
}

func checkCertDates(certFile string, x509Cert []byte) error {
	log.Debug("Check client TLS certificate for valid dates")
	var (
		certPEM []byte
		err     error
	)

	if x509Cert == nil {
		certPEM, err = ioutil.ReadFile(certFile)
		if err != nil {
			return errors.Wrapf(err, "Failed to read file '%s'", certFile)
		}
	} else {
		certPEM = x509Cert
	}

	cert, err := util.GetX509CertificateFromPEM(certPEM)
	if err != nil {
		return err
	}

	notAfter := cert.NotAfter
	currentTime := time.Now().UTC()

	if currentTime.After(notAfter) {
		return errors.New("Certificate provided has expired")
	}

	notBefore := cert.NotBefore
	if currentTime.Before(notBefore) {
		return errors.New("Certificate provided not valid until later date")
	}

	return nil
}
