package logging

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/transport"
)

var _ transport.Transporter = &Transport{}

type Transport struct {
	kind      transport.Kind
	endpoint  string
	operation string
}

func (tr *Transport) Kind() transport.Kind {
	return tr.kind
}

func (tr *Transport) Endpoint() string {
	return tr.endpoint
}

func (tr *Transport) Operation() string {
	return tr.operation
}

func (tr *Transport) RequestHeader() transport.Header {
	return nil
}

func (tr *Transport) ReplyHeader() transport.Header {
	return nil
}

func TestHTTP(t *testing.T) {
	err := errors.New("reply.error")
	bf := bytes.NewBuffer(nil)
	logger := log.NewStdLogger(bf)

	tests := []struct {
		name string
		kind func(logger log.Logger) middleware.Middleware
		err  error
		ctx  context.Context
	}{
		{
			"http-server@fail",
			Server,
			err,
			func() context.Context {
				return transport.NewServerContext(context.Background(), &Transport{kind: transport.KindHTTP, endpoint: "endpoint", operation: "/package.service/method"})
			}(),
		},
		{
			"http-server@succ",
			Server,
			nil,
			func() context.Context {
				return transport.NewServerContext(context.Background(), &Transport{kind: transport.KindHTTP, endpoint: "endpoint", operation: "/package.service/method"})
			}(),
		},
		{
			"http-client@succ",
			Client,
			nil,
			func() context.Context {
				return transport.NewClientContext(context.Background(), &Transport{kind: transport.KindHTTP, endpoint: "endpoint", operation: "/package.service/method"})
			}(),
		},
		{
			"http-client@fail",
			Client,
			err,
			func() context.Context {
				return transport.NewClientContext(context.Background(), &Transport{kind: transport.KindHTTP, endpoint: "endpoint", operation: "/package.service/method"})
			}(),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			bf.Reset()
			next := func(ctx context.Context, req interface{}) (interface{}, error) {
				return "reply", test.err
			}
			next = test.kind(logger)(next)
			v, e := next(test.ctx, "req.args")
			t.Logf("[%s]reply: %v, error: %v", test.name, v, e)
			t.Logf("[%s]log:%s", test.name, bf.String())
		})
	}
}

func Test_extractArgs(t *testing.T) {
	tests := []struct {
		name    string
		reqLen  int
		wantLen int
	}{
		{
			name:    "length@double",
			reqLen:  halfReqLength * 2,
			wantLen: halfReqLength * 2,
		},
		{
			name:    "length@without",
			reqLen:  halfReqLength*2 + 1,
			wantLen: halfReqLength*2 + len(" ... "),
		},
		{
			name:    "length@within",
			reqLen:  halfReqLength*2 - 1,
			wantLen: halfReqLength*2 - 1,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			req := strings.Repeat("-", test.reqLen)
			res := extractArgs(req)
			if len(res) != test.wantLen {
				t.Errorf("want length %d got %d res is (%s)", test.wantLen, len(res), res)
			}
		})
	}
}
