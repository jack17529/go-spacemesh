package grpcserver

import (
	"github.com/golang/protobuf/ptypes/empty"
	pb "github.com/spacemeshos/api/release/go/spacemesh/v1"
	"github.com/spacemeshos/go-spacemesh/activation"
	"github.com/spacemeshos/go-spacemesh/api"
	"github.com/spacemeshos/go-spacemesh/common/types"
	"github.com/spacemeshos/go-spacemesh/common/util"
	"github.com/spacemeshos/go-spacemesh/log"
	"golang.org/x/net/context"
	"google.golang.org/genproto/googleapis/rpc/code"
	rpcstatus "google.golang.org/genproto/googleapis/rpc/status"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// SmesherService exposes endpoints to manage smeshing
type SmesherService struct {
	post     api.PostAPI
	smeshing api.SmeshingAPI
}

// RegisterService registers this service with a grpc server instance
func (s SmesherService) RegisterService(server *Server) {
	pb.RegisterSmesherServiceServer(server.GrpcServer, s)
}

// NewSmesherService creates a new grpc service using config data.
func NewSmesherService(post api.PostAPI, smeshing api.SmeshingAPI) *SmesherService {
	return &SmesherService{post, smeshing}
}

// IsSmeshing reports whether the node is smeshing
func (s SmesherService) IsSmeshing(ctx context.Context, in *empty.Empty) (*pb.IsSmeshingResponse, error) {
	log.Info("GRPC SmesherService.IsSmeshing")

	return &pb.IsSmeshingResponse{IsSmeshing: s.smeshing.Smeshing()}, nil
}

// StartSmeshing requests that the node begin smeshing
func (s SmesherService) StartSmeshing(ctx context.Context, in *pb.StartSmeshingRequest) (*pb.StartSmeshingResponse, error) {
	log.Info("GRPC SmesherService.StartSmeshing")

	// TODO(moshababo): remove DataDir and CommitmentSize from the request proto definition.
	// TODO(moshababo): check why JSON request via HTTP gateway doesn't decode `coinbase` properly

	if in.Coinbase == nil {
		return nil, status.Errorf(codes.InvalidArgument, "`Coinbase` must be provided")
	}

	addr := types.BytesToAddress(in.Coinbase.Address)
	if err := s.smeshing.StartSmeshing(addr); err != nil {
		log.Error("StartSmeshing failure: %v", err) // TODO: should not crash nor print stacktrace.
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.StartSmeshingResponse{
		Status: &rpcstatus.Status{Code: int32(code.Code_OK)},
	}, nil
}

// StopSmeshing requests that the node stop smeshing
func (s SmesherService) StopSmeshing(ctx context.Context, in *pb.StopSmeshingRequest) (*pb.StopSmeshingResponse, error) {
	log.Info("GRPC SmesherService.StopSmeshing")

	if err := s.smeshing.StopSmeshing(); err != nil {
		return nil, err
	}
	return &pb.StopSmeshingResponse{
		Status: &rpcstatus.Status{Code: int32(code.Code_OK)},
	}, nil
}

// SmesherID returns the smesher ID of this node
func (s SmesherService) SmesherID(ctx context.Context, in *empty.Empty) (*pb.SmesherIDResponse, error) {
	log.Info("GRPC SmesherService.SmesherID")

	addr := util.Hex2Bytes(s.smeshing.SmesherID().Key)
	return &pb.SmesherIDResponse{AccountId: &pb.AccountId{Address: addr}}, nil
}

// Coinbase returns the current coinbase setting of this node
func (s SmesherService) Coinbase(ctx context.Context, in *empty.Empty) (*pb.CoinbaseResponse, error) {
	log.Info("GRPC SmesherService.Coinbase")

	addr := s.smeshing.Coinbase()
	return &pb.CoinbaseResponse{AccountId: &pb.AccountId{Address: addr[:]}}, nil
}

// SetCoinbase sets the current coinbase setting of this node
func (s SmesherService) SetCoinbase(ctx context.Context, in *pb.SetCoinbaseRequest) (*pb.SetCoinbaseResponse, error) {
	log.Info("GRPC SmesherService.SetCoinbase")

	if in.Id == nil {
		return nil, status.Errorf(codes.InvalidArgument, "`Id` must be provided")
	}

	addr := types.BytesToAddress(in.Id.Address)
	s.smeshing.SetCoinbase(addr)

	return &pb.SetCoinbaseResponse{
		Status: &rpcstatus.Status{Code: int32(code.Code_OK)},
	}, nil
}

// MinGas returns the current mingas setting of this node
func (s SmesherService) MinGas(ctx context.Context, in *empty.Empty) (*pb.MinGasResponse, error) {
	log.Info("GRPC SmesherService.MinGas")
	return nil, status.Errorf(codes.Unimplemented, "this endpoint is not implemented")
}

// SetMinGas sets the mingas setting of this node
func (s SmesherService) SetMinGas(ctx context.Context, in *pb.SetMinGasRequest) (*pb.SetMinGasResponse, error) {
	log.Info("GRPC SmesherService.SetMinGas")
	return nil, status.Errorf(codes.Unimplemented, "this endpoint is not implemented")
}

// PostStatus returns post data status
func (s SmesherService) PostStatus(ctx context.Context, in *empty.Empty) (*pb.PostStatusResponse, error) {
	log.Info("GRPC SmesherService.PostStatus")

	status, err := s.post.PostStatus()
	if err != nil {
		return nil, err
	}
	return &pb.PostStatusResponse{Status: statusToPbStatus(status)}, nil
}

// PostComputeProviders returns a list of available post compute providers
func (s SmesherService) PostComputeProviders(ctx context.Context, in *empty.Empty) (*pb.PostComputeProvidersResponse, error) {
	log.Info("GRPC SmesherService.PostComputeProviders")

	providers := s.post.PostComputeProviders()

	res := &pb.PostComputeProvidersResponse{}
	res.PostComputeProvider = make([]*pb.PostComputeProvider, len(providers))
	for i, p := range providers {
		hs, err := p.Benchmark()
		if err != nil {
			return nil, err
		}

		res.PostComputeProvider[i] = &pb.PostComputeProvider{
			Id:          uint32(p.Id),
			Model:       p.Model,
			ComputeApi:  pb.ComputeApiClass(p.ComputeAPI), // assuming enum values match.
			Performance: uint64(hs),
		}
	}

	return res, nil
}

// CreatePostData requests that the node begin post init
func (s SmesherService) CreatePostData(ctx context.Context, in *pb.CreatePostDataRequest) (*pb.CreatePostDataResponse, error) {
	log.Info("GRPC SmesherService.CreatePostData")

	if _, err := s.post.CreatePostData(&activation.PostOptions{
		DataDir:           in.Data.Path,
		DataSize:          in.Data.DataSize,
		Append:            in.Data.Append,
		Throttle:          in.Data.Throttle,
		ComputeProviderId: uint(in.Data.ProviderId),
	}); err != nil {
		return nil, err
	}

	return &pb.CreatePostDataResponse{
		Status: &rpcstatus.Status{Code: int32(code.Code_OK)},
	}, nil
}

// StopPostDataCreationSession requests that the node stop ongoing post data creation
func (s SmesherService) StopPostDataCreationSession(ctx context.Context, in *pb.StopPostDataCreationSessionRequest) (*pb.StopPostDataCreationSessionResponse, error) {
	log.Info("GRPC SmesherService.StopPostDataCreationSession")

	if err := s.post.StopPostDataCreationSession(in.DeleteFiles); err != nil {
		return nil, err
	}

	return &pb.StopPostDataCreationSessionResponse{
		Status: &rpcstatus.Status{Code: int32(code.Code_OK)},
	}, nil

}

// STREAMS

// PostDataCreationProgressStream exposes a stream of updates during post init
func (s SmesherService) PostDataCreationProgressStream(request *empty.Empty, stream pb.SmesherService_PostDataCreationProgressStreamServer) error {
	log.Info("GRPC SmesherService.PostDataCreationProgressStream")

	for status := range s.post.PostDataCreationProgressStream() {
		if err := stream.Send(&pb.PostDataCreationProgressStreamResponse{Status: statusToPbStatus(status)}); err != nil {
			return err
		}
	}

	// TODO(moshababo): check if stop streaming on stream.Context().Done() is required

	return nil
}

func statusToPbStatus(status *activation.PostStatus) *pb.PostStatus {
	pbStatus := &pb.PostStatus{}
	pbStatus.InitInProgress = status.InitInProgress
	pbStatus.BytesWritten = status.BytesWritten
	pbStatus.ErrorType = pb.PostStatus_ErrorType(status.ErrorType) // assuming enum values match.
	pbStatus.ErrorMessage = status.ErrorMessage
	if status.LastOptions != nil {
		pbStatus.PostData = &pb.PostData{
			Path:       status.LastOptions.DataDir,
			DataSize:   status.LastOptions.DataSize,
			Append:     status.LastOptions.Append,
			Throttle:   status.LastOptions.Throttle,
			ProviderId: uint32(status.LastOptions.ComputeProviderId),
		}
	}
	return pbStatus
}
