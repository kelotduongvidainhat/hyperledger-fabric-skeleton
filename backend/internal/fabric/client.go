package fabric

import (
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"time"

	"github.com/hyperledger/fabric-gateway/pkg/client"
	"github.com/hyperledger/fabric-gateway/pkg/identity"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

const (
	mspID         = "Org1MSP"
	cryptoPath    = "/home/qwe/hyperledger-fabric-skeleton/network/crypto-config/peerOrganizations/org1.example.com"
	certPath      = cryptoPath + "/users/Admin@org1.example.com/msp/signcerts/cert.pem"
	keyPath       = cryptoPath + "/users/Admin@org1.example.com/msp/keystore"
	tlsCertPath   = cryptoPath + "/peers/peer0.org1.example.com/tls/ca.crt"
	peerEndpoint  = "localhost:7051"
	gatewayPeer   = "peer0.org1.example.com"
	channelName   = "mychannel"
	chaincodeName = "asset-transfer"
)

// FabricClient holds the gateway connection
type FabricClient struct {
	Connection *grpc.ClientConn
	Gateway    *client.Gateway
	Network    *client.Network
	Contract   *client.Contract
}

// NewFabricClient initializes a connection to the Fabric network
func NewFabricClient() (*FabricClient, error) {
	// The gRPC client connection should be shared by all Gateway connections to this endpoint
	clientConnection, err := newGrpcConnection()
	if err != nil {
		return nil, fmt.Errorf("failed to create gRPC connection: %w", err)
	}

	id, err := newIdentity()
	if err != nil {
		return nil, fmt.Errorf("failed to create identity: %w", err)
	}

	sign, err := newSign()
	if err != nil {
		return nil, fmt.Errorf("failed to create signature: %w", err)
	}

	// Create a Gateway connection for this client identity
	gw, err := client.Connect(
		id,
		client.WithSign(sign),
		client.WithClientConnection(clientConnection),
		// Default timeouts for different gRPC calls
		client.WithEvaluateTimeout(5*time.Second),
		client.WithEndorseTimeout(15*time.Second),
		client.WithSubmitTimeout(5*time.Second),
		client.WithCommitStatusTimeout(1*time.Minute),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to gateway: %w", err)
	}

	network := gw.GetNetwork(channelName)
	contract := network.GetContract(chaincodeName)

	return &FabricClient{
		Connection: clientConnection,
		Gateway:    gw,
		Network:    network,
		Contract:   contract,
	}, nil
}

// newGrpcConnection creates a gRPC connection to the Gateway server.
func newGrpcConnection() (*grpc.ClientConn, error) {
	certificate, err := loadCertificate(tlsCertPath)
	if err != nil {
		return nil, err
	}

	certPool := x509.NewCertPool()
	certPool.AddCert(certificate)
	transportCredentials := credentials.NewClientTLSFromCert(certPool, gatewayPeer)

	connection, err := grpc.Dial(peerEndpoint, grpc.WithTransportCredentials(transportCredentials))
	if err != nil {
		return nil, fmt.Errorf("failed to create gRPC connection: %w", err)
	}

	return connection, nil
}

// newIdentity creates a client identity for this Gateway connection using an X.509 certificate.
func newIdentity() (*identity.X509Identity, error) {
	certificate, err := loadCertificate(certPath)
	if err != nil {
		return nil, err
	}

	id, err := identity.NewX509Identity(mspID, certificate)
	if err != nil {
		return nil, err
	}

	return id, nil
}

func loadCertificate(filename string) (*x509.Certificate, error) {
	certificatePEM, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read certificate file: %w", err)
	}
	return identity.CertificateFromPEM(certificatePEM)
}

// newSign creates a function that generates a digital signature from a message digest using a private key.
func newSign() (identity.Sign, error) {
	files, err := ioutil.ReadDir(keyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read private key directory: %w", err)
	}
	if len(files) == 0 {
		return nil, fmt.Errorf("no private key file found in %s", keyPath)
	}

	path := filepath.Join(keyPath, files[0].Name())
	privateKeyPEM, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read private key file: %w", err)
	}

	privateKey, err := identity.PrivateKeyFromPEM(privateKeyPEM)
	if err != nil {
		return nil, err
	}

	sign, err := identity.NewPrivateKeySign(privateKey)
	if err != nil {
		return nil, err
	}

	return sign, nil
}

// Close closes the gateway connection
func (f *FabricClient) Close() {
	if f.Gateway != nil {
		f.Gateway.Close()
	}
	if f.Connection != nil {
		f.Connection.Close()
	}
}
