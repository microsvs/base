package base

import (
	"context"
	"net/http"

	"github.com/microsvs/base/pkg/rpc"
	"github.com/microsvs/base/pkg/types"
	"github.com/segmentio/ksuid"
)

// external http context
func buildContext(r *http.Request) (context.Context, error) {
	var (
		ctx      context.Context = context.Background()
		token    string
		user     *types.User
		retToken *types.Token
		err      error
	)
	ctx = context.WithValue(ctx, rpc.KeyRawRequest, r)
	ctx = context.WithValue(ctx, rpc.KeyTraceID, getTraceIdFromRequest(r))
	ctx = context.WithValue(ctx, rpc.KeyRPCID, getRPCIdFromRequest(r))
	// token
	if token = getTokenFromRequest(r); len(token) > 0 {
		if retToken, err = rpc.GetUserIdFromTokenRPC(ctx, Service2Url(rpc.FGSToken), token); err != nil {
			return nil, err
		}
		if user, err = rpc.GetUserFromIdRPC(ctx, Service2Url(rpc.FGSUser), retToken.UserId); err != nil {
			return nil, err
		}
		ctx = context.WithValue(ctx, rpc.KeyUser, user)
	}
	return ctx, nil
}

func getTraceIdFromRequest(r *http.Request) string {
	var traceid string
	// first: url params from nonce
	if traceid = r.URL.Query().Get("nonce"); len(traceid) > 0 {
		return traceid
	}
	// second: X-Trace-ID from header
	if traceid = r.Header.Get("X-Trace-ID"); len(traceid) > 0 {
		return traceid
	}
	// finally: generate traceid
	return ksuid.New().String()
}

func getRPCIdFromRequest(r *http.Request) string {
	var rpcid string
	if rpcid = r.Header.Get("X-Trace-RPCID"); len(rpcid) > 0 {
		return rpcid
	}
	return ksuid.New().String()
}

func getTokenFromRequest(r *http.Request) string {
	return r.URL.Query().Get("token")
}
