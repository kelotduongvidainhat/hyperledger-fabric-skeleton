package fabric

import (
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"time"

	"github.com/hyperledger/fabric-gateway/pkg/client"
	"github.com/hyperledger/fabric-gateway/pkg/identity"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

const (
	mspID         = "Org1MSP"
	cryptoPath    = "/home/qwe/hyperledger-fabric-skeleton/network/crypto-config/peerOrganizations/org1.example.com"
	tlsCertPath   = cryptoPath + "/peers/peer0.org1.example.com/tls/ca.crt"
	peerEndpoint  = "localhost:7051"
	gatewayPeer   = "peer0.org1.example.com"
	channelName   = "mychannel"
	chaincodeName = "asset-transfer"
)

// FabricClient manages the connection to the Fabric network
type FabricClient struct {
	Connection *grpc.ClientConn
	Store      *EnrollmentStore
}

// NewFabricClient initializes the gRPC connection and enrollment store
func NewFabricClient() (*FabricClient, error) {
	clientConnection, err := newGrpcConnection()
	if err != nil {
		return nil, fmt.Errorf("failed to create gRPC connection: %w", err)
	}

	// Initialize store pointing to crypto-config for now (User1/Admin simulation)
	// In production, this would point to a secure wallet directory
	store := NewEnrollmentStore(cryptoPath + "/users")

	return &FabricClient{
		Connection: clientConnection,
		Store:      store,
	}, nil
}

// executeAction creates a gateway connection for the user and runs the action
func (f *FabricClient) executeAction(userID string, action func(*client.Contract) (interface{}, error)) (interface{}, error) {
	// 1. Load Identity
	// For simulation, we map userID "admin" -> Admin@org1.example.com
	// and "user1" -> User1@org1.example.com
	// If userID is empty, default to Admin
	var label string
	if userID == "user1" {
		label = "user2@org1.example.com"
	} else {
		label = "Admin@org1.example.com"
	}

	id, err := f.Store.GetIdentity(label, mspID)
	if err != nil {
		return nil, fmt.Errorf("failed to load identity for %s: %w", userID, err)
	}

	sign, err := f.Store.GetSigner(label)
	if err != nil {
		return nil, fmt.Errorf("failed to load signer for %s: %w", userID, err)
	}

	// 2. Connect to Gateway
	gw, err := client.Connect(
		id,
		client.WithSign(sign),
		client.WithClientConnection(f.Connection),
		client.WithEvaluateTimeout(5*time.Second),
		client.WithEndorseTimeout(15*time.Second),
		client.WithSubmitTimeout(5*time.Second),
		client.WithCommitStatusTimeout(1*time.Minute),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to gateway: %w", err)
	}
	defer gw.Close()

	// 3. Get Contract
	network := gw.GetNetwork(channelName)
	contract := network.GetContract(chaincodeName)

	// 4. Run Action
	return action(contract)
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

func loadCertificate(filename string) (*x509.Certificate, error) {
	certificatePEM, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read certificate file: %w", err)
	}
	return identity.CertificateFromPEM(certificatePEM)
}

// Close closes the gRPC connection
func (f *FabricClient) Close() {
	if f.Connection != nil {
		f.Connection.Close()
	}
}
