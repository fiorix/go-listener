package fastopen

// SocketOption is the TCP fast open socket option on Linux.
//
// Make sure OS level support for fastopen is enabled:
//	sysctl -w net.ipv4.tcp_fastopen=2
const SocketOption = 23
