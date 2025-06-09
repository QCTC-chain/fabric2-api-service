/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/
/*
Notice: This file has been modified for Hyperledger Fabric SDK Go usage.
Please review third_party pinning scripts and patches for more details.
*/

package lib

import (
	//"crypto/tls"
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"

	//"net/http"
	"gitee.com/china_uni/tjfoc-gm/net/http"
	"gitee.com/china_uni/tjfoc-gm/tls"
	gmx509 "gitee.com/china_uni/tjfoc-gm/x509"
	"github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric-ca/sdkinternal/pkg/util"
	"github.com/pkg/errors"
)

var clientAuthTypes = map[string]tls.ClientAuthType{
	"noclientcert":               tls.NoClientCert,
	"requestclientcert":          tls.RequestClientCert,
	"requireanyclientcert":       tls.RequireAnyClientCert,
	"verifyclientcertifgiven":    tls.VerifyClientCertIfGiven,
	"requireandverifyclientcert": tls.RequireAndVerifyClientCert,
}

// GetCertID returns both the serial number and AKI (Authority Key ID) for the certificate
func GetCertID(bytes []byte) (string, string, error) {
	cert, err := BytesToX509Cert(bytes)
	if err != nil {
		return "", "", err
	}
	serial := util.GetSerialAsHex(cert.SerialNumber)
	aki := hex.EncodeToString(cert.AuthorityKeyId)
	return serial, aki, nil
}

// BytesToX509Cert converts bytes (PEM or DER) to an X509 certificate
func BytesToX509Cert(bytes []byte) (*x509.Certificate, error) {
	dcert, _ := pem.Decode(bytes)
	if dcert != nil {
		bytes = dcert.Bytes
	}
	cert, err := x509.ParseCertificate(bytes)
	if err != nil {
		return nil, errors.Wrap(err, "Buffer was neither PEM nor DER encoding")
	}
	return cert, err
}

func addQueryParm(req *http.Request, name, value string) {
	url := req.URL.Query()
	url.Add(name, value)
	req.URL.RawQuery = url.Encode()
}

// CertificateDecoder is needed to keep track of state, to see how many certificates
// have been returned for each enrollment ID.
type CertificateDecoder struct {
	certIDCount map[string]int
	storePath   string
}

// // LoadPEMCertPool loads a pool of PEM certificates from list of files
// func LoadPEMCertPool(certFiles []string) (*gmx509.CertPool, error) {
// 	certPool := gmx509.NewCertPool()

// 	if len(certFiles) > 0 {
// 		for _, cert := range certFiles {
// 			log.Debugf("Reading cert file: %s", cert)
// 			pemCerts, err := ioutil.ReadFile(cert)
// 			if err != nil {
// 				return nil, err
// 			}

// 			log.Debugf("Appending cert %s to pool", cert)
// 			if !certPool.AppendCertsFromPEM(pemCerts) {
// 				return nil, errors.New("Failed to load cert pool")
// 			}
// 		}
// 	}

// 	return certPool, nil
// }

// SM2证书请求 转换 X509 证书请求
func ParseGMCertificateRequest2X509(sm2req *gmx509.CertificateRequest) *x509.CertificateRequest {
	x509req := &x509.CertificateRequest{
		Raw:                      sm2req.Raw,                      // Complete ASN.1 DER content (CSR, signature algorithm and signature).
		RawTBSCertificateRequest: sm2req.RawTBSCertificateRequest, // Certificate request info part of raw ASN.1 DER content.
		RawSubjectPublicKeyInfo:  sm2req.RawSubjectPublicKeyInfo,  // DER encoded SubjectPublicKeyInfo.
		RawSubject:               sm2req.RawSubject,               // DER encoded Subject.

		Version:            sm2req.Version,
		Signature:          sm2req.Signature,
		SignatureAlgorithm: x509.SignatureAlgorithm(sm2req.SignatureAlgorithm),

		PublicKeyAlgorithm: x509.PublicKeyAlgorithm(sm2req.PublicKeyAlgorithm),
		PublicKey:          sm2req.PublicKey,

		Subject: sm2req.Subject,

		// Attributes is the dried husk of a bug and shouldn't be used.
		Attributes: sm2req.Attributes,

		// Extensions contains raw X.509 extensions. When parsing CSRs, this
		// can be used to extract extensions that are not parsed by this
		// package.
		Extensions: sm2req.Extensions,

		// ExtraExtensions contains extensions to be copied, raw, into any
		// marshaled CSR. Values override any extensions that would otherwise
		// be produced based on the other fields but are overridden by any
		// extensions specified in Attributes.
		//
		// The ExtraExtensions field is not populated when parsing CSRs, see
		// Extensions.
		ExtraExtensions: sm2req.ExtraExtensions,

		// Subject Alternate Name values.
		DNSNames:       sm2req.DNSNames,
		EmailAddresses: sm2req.EmailAddresses,
		IPAddresses:    sm2req.IPAddresses,
	}
	return x509req
}
