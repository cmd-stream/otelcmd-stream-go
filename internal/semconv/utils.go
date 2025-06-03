package semconv

import (
	"net"
	"strconv"

	"go.opentelemetry.io/otel"
)

func protoName(addr net.Addr) string {
	return addr.Network()
}

func addressPort(addr net.Addr) (address string, port int) {
	address, p, err := net.SplitHostPort(addr.String())
	if err != nil {
		return "undefined", 0
	}
	port, err = strconv.Atoi(p)
	if err != nil {
		port = 0
	}
	return address, port
}

func handleErr(err error) {
	if err != nil {
		otel.Handle(err)
	}
}
