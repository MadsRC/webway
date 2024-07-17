package main

import (
	"github.com/MadsRC/sarama"
	"log"
	"net"
)

func main() {
	// Listen on TCP port 2000 on all available unicast and
	// anycast IP addresses of the local system.
	l, err := net.Listen("tcp", ":2000")
	if err != nil {
		log.Fatal(err)
	}
	defer l.Close()
	for {
		// Wait for a connection.
		conn, err := l.Accept()
		if err != nil {
			log.Fatal(err)
		}

		// Handle the connection in a new goroutine.
		// The loop then returns to accepting, so that
		// multiple connections may be served concurrently.

		go handleConn(conn)
	}
}

func handleConn(conn net.Conn) {
	req, n, err := sarama.DecodeRequest(conn)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Received request: %v", req)
	log.Printf("Received integer: %v", n)

	log.Printf("Request body: %+v", req.Body)

	metadataReq := req.Body.(*sarama.MetadataRequest)
	log.Printf("MetadataRequest: %+v", metadataReq)

	clusterID := "test"

	metadataRes := &sarama.MetadataResponse{
		Version:      metadataReq.Version,
		ClusterID:    &clusterID,
		ControllerID: 1,
	}
	metadataRes.AddBroker("127.0.0.1:2000", 1)
	metadataRes.AddTopic("test", 0)
	metadataRes.AddTopicPartition("test", 0, 1, []int32{0}, []int32{1}, []int32{0}, 0)

	log.Printf("Responding with %+v", metadataRes)

	b, err := sarama.Encode(metadataRes, nil)
	if err != nil {
		log.Fatal(err)
	}
	// convert req.CorrelationID to big endian uint32
	b = append([]byte{byte(req.CorrelationID >> 24), byte(req.CorrelationID >> 16), byte(req.CorrelationID >> 8), byte(req.CorrelationID)}, b...)
	msgSize := len(b)
	log.Printf("Message size: %v", msgSize)
	// convert msgSize to big endian uint32
	b = append([]byte{byte(msgSize >> 24), byte(msgSize >> 16), byte(msgSize >> 8), byte(msgSize)}, b...)

	conn.Write(b)

	// Shut down the connection.
	conn.Close()
}
