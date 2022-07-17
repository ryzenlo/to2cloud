package util

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/asn1"
	"encoding/pem"

	"golang.org/x/crypto/ssh"
)

type RSAKeyPair struct {
	PrivateKey string
	PublicKey  string
	CSRCert    string
}

type CSRMeta struct {
	Email            string `json:"email"`
	CommonName       string `json:"common_name"`
	Country          string `json:"country"`
	Province         string `json:"province"`
	Locality         string `json:"locality"`
	Organization     string `json:"organization"`
	OrganizationUnit string `json:"organizational_unit"`
}

func GeneKeyPair(keybits int, isCSR bool, csrMetaParam map[string]string) (RSAKeyPair, error) {
	keyObj, _ := rsa.GenerateKey(rand.Reader, keybits)
	//
	privateKeyBytes := x509.MarshalPKCS1PrivateKey(keyObj)
	privKeyBytesOfPem := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: privateKeyBytes})
	//
	publickeyObj := &keyObj.PublicKey
	publicKeyBytes, _ := x509.MarshalPKIXPublicKey(publickeyObj)
	pubKeyBytesOfPem := pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: publicKeyBytes})
	//
	pair := RSAKeyPair{
		PrivateKey: string(privKeyBytesOfPem),
		PublicKey:  string(pubKeyBytesOfPem),
	}
	if isCSR {
		meta := extractMeta(csrMetaParam)
		//
		var oidEmailAddress = asn1.ObjectIdentifier{1, 2, 840, 113549, 1, 9, 1}
		subj := pkix.Name{
			CommonName:         meta.CommonName,
			Country:            []string{meta.Country},
			Province:           []string{meta.Province},
			Locality:           []string{meta.Locality},
			Organization:       []string{meta.Organization},
			OrganizationalUnit: []string{meta.OrganizationUnit},
			ExtraNames: []pkix.AttributeTypeAndValue{
				{
					Type: oidEmailAddress,
					Value: asn1.RawValue{
						Tag:   asn1.TagIA5String,
						Bytes: []byte(meta.Email),
					},
				},
			},
		}
		template := x509.CertificateRequest{
			Subject:            subj,
			SignatureAlgorithm: x509.SHA256WithRSA,
		}
		csrBytes, _ := x509.CreateCertificateRequest(rand.Reader, &template, keyObj)
		certBytes := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE REQUEST", Bytes: csrBytes})
		//
		pair.CSRCert = string(certBytes)
	} else {
		//
		pub, err := ssh.NewPublicKey(publickeyObj)
		if err != nil {
			return pair, err
		}
		pubKeyBytes := ssh.MarshalAuthorizedKey(pub)
		pair.PublicKey = string(pubKeyBytes)
	}

	return pair, nil
}

func extractMeta(param map[string]string) CSRMeta {
	m := CSRMeta{
		Email:            "mail@example.com",
		CommonName:       "example.com",
		Country:          "Country",
		Province:         "Province",
		Locality:         "City",
		Organization:     "Company Ltd",
		OrganizationUnit: "IT",
	}
	if _, exist := param["email"]; exist && param["email"] != "" {
		m.Email = param["email"]
	}
	if _, exist := param["common_name"]; exist && param["common_name"] != "" {
		m.CommonName = param["common_name"]
	}
	if _, exist := param["country"]; exist && param["country"] != "" {
		m.Country = param["country"]
	}
	if _, exist := param["province"]; exist && param["province"] != "" {
		m.Province = param["province"]
	}
	if _, exist := param["locality"]; exist && param["locality"] != "" {
		m.Locality = param["locality"]
	}
	if _, exist := param["organization"]; exist && param["organization"] != "" {
		m.Organization = param["organization"]
	}
	if _, exist := param["organization_unit"]; exist && param["organization_unit"] != "" {
		m.OrganizationUnit = param["organization_unit"]
	}

	return m
}
