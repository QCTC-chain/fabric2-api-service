/*
Copyright IBM Corp. 2017 All Rights Reserved.

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
/*
Notice: This file has been modified for Hyperledger Fabric SDK Go usage.
Please review third_party pinning scripts and patches for more details.
*/

package util

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"strings"

	"gitee.com/china_uni/tjfoc-gm/sm2"
	gmtls "gitee.com/china_uni/tjfoc-gm/tls"
	gmx509 "gitee.com/china_uni/tjfoc-gm/x509"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/core"

	"github.com/cloudflare/cfssl/csr"
	"github.com/cloudflare/cfssl/helpers"
	factory "github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric-ca/sdkpatch/cryptosuitebridge"
	log "github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric-ca/sdkpatch/logbridge"
	"github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/bccsp"
	cspsigner "github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/bccsp/signer"
	"github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/bccsp/sw"
	"github.com/pkg/errors"
)

// getBCCSPKeyOpts generates a key as specified in the request.
// This supports ECDSA.
func getBCCSPKeyOpts(kr *csr.KeyRequest, ephemeral bool) (opts core.KeyGenOpts, err error) {
	if kr == nil {
		return factory.GetECDSAKeyGenOpts(ephemeral), nil
	}
	log.Debugf("generate key from request: algo=%s, size=%d", kr.Algo(), kr.Size())
	switch kr.Algo() {
	case "ecdsa":
		switch kr.Size() {
		case 256:
			return factory.GetECDSAP256KeyGenOpts(ephemeral), nil
		case 384:
			return factory.GetECDSAP384KeyGenOpts(ephemeral), nil
		case 521:
			// Need to add curve P521 to bccsp
			// return &bccsp.ECDSAP512KeyGenOpts{Temporary: false}, nil
			return nil, errors.New("Unsupported ECDSA key size: 521")
		default:
			return nil, errors.Errorf("Invalid ECDSA key size: %d", kr.Size())
		}
	case "sm2":
		return &bccsp.SM2KeyGenOpts{Temporary: ephemeral}, nil
	default:
		return nil, errors.Errorf("Invalid algorithm: %s", kr.Algo())
	}
}

// GetSignerFromCert load private key represented by ski and return bccsp signer that conforms to crypto.Signer
func GetSignerFromGMCert(cert *gmx509.Certificate, csp core.CryptoSuite) (core.Key, crypto.Signer, error) {
	if csp == nil {
		return nil, nil, errors.New("CSP was not initialized")
	}

	//add by liuhy
	var (
		certPubK core.Key
		err      error
	)
	switch cert.PublicKey.(type) {
	case *sm2.PublicKey:
		log.Infof("gm cert is sm2 puk")
		// sm2cert := sw.ParseX509Certificate2Sm2(cert)
		// get the public key in the right format
		certPubK, err = csp.KeyImport(cert, &bccsp.X509PublicKeyImportOpts{Temporary: true})
		if err != nil {
			return nil, nil, errors.WithMessage(err, "Failed to import certificate's public key")
		}
	case ecdsa.PublicKey:
		log.Infof("ecdsa cert is sm2 puk")
		sm2cert := sw.ParseSm2Certificate2X509(cert)
		// get the public key in the right format
		certPubK, err = csp.KeyImport(sm2cert, &bccsp.X509PublicKeyImportOpts{Temporary: true})
		if err != nil {
			return nil, nil, errors.WithMessage(err, "Failed to import certificate's public key")
		}
	default:
		log.Infof("gm sign cert is default puk")
		// get the public key in the right format
		certPubK, err = csp.KeyImport(cert, &bccsp.SM2PrivateKeyImportOpts{Temporary: true})
		if err != nil {
			return nil, nil, errors.WithMessage(err, "Failed to import certificate's public key")
		}
	}

	// Get the key given the SKI value
	ski := certPubK.SKI()
	privateKey, err := csp.GetKey(ski)
	if err != nil {
		return nil, nil, errors.WithMessage(err, "Could not find matching private key for SKI")
	}
	// BCCSP returns a public key if the private key for the SKI wasn't found, so
	// we need to return an error in that case.
	if !privateKey.Private() {
		return nil, nil, errors.Errorf("The private key associated with the certificate with SKI '%s' was not found", hex.EncodeToString(ski))
	}
	// Construct and initialize the signer
	signer, err := cspsigner.New(csp, privateKey)
	if err != nil {
		return nil, nil, errors.WithMessage(err, "Failed to load ski from bccsp")
	}
	return privateKey, signer, nil
}

// GetSignerFromCert load private key represented by ski and return bccsp signer that conforms to crypto.Signer
func GetSignerFromCert(cert *x509.Certificate, csp core.CryptoSuite) (core.Key, crypto.Signer, error) {
	if csp == nil {
		return nil, nil, errors.New("CSP was not initialized")
	}

	//add by liuhy
	var (
		certPubK core.Key
		err      error
	)

	switch cert.PublicKey.(type) {
	case *sm2.PublicKey:
		log.Infof("gm cert is sm2 puk")
		// sm2cert := sw.ParseX509Certificate2Sm2(cert)
		// get the public key in the right format
		certPubK, err = csp.KeyImport(cert, &bccsp.X509PublicKeyImportOpts{Temporary: true})
		if err != nil {
			return nil, nil, errors.WithMessage(err, "Failed to import certificate's public key")
		}
	case *ecdsa.PublicKey:
		log.Infof("esdsa sign cert is default puk")
		certPubK, err = csp.KeyImport(cert, &bccsp.X509PublicKeyImportOpts{Temporary: true})
		if err != nil {
			return nil, nil, errors.WithMessage(err, "Failed to import certificate's public key")
		}
	default:
		log.Infof("esdsa sign cert is default puk")
		// sm2cert := sw.ParseX509Certificate2Sm2(cert)
		// get the public key in the right format
		certPubK, err = csp.KeyImport(cert, &bccsp.X509PublicKeyImportOpts{Temporary: true})
		if err != nil {
			return nil, nil, errors.WithMessage(err, "Failed to import certificate's public key")
		}
	}

	// // get the public key in the right format
	// certPubK, err := csp.KeyImport(cert, factory.GetX509PublicKeyImportOpts(true))
	// if err != nil {
	// 	return nil, nil, errors.WithMessage(err, "Failed to import certificate's public key")
	// }
	// Get the key given the SKI value
	ski := certPubK.SKI()
	privateKey, err := csp.GetKey(ski)
	if err != nil {
		return nil, nil, errors.WithMessage(err, "Could not find matching private key for SKI")
	}
	// BCCSP returns a public key if the private key for the SKI wasn't found, so
	// we need to return an error in that case.
	if !privateKey.Private() {
		return nil, nil, errors.Errorf("The private key associated with the certificate with SKI '%s' was not found", hex.EncodeToString(ski))
	}
	// Construct and initialize the signer
	signer, err := factory.NewCspSigner(csp, privateKey)
	if err != nil {
		return nil, nil, errors.WithMessage(err, "Failed to load ski from bccsp")
	}
	return privateKey, signer, nil
}

// GetSignerFromCertFile load skiFile and load private key represented by ski and return bccsp signer that conforms to crypto.Signer
func GetSignerFromCertFile(certFile string, csp core.CryptoSuite) (core.Key, crypto.Signer, *x509.Certificate, error) {
	// Load cert file
	certBytes, err := ioutil.ReadFile(certFile)
	if err != nil {
		return nil, nil, nil, errors.Wrapf(err, "Could not read certFile '%s'", certFile)
	}
	// Parse certificate
	parsedCa, err := helpers.ParseCertificatePEM(certBytes)
	if err != nil {
		return nil, nil, nil, err
	}
	// Get the signer from the cert
	key, cspSigner, err := GetSignerFromCert(parsedCa, csp)
	return key, cspSigner, parsedCa, err
}

// BCCSPKeyRequestGenerate generates keys through BCCSP
// somewhat mirroring to cfssl/req.KeyRequest.Generate()
func BCCSPKeyRequestGenerate(req *csr.CertificateRequest, myCSP core.CryptoSuite) (core.Key, crypto.Signer, error) {
	log.Infof("generating key: %+v", req.KeyRequest)
	keyOpts, err := getBCCSPKeyOpts(req.KeyRequest, false)
	if err != nil {
		return nil, nil, err
	}
	key, err := myCSP.KeyGen(keyOpts)
	if err != nil {
		return nil, nil, err
	}
	cspSigner, err := factory.NewCspSigner(myCSP, key)
	if err != nil {
		return nil, nil, errors.WithMessage(err, "Failed initializing CryptoSigner")
	}
	return key, cspSigner, nil
}

// ImportBCCSPKeyFromPEM attempts to create a private BCCSP key from a pem file keyFile
func ImportBCCSPKeyFromPEM(keyFile string, myCSP core.CryptoSuite, temporary bool) (core.Key, error) {
	keyBuff, err := ioutil.ReadFile(keyFile)
	if err != nil {
		return nil, err
	}
	key, err := ImportBCCSPKeyFromPEMBytes(keyBuff, myCSP, temporary)
	if err != nil {
		return nil, errors.WithMessage(err, fmt.Sprintf("Failed parsing private key from key file %s", keyFile))
	}
	return key, nil
}

// ImportBCCSPKeyFromPEMBytes attempts to create a private BCCSP key from a pem byte slice
func ImportBCCSPKeyFromPEMBytes(keyBuff []byte, myCSP core.CryptoSuite, temporary bool) (core.Key, error) {
	keyFile := "pem bytes"

	key, err := factory.PEMtoPrivateKey(keyBuff, nil)
	if err != nil {
		return nil, errors.WithMessage(err, fmt.Sprintf("Failed parsing private key from %s", keyFile))
	}
	switch key.(type) {
	case *sm2.PrivateKey:
		priv, err := sw.PrivateKeyToDER(key.(*sm2.PrivateKey))
		if err != nil {
			return nil, errors.WithMessage(err, fmt.Sprintf("Failed to convert sm2 private key for '%s'", keyFile))
		}

		sk, err := myCSP.KeyImport(priv, &bccsp.SM2PrivateKeyImportOpts{Temporary: temporary})
		if err != nil {
			return nil, errors.WithMessage(err, fmt.Sprintf("Failed to import sm2 private key for '%s'", keyFile))
		}
		return sk, nil
	case *ecdsa.PrivateKey:
		priv, err := factory.PrivateKeyToDER(key.(*ecdsa.PrivateKey))
		if err != nil {
			return nil, errors.WithMessage(err, fmt.Sprintf("Failed to convert ECDSA private key for '%s'", keyFile))
		}
		sk, err := myCSP.KeyImport(priv, factory.GetECDSAPrivateKeyImportOpts(temporary))
		if err != nil {
			return nil, errors.WithMessage(err, fmt.Sprintf("Failed to import ECDSA private key for '%s'", keyFile))
		}
		return sk, nil
	default:
		return nil, errors.Errorf("Failed to import key from %s: invalid secret key type", keyFile)
	}
}

// LoadX509KeyPair reads and parses a public/private key pair from a pair
// of files. The files must contain PEM encoded data. The certificate file
// may contain intermediate certificates following the leaf certificate to
// form a certificate chain. On successful return, Certificate.Leaf will
// be nil because the parsed form of the certificate is not retained.
//
// This function originated from crypto/tls/tls.go and was adapted to use a
// BCCSP Signer
func LoadX509KeyPair(certFile, keyFile []byte, csp core.CryptoSuite) (*tls.Certificate, error) {

	certPEMBlock := certFile

	cert := &tls.Certificate{}
	var skippedBlockTypes []string
	for {
		var certDERBlock *pem.Block
		certDERBlock, certPEMBlock = pem.Decode(certPEMBlock)
		if certDERBlock == nil {
			break
		}
		if certDERBlock.Type == "CERTIFICATE" {
			cert.Certificate = append(cert.Certificate, certDERBlock.Bytes)
		} else {
			skippedBlockTypes = append(skippedBlockTypes, certDERBlock.Type)
		}
	}

	if len(cert.Certificate) == 0 {
		if len(skippedBlockTypes) == 0 {
			return nil, errors.New("Failed to find PEM block in bytes")
		}
		if len(skippedBlockTypes) == 1 && strings.HasSuffix(skippedBlockTypes[0], "PRIVATE KEY") {
			return nil, errors.New("Failed to find certificate PEM data in bytes, but did find a private key; PEM inputs may have been switched")
		}
		return nil, errors.Errorf("Failed to find \"CERTIFICATE\" PEM block in file %s after skipping PEM blocks of the following types: %v", certFile, skippedBlockTypes)
	}

	x509Cert, err := x509.ParseCertificate(cert.Certificate[0])
	if err != nil {
		return nil, err
	}

	_, cert.PrivateKey, err = GetSignerFromCert(x509Cert, csp)
	if err != nil {
		if keyFile != nil {
			log.Debugf("Could not load TLS certificate with BCCSP: %s", err)
			log.Debug("Attempting fallback with provided certfile and keyfile")
			fallbackCerts, err := tls.X509KeyPair(certFile, keyFile)
			if err != nil {
				return nil, errors.Wrap(err, "Could not get the private key that matches the provided cert")
			}
			cert = &fallbackCerts
		} else {
			return nil, errors.WithMessage(err, "Could not load TLS certificate with BCCSP")
		}

	}

	return cert, nil
}

func LoadX509KeyPairGM(certFile, keyFile string, certPem []byte, keyPem []byte, csp core.CryptoSuite) (*gmtls.Certificate, error) {
	var (
		certPEMBlock []byte
		err          error
	)

	if certPem == nil {
		certPEMBlock, err = ioutil.ReadFile(certFile)
		if err != nil {
			return nil, err
		}
	} else {
		certPEMBlock = certPem
	}

	cert := &gmtls.Certificate{}
	var skippedBlockTypes []string
	for {
		var certDERBlock *pem.Block
		certDERBlock, certPEMBlock = pem.Decode(certPEMBlock)
		if certDERBlock == nil {
			break
		}
		if certDERBlock.Type == "CERTIFICATE" {
			cert.Certificate = append(cert.Certificate, certDERBlock.Bytes)
		} else {
			skippedBlockTypes = append(skippedBlockTypes, certDERBlock.Type)
		}
	}

	if len(cert.Certificate) == 0 {
		if len(skippedBlockTypes) == 0 {
			return nil, errors.Errorf("Failed to find PEM block in file %s", certFile)
		}
		if len(skippedBlockTypes) == 1 && strings.HasSuffix(skippedBlockTypes[0], "PRIVATE KEY") {
			return nil, errors.Errorf("Failed to find certificate PEM data in file %s, but did find a private key; PEM inputs may have been switched", certFile)
		}
		return nil, errors.Errorf("Failed to find \"CERTIFICATE\" PEM block in file %s after skipping PEM blocks of the following types: %v", certFile, skippedBlockTypes)
	}

	//add by liuhy
	gmCert, err := gmx509.ParseCertificate(cert.Certificate[0])
	if err != nil {
		return nil, err
	}
	x509Cert := sw.ParseSm2Certificate2X509(gmCert)

	_, cert.PrivateKey, err = GetSignerFromCert(x509Cert, csp)
	if err != nil {
		if len(keyFile) == 0 {
			log.Debugf("Could not load TLS certificate with BCCSP: %s", err)
			log.Debugf("Attempting fallback with certfile %s and keyfile %s", certFile, keyFile)

			if certPem == nil && keyPem == nil {
				fallbackCerts, err := gmtls.LoadX509KeyPair(certFile, keyFile)
				if err != nil {
					return nil, errors.Wrapf(err, "Could not get the private key %s that matches %s", keyFile, certFile)
				}
				cert = &fallbackCerts
			} else {
				//add by liuhy
				fallbackCerts, err := gmtls.X509KeyPair(certPem, keyPem)
				if err != nil {
					return nil, errors.Wrapf(err, "Could not get the private key %s that matches %s", keyFile, certFile)
				}
				cert = &fallbackCerts
			}
		} else {
			return nil, errors.WithMessage(err, "Could not load TLS certificate with BCCSP")
		}

	}

	return cert, nil
}
