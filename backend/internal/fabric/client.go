package fabric

import (
	"crypto/x509"
	"fmt"
	"io/ioutil"


	"github.com/hyperledger/fabric-gateway/pkg/client"
	"github.com/hyperledger/fabric-gateway/pkg/identity"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// Config holds connection details
type Config struct {
	CertPath      string
	KeyPath       string
	TlsCertPath   string
	PeerEndpoint  string
	GatewayPeer   string
	ChannelName   string
	ChaincodeName string
}

// CreateGRPCConnection creates a gRPC connection to the peer
func CreateGRPCConnection(cfg Config) (*grpc.ClientConn, error) {
	certPool := x509.NewCertPool()
	tlsCert, err := ioutil.ReadFile(cfg.TlsCertPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read TLS cert: %w", err)
	}
	if !certPool.AppendCertsFromPEM(tlsCert) {
		return nil, fmt.Errorf("failed to append TLS cert")
	}

	transportCreds := credentials.NewClientTLSFromCert(certPool, "")

	conn, err := grpc.Dial(cfg.PeerEndpoint, grpc.WithTransportCredentials(transportCreds))
	if err != nil {
		return nil, fmt.Errorf("failed to create gRPC connection: %w", err)
	}

	return conn, nil
}

// CreateGateway creates a Gateway instance for a specific user identity
func CreateGateway(conn *grpc.ClientConn, id *identity.X509Identity, sign identity.Sign) (*client.Gateway, error) {
	gateway, err := client.Connect(
		id,
		client.WithSign(sign),
		client.WithClientConnection(conn),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to gateway: %w", err)
	}

	return gateway, nil
}

// SetupConnection (Legacy/Admin) - wraps the above for initial Admin setup
func SetupConnection(cfg Config) (*grpc.ClientConn, *client.Gateway, error) {
	conn, err := CreateGRPCConnection(cfg)
	if err != nil {
		return nil, nil, err
	}

	// Load Admin Identity
	certificate, err := loadCertificate(cfg.CertPath)
	if err != nil {
		return nil, nil, err
	}
	id, err := identity.NewX509Identity("Org1MSP", certificate)
	if err != nil {
		return nil, nil, err
	}
	signer, err := loadPrivateKey(cfg.KeyPath)
	if err != nil {
		return nil, nil, err
	}

	gw, err := CreateGateway(conn, id, signer)
	if err != nil {
		return nil, nil, err
	}

	return conn, gw, nil
}

func loadCertificate(filename string) (*x509.Certificate, error) {
	certificatePEM, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read certificate file: %w", err)
	}
	return identity.CertificateFromPEM(certificatePEM)
}

func loadPrivateKey(filename string) (identity.Sign, error) {
	privateKeyPEM, err := ioutil.ReadFile(filename)
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
