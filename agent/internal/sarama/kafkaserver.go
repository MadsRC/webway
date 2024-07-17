package sarama

import (
	"fmt"
	"github.com/MadsRC/sarama"
	"log"
	"net"
	"sync"
)

var (
	ErrServerStopped = fmt.Errorf("server stopped")
)

type KafkaServer struct {
	opts  kafkaServerOptions
	mu    sync.Mutex
	lis   net.Listener
	serve bool
}

func NewKafkaServer(opt ...KafkaServerOption) (*KafkaServer, error) {
	opts := defaultKafkaServerOptions
	for _, o := range globalKafkaServerOptions {
		o.apply(&opts)
	}
	for _, o := range opt {
		o.apply(&opts)
	}
	s := &KafkaServer{opts: opts}

	return s, nil
}

func (s *KafkaServer) GracefulStop() {}

// Serve accepts incoming connections on the listener lis, creating a new service goroutine for each.
//
// Serve panics if lis is nil.
func (s *KafkaServer) Serve(lis net.Listener) error {
	s.mu.Lock()
	s.serve = true

	defer func() {
		s.mu.Lock()
		if s.lis != nil {
			s.lis.Close()
		}
		s.lis = nil
		s.serve = false
		s.mu.Unlock()
	}()

	s.lis = lis
	s.mu.Unlock()

	for {
		conn, err := lis.Accept()
		if err != nil {
			return err
		}
		go func() {
			s.handleConn(conn)
		}()
	}
}

func (s *KafkaServer) handleConn(conn net.Conn) {
	req, _, err := sarama.DecodeRequest(conn)
	if err != nil {
		log.Println(err)
		conn.Close()
		return
	}
	switch req.Body.(type) {
	case *sarama.MetadataRequest:
		log.Println("metadata request")
	case *sarama.ProduceRequest:
		log.Println("produce request")
	default:
		log.Println("unknown request")
	}
}

type kafkaServerOptions struct{}

var defaultKafkaServerOptions = kafkaServerOptions{}
var globalKafkaServerOptions []KafkaServerOption

type KafkaServerOption interface {
	apply(*kafkaServerOptions)
}

// funcKafkaServerOption wraps a function that modifies kafkaServerOptions into an
// implementation of the KafkaServerOption interface.
type funcKafkaServerOption struct {
	f func(*kafkaServerOptions)
}

func (fdo *funcKafkaServerOption) apply(do *kafkaServerOptions) {
	fdo.f(do)
}

func newFuncKafkaServerOption(f func(*kafkaServerOptions)) *funcKafkaServerOption {
	return &funcKafkaServerOption{
		f: f,
	}
}
