package core

import (
	"context"
	"net"
)

type ContainerConnection interface {
	ResizeTTY(ctx context.Context, width, height uint) error
	ContainerID() string
	Conn() net.Conn
	Close() error
}
