package sarama

import (
	"context"
	"fmt"
	"github.com/MadsRC/sarama"
	pb "github.com/madsrc/webway/gen/go/webway/v1"
	"log"
	"net"
	"strings"
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

	if s.opts.metadataStoreGrpcClient == nil {
		return nil, fmt.Errorf("metadataStoreGrpcClient is required")
	}

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
		c := s.newConn(conn)
		go func() {
			c.Serve(context.Background())
		}()
	}
}

func (s *KafkaServer) newConn(rwc net.Conn) *conn {
	c := &conn{
		rwc:    rwc,
		server: s,
	}
	return c
}

type kafkaServerOptions struct {
	metadataStoreGrpcClient pb.MetadataStoreClient
}

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

func WithMetadataStoreGrpcClient(c pb.MetadataStoreClient) KafkaServerOption {
	return newFuncKafkaServerOption(func(o *kafkaServerOptions) {
		o.metadataStoreGrpcClient = c
	})
}

type conn struct {
	// rwc is the underlying [net.Conn].
	rwc net.Conn
	// server is the server on which the connection arrived.
	server *KafkaServer
}

func (c *conn) Serve(ctx context.Context) {
	req, _, err := sarama.DecodeRequest(c.rwc)
	if err != nil {
		log.Println(err)
		c.rwc.Close()
		return
	}
	switch req.Body.(type) {
	case *sarama.MetadataRequest:
		c.handleMetadataRequest(ctx, req.CorrelationID, req.ClientID, req.Body.(*sarama.MetadataRequest))
	case *sarama.ProduceRequest:
		log.Println("produce request")
	default:
		log.Println("unknown request")
	}
}

func (c *conn) handleMetadataRequest(ctx context.Context, corID int32, clientID string, req *sarama.MetadataRequest) error {
	log.Println("metadata request")
	md, err := c.server.opts.metadataStoreGrpcClient.GetMetadata(ctx, &pb.GetMetadataRequest{
		AvailabilityZone: extractAz(clientID),
	})
	if err != nil {
		return err
	}

	log.Printf("upstream metadata response: %+v", md)

	resp := &sarama.MetadataResponse{
		Version:      req.Version,
		ClusterID:    &md.ClusterId,
		ControllerID: 0,
	}

	for _, agent := range md.Agents {
		log.Printf("adding broker: %v", agent.Id)
		resp.AddBroker(fmt.Sprintf("%s:%d", agent.Hostname, agent.Port), agent.Id)
	}

	for _, topic := range md.Topics {
		log.Printf("adding topic: %v", topic.Name)
		resp.AddTopic(topic.Name, 0)

		for _, partition := range topic.Partitions {
			log.Printf("adding partition: %v", partition.Id)
			resp.AddTopicPartition(topic.Name, partition.Id, 0, []int32{0}, []int32{1}, []int32{0}, 0)
		}
	}

	b, err := sarama.Encode(resp, nil)
	if err != nil {
		return err
	}

	log.Println("metadata response")

	// convert req.CorrelationID to uint32
	b = append([]byte{byte(corID >> 24), byte(corID >> 16), byte(corID >> 8), byte(corID)}, b...)
	msgSize := len(b)
	log.Printf("Message size: %v", msgSize)
	// convert msgSize to big endian uint32
	b = append([]byte{byte(msgSize >> 24), byte(msgSize >> 16), byte(msgSize >> 8), byte(msgSize)}, b...)

	n, err := c.rwc.Write(b)
	if err != nil {
		return err
	}

	log.Printf("Wrote %v bytes", n)

	return nil
}

func extractAz(clientID string) string {
	vals := strings.Split(clientID, ",")
	for _, v := range vals {
		if strings.HasPrefix(v, "webway_az=") {
			return strings.TrimPrefix(v, "az=")
		}
	}
	return ""
}
