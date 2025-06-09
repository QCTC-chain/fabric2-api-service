package lib

import (
	"crypto"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/asn1"
	"fmt"
	"net"
	"net/mail"

	"github.com/cloudflare/cfssl/csr"
	"github.com/cloudflare/cfssl/log"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/core"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/cryptosuite/bccsp/wrapper"

	// "github.com/hyperledger/fabric/bccsp"

	// "github.com/hyperledger/fabric/bccsp/sw"

	"github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/bccsp"
	"github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/bccsp/sw"

	"gitee.com/china_uni/tjfoc-gm/sm2"
	gmx509 "gitee.com/china_uni/tjfoc-gm/x509"
)

// //证书签名
// func signCert(req signer.SignRequest, ca *CA, isTLS bool) (cert []byte, err error) {
// 	/*csr := parseCertificateRequest()
// 	cert, err := sm2.CreateCertificateToMem(template, rootca, csr.pubkey, rootca.privkey)
// 	sm2Cert, err := sm2.parseCertificateFromMem(cert)

// 	var certRecord = certdb.CertificateRecord{
// 		Serial:  sm2Cert.SerialNumber.String(),
// 		AKI:     hex.EncodeToString(sm2Cert.AuthorityKeyId),
// 		CALabel: req.Label,
// 		Status:  "good",
// 		Expiry:  sm2Cert.NotAfter,
// 		PEM:     string(cert),
// 	}*/

// 	block, _ := pem.Decode([]byte(req.Request))
// 	if block == nil {
// 		return nil, fmt.Errorf("decode error")
// 	}
// 	if block.Type != "NEW CERTIFICATE REQUEST" && block.Type != "CERTIFICATE REQUEST" {
// 		return nil, fmt.Errorf("not a csr")
// 	}
// 	template, err := parseCertificateRequest(block.Bytes)
// 	if err != nil {
// 		log.Infof("xxxx gmca.go ParseCertificateRequest error:[%s]", err)
// 		return nil, err
// 	}

// 	//add by liuhy
// 	if req.NotAfter.IsZero() {
// 		template.NotBefore = time.Now().Round(time.Minute)
// 		template.NotAfter = template.NotBefore.Add(defaultIssuedCertificateExpiration).UTC()
// 	} else {
// 		template.NotBefore = time.Now().Round(time.Minute)
// 		template.NotAfter = req.NotAfter
// 	}

// 	certfile := ca.Config.CA.Certfile

// 	rootkey, _, x509cert, err := util.GetSignerFromCertFile(certfile, ca.csp)
// 	if err != nil {

// 		return nil, err
// 	}
// 	rootca := sw.ParseX509Certificate2Sm2(x509cert)

// 	cert, err = sw.CreateCertificateToMem(template, rootca, rootkey)
// 	if err != nil {
// 		return nil, err
// 	}

// 	//tls not save cert to db
// 	if isTLS {
// 		return
// 	}

// 	// log.Infof("template = %v\n cert = %v\n Type = %T", template, cert, template.PublicKey)
// 	clientCert, err := gmx509.ReadCertificateFromMem(cert)
// 	// log.Info("Exit ParseCertificate")
// 	if err == nil {
// 		log.Infof("xxxx gmca.go signCert ok the sign cert len [%d]", len(cert))
// 	}

// 	var certRecord = certdb.CertificateRecord{
// 		Serial:  clientCert.SerialNumber.String(),
// 		AKI:     hex.EncodeToString(clientCert.AuthorityKeyId),
// 		CALabel: req.Label,
// 		Status:  "good",
// 		Expiry:  clientCert.NotAfter,
// 		PEM:     string(cert),
// 	}
// 	//aki := hex.EncodeToString(cert.AuthorityKeyId)
// 	//serial := util.GetSerialAsHex(cert.SerialNumber)

// 	err = ca.certDBAccessor.InsertCertificate(certRecord)
// 	if err != nil {
// 		log.Info("error InsertCertificate!")
// 	}

// 	return
// }

// //生成证书
// func createGmSm2Cert(key bccsp.Key, req *csr.CertificateRequest, priv crypto.Signer) (cert []byte, err error) {
// 	log.Infof("xxx xxx in gmca.go  createGmSm2Cert...key :%T", key)

// 	csrPEM, err := generate(priv, req, key)
// 	if err != nil {
// 		log.Infof("xxxxxxxxxxxxx create csr error:%s", err)
// 	}
// 	log.Infof("xxxxxxxxxxxxx create gm csr completed!")
// 	block, _ := pem.Decode(csrPEM)
// 	if block == nil {
// 		return nil, fmt.Errorf("sm2 csr DecodeFailed")
// 	}

// 	if block.Type != "NEW CERTIFICATE REQUEST" && block.Type != "CERTIFICATE REQUEST" {
// 		return nil, fmt.Errorf("sm2 not a csr")
// 	}
// 	sm2Template, err := parseCertificateRequest(block.Bytes)
// 	if err != nil {
// 		log.Infof("parseCertificateRequest return err:%s", err)
// 		return nil, err
// 	}

// 	if req.CA != nil && req.CA.Expiry != "" {
// 		sm2Template.NotBefore = time.Now().Round(time.Minute)
// 		sm2Template.NotAfter = sm2Template.NotBefore.Add(parseDuration(req.CA.Expiry))
// 	}

// 	log.Infof("key is %T   ---   %T", sm2Template.PublicKey, sm2Template)
// 	cert, err = sw.CreateCertificateToMem(sm2Template, sm2Template, key)
// 	return cert, err
// }

// //证书请求转换成证书 参数为  block.Bytes
// func parseCertificateRequest(csrBytes []byte) (template *gmx509.Certificate, err error) {
// 	csrv, err := gmx509.ParseCertificateRequest(csrBytes)
// 	if err != nil {
// 		// err = cferr.Wrap(cferr.CSRError, cferr.ParseFailed, err)
// 		return
// 	}
// 	err = csrv.CheckSignature()
// 	if err != nil {
// 		// err = cferr.Wrap(cferr.CSRError, cferr.KeyMismatch, err)
// 		return
// 	}

// 	template = &gmx509.Certificate{
// 		Subject:            csrv.Subject,
// 		PublicKeyAlgorithm: csrv.PublicKeyAlgorithm,
// 		PublicKey:          csrv.PublicKey,
// 		SignatureAlgorithm: csrv.SignatureAlgorithm,
// 		DNSNames:           csrv.DNSNames,
// 		IPAddresses:        csrv.IPAddresses,
// 		EmailAddresses:     csrv.EmailAddresses,
// 	}

// 	log.Infof("request algorithm = %v, %v\n", template.PublicKeyAlgorithm, template.SignatureAlgorithm)
// 	log.Infof("publicKey type: %T", template.PublicKey)

// 	// template.NotBefore = time.Now().Round(time.Minute)
// 	// template.NotAfter = template.NotBefore.Add(defaultIssuedCertificateExpiration).UTC()

// 	//log.Infof("-----------csrv = %+v", csrv)
// 	for _, val := range csrv.Extensions {
// 		// Check the CSR for the X.509 BasicConstraints (RFC 5280, 4.2.1.9)
// 		// extension and append to template if necessary
// 		if val.Id.Equal(asn1.ObjectIdentifier{2, 5, 29, 19}) {
// 			var constraints csr.BasicConstraints
// 			var rest []byte

// 			if rest, err = asn1.Unmarshal(val.Value, &constraints); err != nil {
// 				//return nil, cferr.Wrap(cferr.CSRError, cferr.ParseFailed, err)
// 			} else if len(rest) != 0 {
// 				//return nil, cferr.Wrap(cferr.CSRError, cferr.ParseFailed, errors.New("x509: trailing data after X.509 BasicConstraints"))
// 			}

// 			template.BasicConstraintsValid = true
// 			template.IsCA = constraints.IsCA
// 			template.MaxPathLen = constraints.MaxPathLen
// 			template.MaxPathLenZero = template.MaxPathLen == 0
// 		}
// 	}
// 	serialNumber := make([]byte, 20)
// 	_, err = io.ReadFull(rand.Reader, serialNumber)
// 	if err != nil {
// 		return nil, err
// 	}

// 	// SetBytes interprets buf as the bytes of a big-endian
// 	// unsigned integer. The leading byte should be masked
// 	// off to ensure it isn't negative.
// 	serialNumber[0] &= 0x7F

// 	template.SerialNumber = new(big.Int).SetBytes(serialNumber)

// 	return
// }

//cloudflare 证书请求 转成 国密证书请求
func generate(priv crypto.Signer, req *csr.CertificateRequest, key core.Key) (csr []byte, err error) {
	log.Info("xx entry gm generate")
	sigAlgo := signerAlgo(priv)
	if sigAlgo == gmx509.UnknownSignatureAlgorithm {
		return nil, fmt.Errorf("Private key is unavailable")
	}
	log.Info("xx begin create sm2.CertificateRequest")
	var tpl = gmx509.CertificateRequest{
		Subject:            req.Name(),
		SignatureAlgorithm: sigAlgo,
	}
	for i := range req.Hosts {
		if ip := net.ParseIP(req.Hosts[i]); ip != nil {
			tpl.IPAddresses = append(tpl.IPAddresses, ip)
		} else if email, err := mail.ParseAddress(req.Hosts[i]); err == nil && email != nil {
			tpl.EmailAddresses = append(tpl.EmailAddresses, email.Address)
		} else {
			tpl.DNSNames = append(tpl.DNSNames, req.Hosts[i])
		}
	}

	if req.CA != nil {
		err = appendCAInfoToCSRSm2(req.CA, &tpl)
		if err != nil {
			err = fmt.Errorf("sm2 GenerationFailed")
			return
		}
	}
	// if req.SerialNumber != "" {
	// }

	//add by liuhy
	var bccspKey bccsp.Key
	if k, ok := key.(*wrapper.Key); ok {
		bccspKey = k.Key
	} else {
		log.Error("wrapper core.key to bccsp.Key failt")
		return nil, fmt.Errorf("%s", "wrapper core.key to bccsp.Key failt")
	}
	csr, err = sw.CreateSm2CertificateRequestToMem(&tpl, bccspKey)
	log.Info("xx exit generate")
	return csr, err
}

func signerAlgo(priv crypto.Signer) gmx509.SignatureAlgorithm {
	switch pub := priv.Public().(type) {
	case *sm2.PublicKey:
		switch pub.Curve {
		case sm2.P256Sm2():
			return gmx509.SM2WithSM3
		default:
			return gmx509.SM2WithSM3
		}
	default:
		return gmx509.UnknownSignatureAlgorithm
	}
}

// appendCAInfoToCSR appends CAConfig BasicConstraint extension to a CSR
func appendCAInfoToCSR(reqConf *csr.CAConfig, csreq *x509.CertificateRequest) error {
	pathlen := reqConf.PathLength
	if pathlen == 0 && !reqConf.PathLenZero {
		pathlen = -1
	}
	val, err := asn1.Marshal(csr.BasicConstraints{true, pathlen})

	if err != nil {
		return err
	}

	csreq.ExtraExtensions = []pkix.Extension{
		{
			Id:       asn1.ObjectIdentifier{2, 5, 29, 19},
			Value:    val,
			Critical: true,
		},
	}
	return nil
}

// appendCAInfoToCSR appends CAConfig BasicConstraint extension to a CSR
func appendCAInfoToCSRSm2(reqConf *csr.CAConfig, csreq *gmx509.CertificateRequest) error {
	pathlen := reqConf.PathLength
	if pathlen == 0 && !reqConf.PathLenZero {
		pathlen = -1
	}
	val, err := asn1.Marshal(csr.BasicConstraints{true, pathlen})

	if err != nil {
		return err
	}

	csreq.ExtraExtensions = []pkix.Extension{
		{
			Id:       asn1.ObjectIdentifier{2, 5, 29, 19},
			Value:    val,
			Critical: true,
		},
	}

	return nil
}

// func ParseX509Certificate2Sm2(x509Cert *x509.Certificate) *gmx509.Certificate {
// 	sm2cert := &gmx509.Certificate{
// 		Raw:                         x509Cert.Raw,
// 		RawTBSCertificate:           x509Cert.RawTBSCertificate,
// 		RawSubjectPublicKeyInfo:     x509Cert.RawSubjectPublicKeyInfo,
// 		RawSubject:                  x509Cert.RawSubject,
// 		RawIssuer:                   x509Cert.RawIssuer,
// 		Signature:                   x509Cert.Signature,
// 		SignatureAlgorithm:          gmx509.SignatureAlgorithm(x509Cert.SignatureAlgorithm),
// 		PublicKeyAlgorithm:          gmx509.PublicKeyAlgorithm(x509Cert.PublicKeyAlgorithm),
// 		PublicKey:                   x509Cert.PublicKey,
// 		Version:                     x509Cert.Version,
// 		SerialNumber:                x509Cert.SerialNumber,
// 		Issuer:                      x509Cert.Issuer,
// 		Subject:                     x509Cert.Subject,
// 		NotBefore:                   x509Cert.NotBefore,
// 		NotAfter:                    x509Cert.NotAfter,
// 		KeyUsage:                    gmx509.KeyUsage(x509Cert.KeyUsage),
// 		Extensions:                  x509Cert.Extensions,
// 		ExtraExtensions:             x509Cert.ExtraExtensions,
// 		UnhandledCriticalExtensions: x509Cert.UnhandledCriticalExtensions,
// 		//ExtKeyUsage:	[]x509.ExtKeyUsage(x509Cert.ExtKeyUsage) ,
// 		UnknownExtKeyUsage:    x509Cert.UnknownExtKeyUsage,
// 		BasicConstraintsValid: x509Cert.BasicConstraintsValid,
// 		IsCA:                  x509Cert.IsCA,
// 		MaxPathLen:            x509Cert.MaxPathLen,
// 		// MaxPathLenZero indicates that BasicConstraintsValid==true and
// 		// MaxPathLen==0 should be interpreted as an actual maximum path length
// 		// of zero. Otherwise, that combination is interpreted as MaxPathLen
// 		// not being set.
// 		MaxPathLenZero: x509Cert.MaxPathLenZero,
// 		SubjectKeyId:   x509Cert.SubjectKeyId,
// 		AuthorityKeyId: x509Cert.AuthorityKeyId,
// 		// RFC 5280, 4.2.2.1 (Authority Information Access)
// 		OCSPServer:            x509Cert.OCSPServer,
// 		IssuingCertificateURL: x509Cert.IssuingCertificateURL,
// 		// Subject Alternate Name values
// 		DNSNames:       x509Cert.DNSNames,
// 		EmailAddresses: x509Cert.EmailAddresses,
// 		IPAddresses:    x509Cert.IPAddresses,
// 		// Name constraints
// 		PermittedDNSDomainsCritical: x509Cert.PermittedDNSDomainsCritical,
// 		PermittedDNSDomains:         x509Cert.PermittedDNSDomains,
// 		// CRL Distribution Points
// 		CRLDistributionPoints: x509Cert.CRLDistributionPoints,
// 		PolicyIdentifiers:     x509Cert.PolicyIdentifiers,
// 	}
// 	for _, val := range x509Cert.ExtKeyUsage {
// 		sm2cert.ExtKeyUsage = append(sm2cert.ExtKeyUsage, gmx509.ExtKeyUsage(val))
// 	}
// 	return sm2cert
// }
