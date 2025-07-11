/*
Copyright Suzhou Tongji Fintech Research Institute 2017 All Rights Reserved.
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

package x509

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rsa"
	"encoding/asn1"
	"errors"
	"fmt"
	"math/big"

	"gitee.com/china_uni/tjfoc-gm/sm2"
)

const ecPrivKeyVersion = 1

// pkcs1PrivateKey is a structure which mirrors the PKCS#1 ASN.1 for an RSA private key.
type pkcs1PrivateKey struct {
	Version int
	N       *big.Int
	E       int
	D       *big.Int
	P       *big.Int
	Q       *big.Int
	// We ignore these values, if present, because rsa will calculate them.
	Dp   *big.Int `asn1:"optional"`
	Dq   *big.Int `asn1:"optional"`
	Qinv *big.Int `asn1:"optional"`

	AdditionalPrimes []pkcs1AdditionalRSAPrime `asn1:"optional,omitempty"`
}

type pkcs1AdditionalRSAPrime struct {
	Prime *big.Int

	// We ignore these values because rsa will calculate them.
	Exp   *big.Int
	Coeff *big.Int
}

// ParsePKCS1PrivateKey returns an RSA private key from its ASN.1 PKCS#1 DER encoded form.
func ParsePKCS1PrivateKey(der []byte) (*rsa.PrivateKey, error) {
	var priv pkcs1PrivateKey
	rest, err := asn1.Unmarshal(der, &priv)
	if len(rest) > 0 {
		return nil, asn1.SyntaxError{Msg: "trailing data"}
	}
	if err != nil {
		return nil, err
	}

	if priv.Version > 1 {
		return nil, errors.New("x509: unsupported private key version")
	}

	if priv.N.Sign() <= 0 || priv.D.Sign() <= 0 || priv.P.Sign() <= 0 || priv.Q.Sign() <= 0 {
		return nil, errors.New("x509: private key contains zero or negative value")
	}

	key := new(rsa.PrivateKey)
	key.PublicKey = rsa.PublicKey{
		E: priv.E,
		N: priv.N,
	}

	key.D = priv.D
	key.Primes = make([]*big.Int, 2+len(priv.AdditionalPrimes))
	key.Primes[0] = priv.P
	key.Primes[1] = priv.Q
	for i, a := range priv.AdditionalPrimes {
		if a.Prime.Sign() <= 0 {
			return nil, errors.New("x509: private key contains zero or negative prime")
		}
		key.Primes[i+2] = a.Prime
		// We ignore the other two values because rsa will calculate
		// them as needed.
	}

	err = key.Validate()
	if err != nil {
		return nil, err
	}
	key.Precompute()

	return key, nil
}

// MarshalPKCS1PrivateKey converts a private key to ASN.1 DER encoded form.
func MarshalPKCS1PrivateKey(key *rsa.PrivateKey) []byte {
	key.Precompute()

	version := 0
	if len(key.Primes) > 2 {
		version = 1
	}

	priv := pkcs1PrivateKey{
		Version: version,
		N:       key.N,
		E:       key.PublicKey.E,
		D:       key.D,
		P:       key.Primes[0],
		Q:       key.Primes[1],
		Dp:      key.Precomputed.Dp,
		Dq:      key.Precomputed.Dq,
		Qinv:    key.Precomputed.Qinv,
	}

	priv.AdditionalPrimes = make([]pkcs1AdditionalRSAPrime, len(key.Precomputed.CRTValues))
	for i, values := range key.Precomputed.CRTValues {
		priv.AdditionalPrimes[i].Prime = key.Primes[2+i]
		priv.AdditionalPrimes[i].Exp = values.Exp
		priv.AdditionalPrimes[i].Coeff = values.Coeff
	}

	b, _ := asn1.Marshal(priv)
	return b
}

// rsaPublicKey reflects the ASN.1 structure of a PKCS#1 public key.
type rsaPublicKey struct {
	N *big.Int
	E int
}

// parseECPrivateKey parses an ASN.1 Elliptic Curve Private Key Structure.
// The OID for the named curve may be provided from another source (such as
// the PKCS8 container) - if it is provided then use this instead of the OID
// that may exist in the EC private key structure.
func parseECPrivateKey(namedCurveOID *asn1.ObjectIdentifier, der []byte) (key interface{}, err error) {
	var privKey ecPrivateKey
	if _, err := asn1.Unmarshal(der, &privKey); err != nil {
		return nil, errors.New("x509: failed to parse EC private key: " + err.Error())
	}
	if privKey.Version != ecPrivKeyVersion {
		return nil, fmt.Errorf("x509: unknown EC private key version %d", privKey.Version)
	}

	var curve elliptic.Curve
	if namedCurveOID != nil {
		curve = namedCurveFromOID(*namedCurveOID)
	} else {
		curve = namedCurveFromOID(privKey.NamedCurveOID)
	}
	if curve == nil {
		return nil, errors.New("x509: unknown elliptic curve")
	}

	k := new(big.Int).SetBytes(privKey.PrivateKey)
	curveOrder := curve.Params().N
	if k.Cmp(curveOrder) >= 0 {
		return nil, errors.New("x509: invalid elliptic curve private key value")
	}

	switch curve {
	case sm2.P256Sm2():
		k := new(big.Int).SetBytes(privKey.PrivateKey)
		curveOrder := curve.Params().N
		if k.Cmp(curveOrder) >= 0 {
			return nil, errors.New("x509: invalid elliptic curve private key value")
		}
		priv := new(sm2.PrivateKey)
		priv.Curve = curve
		priv.D = k

		privateKey := make([]byte, (curveOrder.BitLen()+7)/8)

		// Some private keys have leading zero padding. This is invalid
		// according to [SEC1], but this code will ignore it.
		for len(privKey.PrivateKey) > len(privateKey) {
			if privKey.PrivateKey[0] != 0 {
				return nil, errors.New("x509: invalid private key length")
			}
			privKey.PrivateKey = privKey.PrivateKey[1:]
		}

		// Some private keys remove all leading zeros, this is also invalid
		// according to [SEC1] but since OpenSSL used to do this, we ignore
		// this too.
		copy(privateKey[len(privateKey)-len(privKey.PrivateKey):], privKey.PrivateKey)
		priv.X, priv.Y = curve.ScalarBaseMult(privateKey)

		return priv, nil

	case elliptic.P224(), elliptic.P256(), elliptic.P384(), elliptic.P521():
		k := new(big.Int).SetBytes(privKey.PrivateKey)
		curveOrder := curve.Params().N
		if k.Cmp(curveOrder) >= 0 {
			return nil, errors.New("x509: invalid elliptic curve private key value")
		}
		priv := new(ecdsa.PrivateKey)
		priv.Curve = curve
		priv.D = k

		privateKey := make([]byte, (curveOrder.BitLen()+7)/8)

		// Some private keys have leading zero padding. This is invalid
		// according to [SEC1], but this code will ignore it.
		for len(privKey.PrivateKey) > len(privateKey) {
			if privKey.PrivateKey[0] != 0 {
				return nil, errors.New("x509: invalid private key length")
			}
			privKey.PrivateKey = privKey.PrivateKey[1:]
		}

		// Some private keys remove all leading zeros, this is also invalid
		// according to [SEC1] but since OpenSSL used to do this, we ignore
		// this too.
		copy(privateKey[len(privateKey)-len(privKey.PrivateKey):], privKey.PrivateKey)
		priv.X, priv.Y = curve.ScalarBaseMult(privateKey)

		return priv, nil
	default:
		return nil, errors.New("x509: invalid private key curve param")
	}
}
