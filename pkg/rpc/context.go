package rpc

import (
	"bytes"
	"context"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"net/http"
	"reflect"

	"github.com/microsvs/base/pkg/types"
	"github.com/mitchellh/mapstructure"
	"github.com/vmihailenco/msgpack"
)

const RPC__CONTEXT = "rpc_context_"

// header key or value
type KeyContext int

const (
	KeyTraceID KeyContext = iota
	KeyRPCID
	KeyService
	KeyRawRequest
	KeyMobile
	KeyUser
	KeyConsoleInfo
	KeyProtocalType
	KeySource
	KeyCuid
	KeyLocation
	KeySuid
	KeyAppVersion
	KeySourceVersion
	KeyToken
	KeyCid
	KeyDevice
	KeyRemoteIp
)

// 用于区分内部调用，还是外部调用
type ProtocalType string

const (
	HTTP ProtocalType = "http"
	RPC               = "rpc"
)

func gobRegister() {
	gob.Register(map[KeyContext]interface{}{})
	gob.Register(&types.User{})
	gob.Register(&types.ConsoleInfo{})
}

func init() {
	gobRegister()
}

// context.Context storage request in gob encoder
func ContextToHTTPRequest(ctx context.Context, r *http.Request) error {
	var (
		gobVal string
		err    error
	)
	request := GetContextFromKey(ctx, KeyRawRequest, &http.Request{}).(*http.Request)
	r.Header = copyHeader(request.Header)
	rpcid := GetContextFromKey(ctx, KeyRPCID, "0").(string)
	traceid := GetContextFromKey(ctx, KeyTraceID, "-").(string)
	user := GetContextFromKey(ctx, KeyUser, &types.User{}).(*types.User)
	console := GetContextFromKey(ctx, KeyConsoleInfo, &types.ConsoleInfo{}).(*types.ConsoleInfo)
	values := map[KeyContext]interface{}{
		KeyRPCID:       rpcid,
		KeyTraceID:     traceid,
		KeyUser:        user,
		KeyConsoleInfo: console,
	}
	if gobVal, err = toMsgpack(values); err != nil {
		return err
	}
	/*
		if gobVal, err = toGOB(values); err != nil {
			return err
		}
	*/
	r.Header.Set(RPC__CONTEXT, gobVal)
	r.Header.Set(fmt.Sprintf("%d", KeyProtocalType), RPC)
	return nil
}

func ContextFromHTTPRequest(ctx context.Context, r *http.Request) (context.Context, error) {
	var (
		err error
		m   map[KeyContext]interface{}
	)
	if ctx == nil {
		ctx = context.Background()
	}
	ctx = context.WithValue(ctx, KeyRawRequest, r)
	protocal := r.Header.Get(fmt.Sprintf("%d", KeyProtocalType))
	ctx = context.WithValue(ctx, KeyProtocalType, protocal)

	gobVal := r.Header.Get(RPC__CONTEXT)
	if m, err = fromMsgpack(gobVal); err != nil {
		return nil, err
	}
	/*
		if m, err = fromGOB(gobVal); err != nil {
			return nil, err
		}
	*/
	rpcid := m[KeyRPCID].(string)
	traceid := m[KeyTraceID].(string)
	var (
		user    = new(types.User)
		console = new(types.ConsoleInfo)
	)
	if err = mapstructure.Decode(m[KeyUser], user); err != nil {
		return nil, err
	}
	if err = mapstructure.Decode(m[KeyConsoleInfo], console); err != nil {
		return nil, err
	}

	//user := m[KeyUser].(*types.User)
	//console := m[KeyConsoleInfo].(*types.ConsoleInfo)
	ctx = context.WithValue(ctx, KeyRPCID, rpcid)
	ctx = context.WithValue(ctx, KeyTraceID, traceid)
	ctx = context.WithValue(ctx, KeyUser, user)
	ctx = context.WithValue(ctx, KeyConsoleInfo, console)
	return ctx, nil
}

func toMsgpack(target map[KeyContext]interface{}) (string, error) {
	var (
		buf = new(bytes.Buffer)
		err error
	)
	enc := msgpack.NewEncoder(buf)
	if err = enc.Encode(target); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf.Bytes()), nil
}

func fromMsgpack(target string) (map[KeyContext]interface{}, error) {
	var (
		buf = new(bytes.Buffer)
		bts []byte
		err error
		m   = map[KeyContext]interface{}{}
	)
	if bts, err = hex.DecodeString(target); err != nil {
		return nil, err
	}
	buf.Write(bts)
	dec := msgpack.NewDecoder(buf)
	if err = dec.Decode(&m); err != nil {
		fmt.Printf("msgpack decode body failed. err=%s\n", err.Error())
		return nil, err
	}
	fmt.Printf("parse body, result: %v\n", m)
	return m, nil
}

func toGOB(target map[KeyContext]interface{}) (string, error) {
	var (
		buf = new(bytes.Buffer)
		err error
	)
	enc := gob.NewEncoder(buf)
	if err = enc.Encode(target); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf.Bytes()), nil
}

func fromGOB(target string) (map[KeyContext]interface{}, error) {
	var (
		buf = new(bytes.Buffer)
		bts []byte
		err error
		m   = map[KeyContext]interface{}{}
	)
	if bts, err = hex.DecodeString(target); err != nil {
		return nil, err
	}
	buf.Write(bts)
	dec := gob.NewDecoder(buf)
	if err = dec.Decode(&m); err != nil {
		return nil, err
	}
	fmt.Printf("parse body success, result: %v\n", m)
	return m, nil
}

func GetContextFromKey(ctx context.Context, key KeyContext, def interface{}) interface{} {
	if ctx == nil || reflect.ValueOf(ctx).IsNil() {
		return def
	}
	if value := ctx.Value(key); value != nil {
		return value
	}
	return def
}

func copyHeader(header http.Header) http.Header {
	h := make(http.Header)
	for k, vs := range header {
		h[k] = vs
	}
	return h
}
