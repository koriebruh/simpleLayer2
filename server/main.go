package main

import (
	"context"
	"crypto/ecdsa"
	"encoding/json"
	"fmt"
	"github.com/cockroachdb/errors/grpc/status"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/log"
	pb "github.com/koriebruh/simpleLayer2/proto/layer2"
	"google.golang.org/grpc/codes"
	"math/big"
	"os"
	"sync"
)

type Layer2Handler struct {
	pb.UnimplementedLayer2ServiceServer
	mu          sync.Mutex
	transaction []*pb.TransactionRequest
}

func (l *Layer2Handler) SubmitTransaction(ctx context.Context, req *pb.TransactionRequest) (*pb.TransactionResponse, error) {
	log.Info(fmt.Sprintf("Incoming Request %v", req))

	// logic check before we save in temp storage
	client, err := ethclient.DialContext(ctx, "https://mainnet.infura.io/v3/7108f6b019944d2082df7b667e6b1f4a")
	if err != nil {
		return &pb.TransactionResponse{
			TransactionId: req.TransactionId,
			Status:        "failed",
			Message:       "Internal Server Error",
		}, err
	}

	pubKey, err := crypto.UnmarshalPubkey([]byte(req.Sender))
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "Invalid sender public key: %v", err)
	}

	if err = filter(ctx, client, *pubKey, big.NewInt(req.Amount)); err != nil {
		return nil, err
	}

	if err = addTxPool(req); err != nil {
		return nil, err
	}

	return &pb.TransactionResponse{
		TransactionId: req.TransactionId,
		Status:        "success",
		Message:       "Transaction processed successfully",
	}, nil
}

func (l *Layer2Handler) MonitorBatchStatus(req *pb.BatchStatusRequest, stream pb.Layer2Service_MonitorBatchStatusServer) error {
	//TODO implement me
	panic("implement me")
}

func (l *Layer2Handler) TriggerBatchProcessing(ctx context.Context, req *pb.BatchProcessingRequest) (*pb.BatchProcessingResponse, error) {
	//TODO implement me
	panic("implement me")
}

// filter checking insert User
func filter(ctx context.Context, c *ethclient.Client, sender ecdsa.PublicKey, valueSend *big.Int) error {
	// checking amount sender
	address := crypto.PubkeyToAddress(sender)
	balanceRN, err := c.BalanceAt(ctx, address, nil)
	if err != nil {
		return status.Errorf(codes.Internal, "Failed to retrieve balance: %v", err)
	}

	// validate
	if balanceRN.Cmp(valueSend) < 0 {
		return status.Errorf(codes.InvalidArgument, "Insufficient funds for transaction: %v", err)
	}

	return nil
	//// DO CHECK MORE
	//// DO CHECK MORE
}

// addTxPool save transaction  the corresponding transaction to the transaction_pool.txt file
func addTxPool(req *pb.TransactionRequest) error {
	file, err := os.OpenFile("store/transaction_pool.txt", os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		log.Info("Error OpenFile")
	}

	m, err := json.Marshal(req)
	if err != nil {
		return status.Errorf(codes.Internal, "Failed to retrieve balance: %v", err)
	}

	_, err = file.WriteString(string(m))
	if err != nil {
		return status.Errorf(codes.Internal, "Failed to retrieve balance: %v", err)
	}

	return nil
}

func main() {

}
