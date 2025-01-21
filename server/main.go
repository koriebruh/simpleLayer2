package main

import (
	"bufio"
	"context"
	"crypto/ecdsa"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/cockroachdb/errors/grpc/status"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/log"
	pb "github.com/koriebruh/simpleLayer2/proto/layer2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	log2 "log"
	"math/big"
	"net"
	"os"
	"strings"
	"sync"
)

type Layer2Handler struct {
	pb.UnimplementedLayer2ServiceServer
	mu          sync.Mutex
	transaction []*pb.TransactionRequest
	*ethclient.Client
}

func NewLayer2Handler(unimplementedLayer2ServiceServer pb.UnimplementedLayer2ServiceServer, mu sync.Mutex, transaction []*pb.TransactionRequest, client *ethclient.Client) *Layer2Handler {
	return &Layer2Handler{UnimplementedLayer2ServiceServer: unimplementedLayer2ServiceServer, mu: mu, transaction: transaction, Client: client}
}

func (l *Layer2Handler) SubmitTransaction(ctx context.Context, req *pb.TransactionRequest) (*pb.TransactionResponse, error) {
	log.Info(fmt.Sprintf("Incoming Request %v", req))
	client := l.Client

	publicKeyHex := strings.TrimPrefix(req.Sender, "0x")
	publicKeyBytes, err := hex.DecodeString(publicKeyHex)
	if err != nil {
		return nil, fmt.Errorf("failed to decode public key: %v", err)
	}

	// Mengonversi ke public key ECDSA
	pubKey, err := crypto.UnmarshalPubkey(publicKeyBytes)
	if err != nil {
		log2.Fatalf("Error unmarshalling public key: %v", err)
		return nil, err
	}

	amount := new(big.Int).SetInt64(req.Amount)
	if err := filter(ctx, client, *pubKey, amount); err != nil {
		return nil, err
	}

	if err := addTxPool(req); err != nil {
		return nil, err
	}

	return &pb.TransactionResponse{
		TransactionId: req.TransactionId,
		Status:        "success",
		Message:       "Transaction processed successfully",
	}, nil
}

func (l *Layer2Handler) MonitorBatchStatus(req *pb.BatchStatusRequest, stream pb.Layer2Service_MonitorBatchStatusServer) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	file, err := os.Open("store/transaction_pool.txt")
	if err != nil {
		return status.Errorf(codes.Internal, "Gagal membuka transaction pool: %v", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		// Cek jika context sudah dibatalkan
		if stream.Context().Err() != nil {
			return status.Errorf(codes.Canceled, "Stream dibatalkan oleh client")
		}

		line := scanner.Text()
		var tx pb.TransactionRequest
		if err := json.Unmarshal([]byte(line), &tx); err != nil {
			continue
		}

		// Kirim status untuk setiap transaksi
		status := &pb.BatchStatusResponse{
			BatchId:  req.BatchId,
			Status:   "pending",
			Progress: tx.TransactionId,
			Message:  fmt.Sprintf("Transaction %s is pending", tx.TransactionId),
		}

		if err := stream.Send(status); err != nil {
			return fmt.Errorf("Failed to send batch status: %e\", err")
		}
	}

	if err := scanner.Err(); err != nil {
		return status.Errorf(codes.Internal, "Error reading transaction pool: %e", err)
	}

	return nil
}

func (l *Layer2Handler) TriggerBatchProcessing(ctx context.Context, req *pb.BatchProcessingRequest) (*pb.BatchProcessingResponse, error) {
	client := l.Client

	// Private key asli (dengan prefix 0x)
	systemPrivateKey := "0xbeda603dae5f7fa6cf6235c27f160ad80e1fa5faf8722519e07e577681f6cf40"

	// Hapus prefix "0x" jika ada
	systemPrivateKey = strings.TrimPrefix(systemPrivateKey, "0x")

	// Hapus spasi yang tidak diinginkan
	systemPrivateKey = strings.TrimSpace(systemPrivateKey)

	// Log private key yang telah dibersihkan dan panjangnya
	log2.Printf("Cleaned Private Key: %s", systemPrivateKey)
	log2.Printf("Private Key Length: %d", len(systemPrivateKey))

	// Validasi hex string untuk private key
	if !isValidHex(systemPrivateKey) {
		return nil, fmt.Errorf("private key tidak valid: mengandung karakter yang tidak valid atau panjangnya salah")
	}

	// Debug: log string private key untuk memastikan itu adalah hex yang valid
	log2.Printf("Private Key (Hex): %s", systemPrivateKey)

	// Pastikan bahwa hex string yang diterima valid
	decodedKey, err := hex.DecodeString(systemPrivateKey)
	if err != nil {
		log2.Printf("Gagal mendekodekan hex: %v", err)
		return nil, fmt.Errorf("gagal mendekodekan hex untuk private key: %v", err)
	}

	// Debug: log decoded key dalam format byte untuk melihat apakah sudah benar
	log2.Printf("Decoded Key (Bytes): %v", decodedKey)

	// Parsing private key dari string hex yang valid
	privKey, err := crypto.HexToECDSA(systemPrivateKey)
	if err != nil {
		return nil, fmt.Errorf("gagal memproses private key: %v", err)
	}

	// Mengonversi private key menjadi format hex yang benar
	privKeyHex := hex.EncodeToString(privKey.D.Bytes()) // Menggunakan D.Bytes() dan encode ke hex

	// Debugging: log private key dalam format hex
	log2.Printf("Private Key Hex: %s", privKeyHex)

	// Lanjutkan dengan proses batch insertion
	err = batchInsert(ctx, client, privKeyHex)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Gagal memproses batch: %v", err)
	}

	if err := cleanTransactionPool(); err != nil {
		log.Warn("Gagal membersihkan transaction pool", "error", err)
	}

	return &pb.BatchProcessingResponse{
		BatchId: req.TriggerBy,
		Status:  "success",
		Message: "Batch berhasil diproses",
	}, nil
}

// Fungsi untuk memvalidasi sebuah string hex (hanya karakter hexadecimal dan panjang 64)
func isValidHex(hexStr string) bool {
	// Pastikan string memiliki panjang tepat 64 karakter
	if len(hexStr) != 64 {
		log2.Printf("Panjang hex tidak valid: Diharapkan 64 karakter, tapi didapat %d", len(hexStr))
		return false
	}

	// Cek apakah string hanya mengandung karakter hex yang valid
	for _, c := range hexStr {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
			log2.Printf("Karakter tidak valid '%c' ditemukan dalam private key", c)
			return false
		}
	}

	return true
}

// filter checking insert User
func filter(ctx context.Context, c *ethclient.Client, sender ecdsa.PublicKey, valueSend *big.Int) error {
	// checking amount sender
	address := crypto.PubkeyToAddress(sender)
	balanceRN, err := c.BalanceAt(ctx, address, nil)
	if err != nil {
		return status.Errorf(codes.Internal, "failed to retrieve balance: %v", err)
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

// cleanTransactionPool - fungsi helper untuk membersihkan transaction pool
func cleanTransactionPool() error {
	return os.WriteFile("store/transaction_pool.txt", []byte(""), 0644)
}

func main() {

	ctx := context.Background()

	//var URL = "http://127.0.0.1:8545"
	var URL = "https://sepolia.infura.io/v3/7108f6b019944d2082df7b667e6b1f4a"
	ethClient, err := ethclient.DialContext(ctx, URL)
	if err != nil {
		log2.Fatalf("Failed to connect to Ethereum client: %e\n", err)
	}

	layer2Handler := NewLayer2Handler(
		pb.UnimplementedLayer2ServiceServer{},
		sync.Mutex{},
		[]*pb.TransactionRequest{},
		ethClient)

	//GRPC SERVER
	server := grpc.NewServer()
	pb.RegisterLayer2ServiceServer(server, layer2Handler)

	listener, err := net.Listen("tcp", ":50051")
	if err != nil {
		log2.Fatalf("Failed to listen on port 50051: %v", err)
	}

	log2.Println("Server gRPC berjalan pada port 50051...")
	if err := server.Serve(listener); err != nil {
		log2.Fatalf("Failed to serve gRPC server: %v", err)
	}
}
