package grpc

import pb "github.com/Yakwilik/MRGAbackend/internal/pb/ailawyer"

func (a *Implementation) TestStreamFromServer(request *pb.TestMessageRequest, server pb.Backend_TestStreamFromServerServer) error {
	//TODO implement me
	panic("implement me")
}

func (a *Implementation) TestStreamFromClient(server pb.Backend_TestStreamFromClientServer) error {
	return server.SendAndClose(&pb.TestMessageResponse{})
}

func (a *Implementation) TestStreamClientAndServer(server pb.Backend_TestStreamClientAndServerServer) error {
	//TODO implement me
	panic("implement me")
}
