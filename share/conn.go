package chshare

import (
	"fmt"
	"io"
	"sync"
	"sync/atomic"
)

// ChannelConn is a virtual open "socket", either
//      1) created by a ChannelEndpoint to wrap communication with a local network resource
//      2) created by the proxy session to wrap a single ssh communication channel with a remote endpoint
type ChannelConn interface {
	io.ReadWriteCloser

	// WaitForClose blocks until the Close() method has been called and completed. The error returned
	// from the first Close() is returned
	WaitForClose() error

	// CloseWrite shuts down the writing side of the "socket". Corresponds to net.TCPConn.CloseWrite().
	// this method is called when end-of-stream is reached reading from the other ChannelConn of a pair
	// pair are connected via a ChannelPipe. It allows for protocols like HTTP 1.0 in which a client
	// sends a request, closes the write side of the socket, then reads the response, and a server reads
	// a request until end-of-stream before sending a response.
	CloseWrite() error

	// GetNumBytesRead returns the number of bytes read so far on a ChannelConn
	GetNumBytesRead() int64

	// GetNumBytesWritten returns the number of bytes written so far on a ChannelConn
	GetNumBytesWritten() int64
}

var nextBasicConnID int32

// AllocBasicConnID allocates a unique ChannelConn ID number, for logging purposes
func AllocBasicConnID() int32 {
	return atomic.AddInt32(&nextBasicConnID, 1)
}

// BasicConn is a base common implementation for local ChannelConn
type BasicConn struct {
	ChannelConn
	*Logger
	ID              int32
	Strname         string
	Lock            sync.Mutex
	Done            chan struct{}
	CloseOnce       sync.Once
	CloseErr        error
	NumBytesRead    int64
	NumBytesWritten int64
}

// InitBasicConn initializes the BasicConn portion of a new connection object
func (c *BasicConn) InitBasicConn(logger *Logger, namef string, args ...interface{}) {
	c.ID = AllocBasicConnID()
	c.Strname = fmt.Sprintf("[%d]", c.ID) + fmt.Sprintf(namef, args...)
	c.Logger = logger.Fork(c.Strname)
	c.Done = make(chan struct{})
}

// GetNumBytesRead returns the number of bytes read so far on a ChannelConn
func (c *BasicConn) GetNumBytesRead() int64 {
	return atomic.LoadInt64(&c.NumBytesRead)
}

// GetNumBytesWritten returns the number of bytes written so far on a ChannelConn
func (c *BasicConn) GetNumBytesWritten() int64 {
	return atomic.LoadInt64(&c.NumBytesWritten)
}

// GetNumBytesWritten returns the number of bytes written so far on a ChannelConn
func (c *BasicConn) String() string {
	return c.Strname
}
