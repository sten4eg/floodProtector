package floodProtecor

import (
	"github.com/puzpuzpuz/xsync/v3"
	"math/rand/v2"
	"net"
	"testing"
	"time"
)

type mockTCPAcceptor struct {
	connections chan *MockTCPConn
	err         error
}

type MockTCPConn struct {
	addr *net.TCPAddr
}

func (m *MockTCPConn) Read(b []byte) (n int, err error) {
	return 0, nil
}

func (m *MockTCPConn) Write(b []byte) (n int, err error) {
	return 0, nil
}

func (m *MockTCPConn) Close() error {
	return nil
}

func (m *MockTCPConn) LocalAddr() net.Addr {
	return nil
}

func (m *MockTCPConn) RemoteAddr() net.Addr {
	return m.addr
}

func (m *MockTCPConn) SetDeadline(t time.Time) error {
	return nil
}

func (m *MockTCPConn) SetReadDeadline(t time.Time) error {
	return nil
}

func (m *MockTCPConn) SetWriteDeadline(t time.Time) error {
	return nil
}

func (m *mockTCPAcceptor) AcceptTCP() (net.Conn, error) {
	if m.err != nil {
		return nil, m.err
	}
	select {
	case conn := <-m.connections:

		return conn, nil
	}
}
func BenchmarkAcceptTCPWithFSM(b *testing.B) {
	floodProtection := xsync.NewMapOf[string, connectionInfo]()

	acceptor := &mockTCPAcceptor{
		connections: make(chan *MockTCPConn, 10000),
	}

	for i := 0; i < 10000; i++ {
		ip := net.IPv4(byte(rand.Uint()%254), byte(rand.Uint()%254), byte(0), byte(rand.Uint()%254)) // Несколько IP для симуляции нагрузки
		mockAddr := &net.TCPAddr{IP: ip, Port: int(rand.Uint() % 10000)}
		acceptor.connections <- &MockTCPConn{addr: mockAddr}
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		cn, e := AcceptTCP(acceptor, floodProtection)
		if e != nil {
			// если старое соединение закрылось то мы открываем новое
			Add(acceptor)
			continue
		}
		// а тут просто пихаем соединение назад в очередь
		acceptor.connections <- (cn).(*MockTCPConn)
	}

}

func Add(tc *mockTCPAcceptor) {
	ip := net.IPv4(byte(rand.Uint()%254), byte(rand.Uint()%254), byte(0), byte(rand.Uint()%254))
	mockAddr := &net.TCPAddr{IP: ip, Port: int(rand.Uint() % 10000)}
	tc.connections <- &MockTCPConn{addr: mockAddr}
}
