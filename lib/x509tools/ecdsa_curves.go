/*
 * Copyright (c) SAS Institute Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package x509tools

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"encoding/asn1"
	"errors"
	"fmt"
	"math/big"
	"strconv"
	"strings"
)

// Predefined named ECDSA curve
type CurveDefinition struct {
	Bits  uint
	Curve elliptic.Curve
	Oid   asn1.ObjectIdentifier
}

var DefinedCurves = []CurveDefinition{
	{256, elliptic.P256(), asn1.ObjectIdentifier{1, 2, 840, 10045, 3, 1, 7}},
	{384, elliptic.P384(), asn1.ObjectIdentifier{1, 3, 132, 0, 34}},
	{521, elliptic.P521(), asn1.ObjectIdentifier{1, 3, 132, 0, 35}},
}

// Return the DER encoding of the ASN.1 OID of this named curve
func (def *CurveDefinition) ToDer() []byte {
	der, err := asn1.Marshal(def.Oid)
	if err != nil {
		panic(err)
	}
	return der
}

// Return the names of all supported ECDSA curves
func SupportedCurves() string {
	curves := make([]string, len(DefinedCurves))
	for i, def := range DefinedCurves {
		curves[i] = strconv.FormatUint(uint64(def.Bits), 10)
	}
	return strings.Join(curves, ", ")
}

// Get a curve by its ASN.1 object identifier
func CurveByOid(oid asn1.ObjectIdentifier) (*CurveDefinition, error) {
	for _, def := range DefinedCurves {
		if oid.Equal(def.Oid) {
			return &def, nil
		}
	}
	return nil, fmt.Errorf("Unsupported ECDSA curve with OID: %s\nSupported curves: %s", oid, SupportedCurves())
}

// Get a curve by a dotted decimal OID string
func CurveByOidString(oidstr string) (*CurveDefinition, error) {
	parts := strings.Split(oidstr, ".")
	oid := make(asn1.ObjectIdentifier, 0, len(parts))
	for _, n := range parts {
		v, err := strconv.Atoi(n)
		if err != nil {
			return nil, errors.New("invalid OID")
		}
		oid = append(oid, v)
	}
	return CurveByOid(oid)
}

// Get a curve by the DER encoding of its OID
func CurveByDer(der []byte) (*CurveDefinition, error) {
	var oid asn1.ObjectIdentifier
	_, err := asn1.Unmarshal(der, &oid)
	if err != nil {
		return nil, err
	}
	return CurveByOid(oid)
}

// Get a curve by an elliptic.Curve value
func CurveByCurve(curve elliptic.Curve) (*CurveDefinition, error) {
	for _, def := range DefinedCurves {
		if curve == def.Curve {
			return &def, nil
		}
	}
	return nil, fmt.Errorf("Unsupported ECDSA curve: %v\nSupported curves: %s", curve, SupportedCurves())
}

// Get a curve by a number of bits
func CurveByBits(bits uint) (*CurveDefinition, error) {
	for _, def := range DefinedCurves {
		if bits == def.Bits {
			return &def, nil
		}
	}
	return nil, fmt.Errorf("Unsupported ECDSA curve: %v\nSupported curves: %s", bits, SupportedCurves())
}

// Decode an ECDSA public key from its DER encoding. Both octet and bitstring
// encodings are supported.
func DerToPoint(curve elliptic.Curve, der []byte) (*big.Int, *big.Int) {
	var blob []byte
	switch der[0] {
	case asn1.TagOctetString:
		_, err := asn1.Unmarshal(der, &blob)
		if err != nil {
			return nil, nil
		}
	case asn1.TagBitString:
		var bits asn1.BitString
		_, err := asn1.Unmarshal(der, &bits)
		if err != nil {
			return nil, nil
		}
		blob = bits.Bytes
	default:
		return nil, nil
	}
	return elliptic.Unmarshal(curve, blob)
}

func PointToDer(pub *ecdsa.PublicKey) []byte {
	blob := elliptic.Marshal(pub.Curve, pub.X, pub.Y)
	der, err := asn1.Marshal(blob)
	if err != nil {
		return nil
	}
	return der
}
