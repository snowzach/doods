package detector

import (
	"context"
	"strings"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"github.com/snowzach/doods/odrpc"
)

// AuthFuncOverride will handle authentication
func (m *Mux) AuthFuncOverride(ctx context.Context, fullMethodName string) (context.Context, error) {

	// Auth disabled
	if m.authKey == "" {
		return ctx, nil
	}

	// Get request metadata
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ctx, status.Errorf(codes.PermissionDenied, "Permission Denied")
	}

	// Get the user pubKeyString
	if mdfirst(md, odrpc.DoodsAuthKeyHeader) != m.authKey {
		return ctx, status.Errorf(codes.PermissionDenied, "Invalid Login")
	}

	return ctx, nil

}

func mdfirst(md metadata.MD, key string) string {
	val := md.Get(strings.ToLower(key))
	if len(val) > 0 {
		return val[0]
	}
	return ""
}
