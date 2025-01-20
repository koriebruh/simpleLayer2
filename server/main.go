package main

import (
	"bufio"
	"context"
	"crypto/ecdsa"
	"encoding/json"
	"fmt"
	"github.com/cockroachdb/errors/grpc/status"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
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

	// do read file

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

// batchInsert for insert
// private key for signed by the system to represent the sender, and open batch insert to blockchain
func batchInsert(ctx context.Context, client *ethclient.Client, privateKeyHex string) error {
	file, err := os.Open("store/transaction_pool.txt")
	if err != nil {
		return fmt.Errorf("error opening transaction pool file: %v", err)
	}
	defer file.Close()

	// Extract privateKey System
	privateKey, err := crypto.HexToECDSA(privateKeyHex)
	if err != nil {
		return fmt.Errorf("invalid private key: %v", err)
	}

	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return fmt.Errorf("invalid public key")
	}

	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
	nonce, err := client.PendingNonceAt(ctx, fromAddress)
	if err != nil {
		return fmt.Errorf("failed to get nonce: %v", err)
	}

	chainID, err := client.ChainID(ctx)
	if err != nil {
		return fmt.Errorf("failed to get chain id: %v", err)
	}

	// Get suggested gas tip cap (miner tip)
	gasTipCap, err := client.SuggestGasTipCap(ctx)
	if err != nil {
		return fmt.Errorf("failed to get gas tip cap: %v", err)
	}

	// Get suggested
	gasPrice, err := client.SuggestGasPrice(ctx)
	if err != nil {
		return fmt.Errorf("failed to get gas fee cap: %v", err)
	}

	// Calculate fee cap: max(2 * tipCap, gasPrice)
	gasFeeCap := new(big.Int).Mul(gasTipCap, big.NewInt(2))
	if gasFeeCap.Cmp(gasPrice) < 0 {
		gasFeeCap = gasPrice
	}

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()

		var tx pb.TransactionRequest
		if err := json.Unmarshal([]byte(line), &tx); err != nil {
			fmt.Println("Error unmarshaling JSON:", err)
			continue
		}

		amount := big.NewInt(tx.Amount)
		toAddress := common.HexToAddress(tx.Receiver)
		txData := &types.DynamicFeeTx{
			ChainID:   chainID,
			Nonce:     nonce,
			To:        &toAddress,
			Value:     amount,
			Gas:       21000,
			GasTipCap: gasTipCap, // max priority fee per gas
			GasFeeCap: gasFeeCap, // max fee per gas
		}

		transaction := types.NewTx(txData)
		signer := types.NewLondonSigner(chainID)

		// Sign the transaction
		signedTx, err := types.SignTx(transaction, signer, privateKey)
		if err != nil {
			fmt.Printf("Failed to sign transaction: %v\n", err)
			continue
		}

		// Send the signed transaction
		err = client.SendTransaction(ctx, signedTx)
		if err != nil {
			fmt.Printf("Failed to send transaction: %v\n", err)
			continue
		}

		fmt.Printf("Transaction sent: ID=%s, Hash=%s\n", tx.TransactionId, signedTx.Hash().Hex())
		nonce++
	}

	return nil
}

func main() {

}
