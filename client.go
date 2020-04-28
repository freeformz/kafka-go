package kafka

import (
	"context"
	"errors"
	"net"
	"time"

	"github.com/segmentio/kafka-go/protocol"
)

// Client is a high-level API to interract with kafka brokers.
//
// All methods of the Client type accept a context as first argument, which may
// be used to asynchronously cancel the requests.
//
// Clients are safe to use concurrently from multiple goroutines, as long as
// their configuration is not changed after it was first used.
type Client struct {
	// Address of the kafka cluster (or specific broker) that the client will be
	// sending requests to.
	//
	// This field is optional, the address may be provided in each request
	// instead. The request address takes precedence if both were specified.
	Addr net.Addr

	// Time limit for requests sent by this client.
	//
	// If zero, no timeout is applied.
	Timeout time.Duration

	// A transport used to communicate with the kafka brokers.
	//
	// If nil, DefaultTransport is used.
	Transport RoundTripper
}

func (c *Client) roundTrip(ctx context.Context, addr net.Addr, msg protocol.Message) (protocol.Message, error) {
	if c.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, c.Timeout)
		defer cancel()
	}

	if addr == nil {
		if addr = c.Addr; addr == nil {
			return nil, errNoAddr
		}
	}

	return c.transport().RoundTrip(ctx, addr, msg)
}

func (c *Client) transport() RoundTripper {
	if c.Transport != nil {
		return c.Transport
	}
	return DefaultTransport
}

func (c *Client) timeout(ctx context.Context) time.Duration {
	timeout := c.Timeout

	if deadline, ok := ctx.Deadline(); ok {
		if remain := time.Until(deadline); remain < timeout {
			timeout = remain
		}
	}

	buffer := timeout / 4
	if buffer > time.Second {
		buffer = time.Second
	}

	return timeout - buffer
}

func (c *Client) timeoutMs(ctx context.Context) int32 {
	return milliseconds(c.timeout(ctx))
}

var (
	errNoAddr = errors.New("no address was given for the kafka cluster in the request or on the client")
)