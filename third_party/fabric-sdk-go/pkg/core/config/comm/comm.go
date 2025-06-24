/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package comm

import (
	// "crypto/tls"
	// "crypto/x509"
	"gitee.com/china_uni/tjfoc-gm/tls"
	"gitee.com/china_uni/tjfoc-gm/x509"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/core"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/cryptosuite"
	"github.com/pkg/errors"
)

// TLSConfig returns the appropriate config for TLS including the root CAs,
// certs for mutual TLS, and server host override. Works with certs loaded either from a path or embedded pem.
// modify by liuhy for gm tls
func TLSConfig(cert *x509.Certificate, serverName string, config fab.EndpointConfig, isGMTLS bool) (*tls.Config, error) {

	if cert != nil {
		config.TLSCACertPool().Add(cert)
	}

	certPool, err := config.TLSCACertPool().Get()
	if err != nil {
		return nil, err
	}

	//add by liuhy for gm tls
	if isGMTLS {
		return &tls.Config{GMSupport: &tls.GMSupport{}, RootCAs: certPool, Certificates: config.TLSClientCerts(), ServerName: serverName}, nil
	}
	return &tls.Config{RootCAs: certPool, Certificates: config.TLSClientCerts(), ServerName: serverName}, nil
}

// TLSCertHash is a utility method to calculate the SHA256 hash of the configured certificate (for usage in channel headers)
func TLSCertHash(config fab.EndpointConfig, isSM3 bool) ([]byte, error) {
	certs := config.TLSClientCerts()
	if len(certs) == 0 {
		return computeHash([]byte(""), isSM3)
	}

	cert := certs[0]
	if len(cert.Certificate) == 0 {
		return computeHash([]byte(""), isSM3)
	}

	return computeHash(cert.Certificate[0], isSM3)
}

// computeHash computes hash for given bytes using underlying cryptosuite default
func computeHash(msg []byte, isSM3 bool) ([]byte, error) {
	var opts core.HashOpts
	if isSM3 {
		opts = cryptosuite.GetSM3Opts()
	} else {
		opts = cryptosuite.GetSHA256Opts()
	}
	h, err := cryptosuite.GetDefault().Hash(msg, opts)
	if err != nil {
		return nil, errors.WithMessage(err, "failed to compute tls cert hash")
	}
	return h, err
}
