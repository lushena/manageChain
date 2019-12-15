package sdk

import (
	"crypto"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/hyperledger/fabric/bccsp"
	"github.com/hyperledger/fabric/bccsp/factory"
	"github.com/hyperledger/fabric/bccsp/signer"
	"github.com/hyperledger/fabric/common/tools/cryptogen/ca"
	"github.com/hyperledger/fabric/common/tools/cryptogen/msp"
	"github.com/pkg/errors"
)

const (
	defaultCountry  = "CN"
	defaultProvince = "Zhejiang"
	defaultLocality = "Hangzhou"
	defaultAddress  = "company"
	defaultCode     = "310000"
	defaultUnit     = "" // must be empty

	enableNodeOUs  = false // must be false if you want use the same admincert in peerorg and ordererorg
	caFold         = "ca"
	tlscaFold      = "tlsca"
	peersFold      = "peers"
	orderersFold   = "orderers"
	usersFold      = "users"
	mspFold        = "msp"
	tlsFold        = "tls"
	adminBaseName  = "Admin"
	admincertsFold = "admincerts"
	cacertsFold    = "cacerts"
	tlscertsFold   = "tlscacerts"
)

// NodeType represents the type of node
type NodeType int

const (
	// PeerNode is the type of peer's certificate
	PeerNode NodeType = iota
	// OrdererNode is the type of orderer's certificate
	OrdererNode
)

// CA ...
type CA struct {
	ca      *ca.CA
	tlsca   *ca.CA
	baseDir string
	orgName string
}

type cafiles struct {
	MSPID      string
	Org        string
	AdminCerts map[string][]byte
	CACerts    map[string][]byte
	TLSCACerts map[string][]byte
}

func readFiles(dir string) (map[string][]byte, error) {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	fileMap := make(map[string][]byte)
	for _, file := range files {
		if !file.IsDir() {
			content, err := ioutil.ReadFile(path.Join(dir, file.Name()))
			if err != nil {
				return nil, err
			}
			fileMap[file.Name()] = content
		}

	}
	return fileMap, nil
}

func writeFiles(dir string, files map[string][]byte) error {
	_, err := os.Stat(dir)
	if err == nil {
		return errors.Errorf("directory [%s] already exists", dir)
	}
	if !os.IsNotExist(err) {
		return err
	}
	if err = os.MkdirAll(dir, os.ModePerm); err != nil {
		return err

	}
	for name, content := range files {
		if err = ioutil.WriteFile(path.Join(dir, name), content, os.ModePerm); err != nil {
			return err
		}
	}

	return nil
}

// AdminMSPDir ...
func (ca *CA) AdminMSPDir() string {
	adminCommonName := fmt.Sprintf("%s@%s", adminBaseName, ca.orgName)
	return path.Join(ca.baseDir, usersFold, adminCommonName, mspFold)
}

// MSPDir ...
func (ca *CA) MSPDir() string {
	return path.Join(ca.baseDir, mspFold)
}

// AdminCommonName ...
func (ca *CA) AdminCommonName() string {
	return fmt.Sprintf("%s@%s", adminBaseName, ca.orgName)
}

// MSPBytes returns the bytes of msp folder
func (ca *CA) MSPBytes(mspID string) ([]byte, error) {
	bundle := &cafiles{}
	bundle.Org = ca.orgName
	bundle.MSPID = mspID
	var err error
	bundle.AdminCerts, err = readFiles(path.Join(ca.baseDir, mspFold, admincertsFold))
	if err != nil {
		logger.Error("Error reading admincerts", err)
		return nil, err
	}

	bundle.CACerts, err = readFiles(path.Join(ca.baseDir, mspFold, cacertsFold))
	if err != nil {
		logger.Error("Error reading cacerts", err)
		return nil, err
	}

	bundle.TLSCACerts, err = readFiles(path.Join(ca.baseDir, mspFold, tlscertsFold))
	if err != nil {
		logger.Error("Error reading tlscacerts", err)
		return nil, err
	}

	return json.Marshal(bundle)
}

// WriteMSPDir read msp bytes and writes msp certs into directory,
// and returns the mspPath and mspID
func WriteMSPDir(baseDir string, data []byte) (string, string, error) {
	bundle := &cafiles{}
	err := json.Unmarshal(data, bundle)
	if err != nil {
		logger.Error("Error unmarshaling data to cafiles", err)
		return "", "", err
	}

	dir := path.Join(baseDir, bundle.MSPID)

	_, err = os.Stat(dir)
	if err == nil {
		return "", "", errors.Errorf("directory [%s] already exists", dir)
	}

	if !os.IsNotExist(err) {
		return "", "", err
	}

	if err = os.MkdirAll(dir, os.ModePerm); err != nil {
		return "", "", err
	}

	if err = writeFiles(path.Join(dir, admincertsFold), bundle.AdminCerts); err != nil {
		logger.Error("Error writing admincerts", err)
		return "", "", err
	}

	if err = writeFiles(path.Join(dir, cacertsFold), bundle.CACerts); err != nil {
		logger.Error("Error writing cacerts", err)
		return "", "", err
	}

	if err = writeFiles(path.Join(dir, tlscertsFold), bundle.TLSCACerts); err != nil {
		logger.Error("Error writing tlscacerts", err)
		return "", "", err
	}

	return dir, bundle.MSPID, nil

}

// TLSCACert ...
func (ca *CA) TLSCACert() []byte {
	return pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: ca.tlsca.SignCert.Raw,
	})
}

// CertConfig ...
type CertConfig struct {
	CN       string
	SAN      []string
	NodeType NodeType
}

// NodeMSPDir ...
func (ca *CA) NodeMSPDir(cn string, nodeType NodeType) string {
	var nodeDir string
	switch nodeType {
	case PeerNode:
		nodeDir = peersFold
	case OrdererNode:
		nodeDir = orderersFold
	}
	return path.Join(ca.baseDir, nodeDir, cn, mspFold)

}

// NodeTLSDir ...
func (ca *CA) NodeTLSDir(cn string, nodeType NodeType) string {
	var nodeDir string
	switch nodeType {
	case PeerNode:
		nodeDir = peersFold
	case OrdererNode:
		nodeDir = orderersFold
	}
	return path.Join(ca.baseDir, nodeDir, cn, tlsFold)
}

// GenerateMSP ...
func (ca *CA) GenerateMSP(nodes []*CertConfig, users []string) error {

	for _, node := range nodes {
		var nodeType int
		var nodeDir string
		switch node.NodeType {
		case PeerNode:
			nodeType = msp.PEER
			nodeDir = peersFold
		case OrdererNode:
			nodeType = msp.ORDERER
			nodeDir = orderersFold
		}
		nodeBaseDir := path.Join(ca.baseDir, nodeDir)

		err := ca.generateMSP(nodeBaseDir, node.CN, node.SAN, nodeType)
		if err != nil {
			logger.Errorf("Error generating msp for %s: %s", node.CN, err)
			return err
		}
	}

	userBaseDir := path.Join(ca.baseDir, usersFold)
	for _, user := range users {
		cn := fmt.Sprintf("%s@%s", user, ca.orgName)
		err := ca.generateMSP(userBaseDir, user, nil, msp.CLIENT)
		if err != nil {
			logger.Errorf("Error generating msp for %s: %s", cn, err)
			return err
		}
	}

	return nil
}

func (ca *CA) generateMSP(baseDir string, commonName string, san []string, nodeType int) error {

	mspDir := path.Join(baseDir, commonName)
	if _, err := os.Stat(mspDir); os.IsNotExist(err) {
		err := msp.GenerateLocalMSP(mspDir, commonName, san, ca.ca, ca.tlsca, nodeType, enableNodeOUs)
		if err != nil {
			logger.Errorf("Error generating local MSP for %s:\n%v\n", commonName, err)
			return err
		}
		adminCommonName := fmt.Sprintf("%s@%s", adminBaseName, ca.orgName)
		if adminCommonName != commonName {
			// copy admin cert
			userDir := path.Join(ca.baseDir, usersFold)
			adminCertDir := path.Join(mspDir, mspFold, admincertsFold)
			err = copyAdminCert(userDir, adminCertDir, adminCommonName)
			if err != nil {
				logger.Error("Error copying admin cert", err)
				return err
			}
		}

		return nil

	}

	logger.Infof("%s already existes, skip", commonName)
	return nil
}

func (ca *CA) prepare() error {
	// generate msp

	err := msp.GenerateVerifyingMSP(path.Join(ca.baseDir, mspFold), ca.ca, ca.tlsca, enableNodeOUs)
	if err != nil {
		logger.Error("Error generating verifying msp", err)
		return err
	}

	// generate admin
	userDir := path.Join(ca.baseDir, usersFold)
	commonName := fmt.Sprintf("%s@%s", adminBaseName, ca.orgName)
	err = ca.generateMSP(userDir, commonName, nil, msp.CLIENT)
	if err != nil {
		logger.Errorf("Error generating msp for %s: %s", commonName, err)
		return err
	}

	// copy admin to msp
	adminDir := path.Join(ca.baseDir, mspFold, admincertsFold)
	err = copyAdminCert(userDir, adminDir, commonName)
	if err != nil {
		logger.Error("Error copying admin cert", err)
		return err
	}

	return nil
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, in)
	cerr := out.Close()
	if err != nil {
		return err
	}
	return cerr
}

func copyAdminCert(usersDir, adminCertsDir, adminUserName string) error {
	if _, err := os.Stat(filepath.Join(adminCertsDir,
		adminUserName+"-cert.pem")); err == nil {
		return nil
	}
	// delete the contents of admincerts
	err := os.RemoveAll(adminCertsDir)
	if err != nil {
		return err
	}
	// recreate the admincerts directory
	err = os.MkdirAll(adminCertsDir, 0755)
	if err != nil {
		return err
	}
	err = copyFile(filepath.Join(usersDir, adminUserName, "msp", "signcerts",
		adminUserName+"-cert.pem"), filepath.Join(adminCertsDir,
		adminUserName+"-cert.pem"))
	if err != nil {
		return err
	}
	return nil

}

// NewCA ...
// Create new CA in mspDir
func NewCA(mspDir string, orgName string) (*CA, error) {
	commonName := orgName
	ca, err := newCA(path.Join(mspDir, caFold), orgName, commonName)
	if err != nil {
		return nil, err
	}

	tlsca, err := newCA(path.Join(mspDir, tlscaFold), orgName, commonName)
	if err != nil {
		return nil, err
	}

	newCA := &CA{
		ca:      ca,
		tlsca:   tlsca,
		baseDir: mspDir,
		orgName: orgName,
	}

	// prepare for msp and admin certs
	err = newCA.prepare()
	if err != nil {
		logger.Error("Error preparing", err)
		return nil, err
	}

	return newCA, nil
}

// ConstructCAFromDir ...
func ConstructCAFromDir(mspDir string) (*CA, error) {
	ca, err := constructCAFromDir(path.Join(mspDir, caFold))
	if err != nil {
		return nil, err
	}

	tlsca, err := constructCAFromDir(path.Join(mspDir, tlscaFold))
	if err != nil {
		return nil, err
	}
	return &CA{
		ca:      ca,
		tlsca:   tlsca,
		baseDir: mspDir,
		orgName: ca.Name,
	}, nil
}

// Create a new one in baseDir
func newCA(baseDir, orgName, commonName string) (*ca.CA, error) {
	return ca.NewCA(baseDir, orgName, commonName, defaultCountry, defaultProvince, defaultLocality, defaultUnit, defaultAddress, defaultCode)
}

// Constructed from existing files or
func constructCAFromDir(baseDir string) (*ca.CA, error) {
	fs, err := ioutil.ReadDir(baseDir)
	if err != nil {
		logger.Errorf("Error reading dir %s: %s", baseDir, err)
		return nil, err
	}
	if len(fs) == 0 {
		return nil, errors.New("No files found in dir")
	}

	// construct from files
	signer, err := getSignerFromKeystore(baseDir)
	if err != nil {
		logger.Error("Error getting signer from keystore", err)
		return nil, err
	}
	cert, err := getCertFromDir(baseDir)
	if err != nil {
		logger.Error("Error getting certificate from dir", err)
		return nil, err
	}
	country, province, locality, unit, address, code := retriveInfoFromCert(cert)

	return &ca.CA{
		Name:               cert.Subject.CommonName,
		Country:            country,
		Province:           province,
		Locality:           locality,
		OrganizationalUnit: unit,
		StreetAddress:      address,
		PostalCode:         code,
		SignCert:           cert,
		Signer:             signer,
	}, nil

}

func retriveInfoFromCert(cert *x509.Certificate) (country, province, locality, unit, address, code string) {
	if len(cert.Subject.Country) > 0 {
		country = cert.Subject.Country[0]
	}
	if len(cert.Subject.Province) > 0 {
		province = cert.Subject.Province[0]
	}
	if len(cert.Subject.Locality) > 0 {
		locality = cert.Subject.Locality[0]
	}
	if len(cert.Subject.OrganizationalUnit) > 0 {
		unit = cert.Subject.OrganizationalUnit[0]
	}
	if len(cert.Subject.StreetAddress) > 0 {
		address = cert.Subject.StreetAddress[0]
	}
	if len(cert.Subject.PostalCode) > 0 {
		code = cert.Subject.PostalCode[0]
	}
	return
}

// AdminCert ...
func (ca *CA) AdminCert() ([]byte, error) {
	files, err := readFiles(path.Join(ca.baseDir, mspFold, admincertsFold))
	if err != nil {
		return nil, err
	}
	for _, v := range files {
		return v, nil
	}
	return nil, errors.New("no admin cert can be found")
}

// RootCert ...
func (ca *CA) RootCert() ([]byte, error) {
	files, err := readFiles(path.Join(ca.baseDir, mspFold, cacertsFold))
	if err != nil {
		return nil, err
	}
	for _, v := range files {
		return v, nil
	}
	return nil, errors.New("no root cert can be found")

}

func getCertFromDir(baseDir string) (*x509.Certificate, error) {
	fs, err := ioutil.ReadDir(baseDir)
	if err != nil {
		logger.Errorf("Error reading dir %s: %s", baseDir, err)
		return nil, err
	}
	for _, f := range fs {
		if !f.IsDir() {
			if strings.HasSuffix(f.Name(), "-cert.pem") {
				bytes, err := ioutil.ReadFile(path.Join(baseDir, f.Name()))
				if err != nil {
					logger.Error("Error reading cert from file", err)
					return nil, err
				}
				block, _ := pem.Decode(bytes)
				if block.Type != "CERTIFICATE" {
					return nil, errors.New("Wrong block type, expected CERTIFICATE, got " + block.Type)
				}
				return x509.ParseCertificate(block.Bytes)
			}
		}

	}

	return nil, errors.New("No certificate found in dir")
}

func getSignerFromKeystore(keystorePath string) (crypto.Signer, error) {
	opts := &factory.FactoryOpts{
		ProviderName: "SW",
		SwOpts: &factory.SwOpts{
			HashFamily: "SHA2",
			SecLevel:   256,

			FileKeystore: &factory.FileKeystoreOpts{
				KeyStorePath: keystorePath,
			},
		},
	}

	csp, err := factory.GetBCCSPFromOpts(opts)
	if err != nil {
		logger.Error("Error getting bccsp from opts", err)
		return nil, err
	}

	files, err := ioutil.ReadDir(keystorePath)
	if err != nil {
		logger.Errorf("Error reading dir %s: %s", keystorePath, err)
		return nil, err
	}

	for _, file := range files {
		if !file.IsDir() {
			if strings.HasSuffix(file.Name(), "_sk") {
				rawKey, err := ioutil.ReadFile(path.Join(keystorePath, file.Name()))
				if err != nil {
					return nil, err
				}

				block, _ := pem.Decode(rawKey)
				priv, err := csp.KeyImport(block.Bytes, &bccsp.ECDSAPrivateKeyImportOpts{Temporary: true})
				if err != nil {
					logger.Error("Error importing key", err)
					return nil, err
				}
				return signer.New(csp, priv)
			}

		}
	}
	return nil, errors.New("No key found in keystore")
}
