package wire

import (
	"context"
	"net"
	"sync"

	"github.com/jeroenrinzema/kafka-wire/internal/buffer"
	"github.com/jeroenrinzema/kafka-wire/internal/protocol"
	"go.uber.org/zap"
)

func NewServer(options ...OptionFn) *Server {
	return &Server{
		logger: zap.NewNop(),
	}
}

type Server struct {
	logger *zap.Logger
	wg     sync.WaitGroup
}

// ListenAndServe opens a new Kafka server on the preconfigured address and
// starts accepting and serving incoming client connections.
func (srv *Server) ListenAndServe(address string) error {
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return err
	}

	return srv.Serve(listener)
}

// Serve accepts and serves incoming Kafka client connections using the
// preconfigured configurations. The given listener will be closed once the
// server is gracefully closed.
func (srv *Server) Serve(listener net.Listener) error {
	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			return err
		}

		srv.wg.Add(1)

		go func() {
			defer srv.wg.Done()
			ctx := context.Background()
			err = srv.serve(ctx, conn)
			if err != nil {
				srv.logger.Error("an unexpected error got returned while serving a client connection", zap.Error(err))
			}
		}()
	}
}

func (srv *Server) serve(ctx context.Context, conn net.Conn) (err error) {
	reader := buffer.NewReader(conn, buffer.DefaultBufferSize)

	for {
		_, err = reader.ReadMessage()
		if err != nil {
			return err
		}

		header := &protocol.RequestHeader{}
		err = header.Decode(reader)
		if err != nil {
			return err
		}

		srv.logger.Debug("incoming request!", zap.Any("header", header))
	}
}
