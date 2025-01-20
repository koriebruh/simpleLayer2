package main

import (
	"context"
	pb "github.com/koriebruh/simpleLayer2/proto/layer2"
	"sync"
)

type layer2Server struct {
	pb.UnimplementedLayer2ServiceServer
	mu          sync.Mutex
	transaction []*pb.TransactionRequest
}

func (l *layer2Server) SubmitTransaction(ctx context.Context, request *pb.TransactionRequest) (*pb.TransactionResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (l *layer2Server) MonitorBatchStatus(ctx context.Context, request *pb.BatchStatusRequest, stream pb.Layer2Service_MonitorBatchStatusClient) error {
	//TODO implement me
	panic("implement me")
}

func (l *layer2Server) TriggerBatchProcessing(ctx context.Context, request *pb.BatchProcessingRequest) (*pb.BatchProcessingResponse, error) {
	//TODO implement me
	panic("implement me")
}

func main() {

}
