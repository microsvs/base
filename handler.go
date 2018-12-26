package base

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/graphql-go/graphql"
	"github.com/graphql-go/graphql/gqlerrors"
	"github.com/microsvs/base/cmd/discovery"
	"github.com/microsvs/base/pkg/env"
	"github.com/microsvs/base/pkg/errors"
	"github.com/microsvs/base/pkg/log"
	"github.com/microsvs/base/pkg/rpc"
	"github.com/microsvs/base/pkg/tracing"
	"github.com/microsvs/base/pkg/types"
	"github.com/microsvs/handler"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/uber/jaeger-lib/metrics"
	jprom "github.com/uber/jaeger-lib/metrics/prometheus"
	"github.com/urfave/negroni"
)

// dns configuration
var BaseDomain string = "api.xhj.com"
var metricsFactory metrics.Factory = jprom.New().Namespace(
	func() string {
		name, _ := env.Get(env.ServiceName)
		return name
	}(),
	nil,
)

type PHASES string

const (
	BEFORE PHASES = "before"
	AFTER         = "after"
)

type Daemon struct {
	service     rpc.FGService
	extHandlers map[string]http.Handler
	schema      *graphql.Schema
	handler     *handler.Handler
	middlewares *negroni.Negroni
	phasesMap   map[PHASES][]http.HandlerFunc
}

// example: "FGError:40011:invalid user"
func customErrorFormat(errs []gqlerrors.FormattedError) interface{} {
	if len(errs) <= 0 {
		return &types.CustomError{}
	}
	msg := errs[0].Message
	if strings.HasPrefix(msg, errors.FGErrorPrefix) {
		fields := strings.SplitN(strings.TrimLeft(msg, errors.FGErrorPrefix), ":", 2)
		if len(fields) >= 2 {
			code, _ := strconv.Atoi(fields[0])
			return &types.CustomError{
				ErrCode: code,
				ErrMsg:  fields[1],
			}
		}
	}
	ce := &types.CustomError{
		ErrCode: int(errors.FGEInternalError),
		ErrMsg:  msg,
	}
	for code, errMsg := range errors.GetErrors() {
		if msg == errMsg {
			ce.ErrCode = int(code)
			break
		}
	}
	return ce
}

func NewGLDaemon(service rpc.FGService, schema *graphql.Schema) (*Daemon, error) {
	var d *Daemon
	if schema == nil {
		return nil, errors.GraphqlObjectIsNull
	}
	config := handler.NewConfig()
	config.Schema = schema
	config.HandlerErrorResp = customErrorFormat
	d = &Daemon{
		service:     service,
		extHandlers: make(map[string]http.Handler),
		schema:      schema,
		handler:     handler.New(config),
		middlewares: negroni.New(func() *negroni.Recovery {
			recovery := negroni.NewRecovery()
			recovery.PrintStack = false
			recovery.Formatter = nil
			return recovery
		}()),
		phasesMap: make(map[PHASES][]http.HandlerFunc),
	}

	// init global tracer
	tracer := tracing.Init(
		service.String(),
		metricsFactory.Namespace(service.String(), nil),
		func() string {
			agent, _ := env.Get(env.TracerAgent)
			return agent
		}(),
	)
	opentracing.SetGlobalTracer(tracer)
	return d, nil
}

func (d *Daemon) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var (
		ctx context.Context
		err error
	)
	defer func() {
		if err != nil {
			GLReturnError(err, w)
		}
	}()
	switch getProtocalType(r) {
	case rpc.HTTP: // external request
		if ctx, err = buildContext(r); err != nil {
			// error handler
			return
		}
		ctx = context.WithValue(ctx, rpc.KeyService, d.service.String())
	case rpc.RPC: // interval request, between microservices
		if ctx, err = rpc.ContextFromHTTPRequest(ctx, r); err != nil {
			return
		}
	}
	d.handler.ContextHandler(ctx, w, r)
	return
}

func GLReturnError(err error, w http.ResponseWriter) {
	tmp := customErrorFormat(
		[]gqlerrors.FormattedError{
			gqlerrors.FormattedError{
				Message: err.Error(),
			},
		},
	)
	resp := map[string]interface{}{
		"data":  nil,
		"error": tmp,
	}
	w.WriteHeader(http.StatusOK)
	bts, _ := json.MarshalIndent(resp, "", "\t")
	w.Write(bts)
	return
}

func Service2Url(service rpc.FGService) string {
	host := fmt.Sprintf("dns/%s.%s", service, BaseDomain)
	return discovery.KVRead(host, host)
}

// get protocal type from http.Request, http or rpc
func getProtocalType(r *http.Request) rpc.ProtocalType {
	tmp := r.Header.Get(fmt.Sprintf("%d", rpc.KeyProtocalType))
	if tmp == "" {
		return rpc.HTTP
	}
	return rpc.ProtocalType(tmp)
}

func (d *Daemon) BeforeRouter(middleware ...http.HandlerFunc) {
	d.phasesMap[BEFORE] = append([]http.HandlerFunc{}, middleware...)
}

func (d *Daemon) AfterRouter(middleware ...http.HandlerFunc) {
	d.phasesMap[AFTER] = append([]http.HandlerFunc{}, middleware...)
}

// generate router handlers
func (d *Daemon) Listen() {
	mux := tracing.NewServeMux(opentracing.GlobalTracer())
	mux.Handle("/graphql", d)
	mux.Handle("/", http.FileServer(assetFS()))
	for url, handler := range d.extHandlers {
		mux.Handle(url, handler)
	}

	// before router
	for _, fn := range d.phasesMap[BEFORE] {
		d.middlewares.UseHandlerFunc(fn)
	}

	d.middlewares.UseHandler(mux)

	// after router
	for _, fn := range d.phasesMap[AFTER] {
		d.middlewares.UseHandlerFunc(fn)
	}

	log.InfoRaw("service %s start at %d", d.service.String(), d.service)
	http.ListenAndServe(fmt.Sprintf("0.0.0.0:%d", d.service), d.middlewares)
	return
}
