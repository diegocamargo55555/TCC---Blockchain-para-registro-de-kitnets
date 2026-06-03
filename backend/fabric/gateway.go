package fabric

import (
	"crypto/x509"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"time"

	"github.com/hyperledger/fabric-gateway/pkg/client"
	"github.com/hyperledger/fabric-gateway/pkg/identity"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

const (
	channelName   = "cartoriochannel"
	chaincodeName = "cartorio"
	mspID         = "Org1MSP"
	peerEndpoint  = "localhost:7051"
	gatewayPeer   = "peer0.org1.example.com"
)

var (
	cryptoPath = filepath.Join("..", "fabric-samples", "test-network", "organizations", "peerOrganizations", "org1.example.com")
	certPath   = filepath.Join(cryptoPath, "users", "User1@org1.example.com", "msp", "signcerts", "User1@org1.example.com-cert.pem")
	keyPath    = filepath.Join(cryptoPath, "users", "User1@org1.example.com", "msp", "keystore")
	tlsCertPath = filepath.Join(cryptoPath, "peers", gatewayPeer, "tls", "ca.crt")
)

type GatewayConnection struct {
	Gateway  *client.Gateway
	Contract *client.Contract
}

func Connect() (*GatewayConnection, error) {
	clientConnection, err := newGrpcConnection()
	if err != nil {
		return nil, fmt.Errorf("falha ao criar conexao gRPC: %w", err)
	}

	id, err := newIdentity()
	if err != nil {
		return nil, fmt.Errorf("falha ao criar identidade: %w", err)
	}

	sign, err := newSign()
	if err != nil {
		return nil, fmt.Errorf("falha ao criar assinatura: %w", err)
	}

	gateway, err := client.Connect(
		id,
		client.WithSign(sign),
		client.WithClientConnection(clientConnection),
		client.WithEvaluateTimeout(5*time.Second),
		client.WithEndorseTimeout(15*time.Second),
		client.WithSubmitTimeout(5*time.Second),
		client.WithCommitStatusTimeout(1*time.Minute),
	)
	if err != nil {
		return nil, fmt.Errorf("falha ao conectar gateway: %w", err)
	}

	network := gateway.GetNetwork(channelName)
	contract := network.GetContract(chaincodeName)

	return &GatewayConnection{
		Gateway:  gateway,
		Contract: contract,
	}, nil
}

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
		return nil, fmt.Errorf("falha grpc dial: %w", err)
	}

	return connection, nil
}

func newIdentity() (identity.Identity, error) {
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

func newSign() (identity.Sign, error) {
	files, err := os.ReadDir(keyPath)
	if err != nil {
		return nil, fmt.Errorf("falha ao ler keystore dir: %w", err)
	}

	if len(files) == 0 {
		return nil, fmt.Errorf("keystore folder vazio")
	}

	privateKeyPath := path.Join(keyPath, files[0].Name())
	privateKeyBytes, err := os.ReadFile(privateKeyPath)
	if err != nil {
		return nil, fmt.Errorf("falha ao ler private key: %w", err)
	}

	privateKey, err := identity.PrivateKeyFromPEM(privateKeyBytes)
	if err != nil {
		return nil, fmt.Errorf("falha ao decodificar private key: %w", err)
	}

	sign, err := identity.NewPrivateKeySign(privateKey)
	if err != nil {
		return nil, err
	}

	return sign, nil
}

func loadCertificate(filename string) (*x509.Certificate, error) {
	certificatePEM, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("falha ao ler certificado: %w", err)
	}
	return identity.CertificateFromPEM(certificatePEM)
}
