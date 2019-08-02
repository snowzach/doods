package server

import (
	"context"

	emptypb "github.com/golang/protobuf/ptypes/empty"

	"github.com/snowzach/doods/conf"
	"github.com/snowzach/doods/server/rpc"
)

// Version returns the version
func (s *Server) Version(ctx context.Context, _ *emptypb.Empty) (*rpc.VersionResponse, error) {

	return &rpc.VersionResponse{
		Version: conf.GitVersion,
	}, nil

}
