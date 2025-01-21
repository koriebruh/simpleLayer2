package main

import (
	"context"
	"google.golang.org/grpc/credentials/insecure"
	"log"
	"time"

	pb "github.com/koriebruh/simpleLayer2/proto/layer2"
	"google.golang.org/grpc"
)

func main() {
	// Dial ke server gRPC
	conn, err := grpc.Dial(
		":50051",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		log.Fatalf("Gagal terhubung ke server: %v", err)
	}
	defer conn.Close()
	defer conn.Close()

	client := pb.NewLayer2ServiceClient(conn)

	// 1. SubmitTransaction
	log.Println("Mengirim SubmitTransaction...")
	req := &pb.TransactionRequest{
		TransactionId: "12345",
		Sender:        "0x04b631a8c2fda7d3c00254729abddcb3ddcb93c9698f4e35960c702b4895ee0b453e234ce278e56d39c647352f706a59cc8690d953beb061ed2955dead20018517",
		Receiver:      "0x0D0475Cfa45b5E8C4b21B4F84A4322f17D77c2a2",
		Amount:        1000, // Uang dalam wei
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	res, err := client.SubmitTransaction(ctx, req)
	if err != nil {
		log.Fatalf("Error SubmitTransaction: %v", err)
	}
	log.Printf("SubmitTransaction Response: %v", res)

	// 2. MonitorBatchStatus
	log.Println("Memulai MonitorBatchStatus...")
	go monitorBatchStatus(client)

	// 3. TriggerBatchProcessing
	log.Println("Memanggil TriggerBatchProcessing...")
	triggerBatchProcessing(client)
}

func monitorBatchStatus(client pb.Layer2ServiceClient) {
	req := &pb.BatchStatusRequest{
		BatchId: "batch123",
	}

	// Membuat stream untuk menerima respons batch status
	stream, err := client.MonitorBatchStatus(context.Background(), req)
	if err != nil {
		log.Fatalf("Error MonitorBatchStatus: %v", err)
	}

	for {
		// Menerima setiap respons dari stream
		res, err := stream.Recv()
		if err != nil {
			log.Printf("Stream selesai atau error: %v", err)
			break
		}

		log.Printf("Batch Status Response: %v", res)
	}
}

func triggerBatchProcessing(client pb.Layer2ServiceClient) {
	req := &pb.BatchProcessingRequest{
		TriggerBy: "admin123",
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	res, err := client.TriggerBatchProcessing(ctx, req)
	if err != nil {
		log.Fatalf("Error TriggerBatchProcessing: %v", err)
	}

	log.Printf("TriggerBatchProcessing Response: %v", res)
}
