package collector

import (
	"encoding/json"
	"os"
	"time"
)

// MockCollector provides deterministic test data for testing
// It implements the Collector interface
type MockCollector struct {
	connections []Connection
}

// NewMockCollector creates a new mock collector with default test data
func NewMockCollector() *MockCollector {
	return &MockCollector{
		connections: getDefaultTestConnections(),
	}
}

// NewMockCollectorFromFile creates a mock collector from a JSON fixture file
func NewMockCollectorFromFile(filename string) (*MockCollector, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	
	var connections []Connection
	if err := json.Unmarshal(data, &connections); err != nil {
		return nil, err
	}
	
	return &MockCollector{
		connections: connections,
	}, nil
}

// GetConnections returns the mock connections
func (m *MockCollector) GetConnections() ([]Connection, error) {
	// Return a copy to avoid mutation
	result := make([]Connection, len(m.connections))
	copy(result, m.connections)
	return result, nil
}

// AddConnection adds a connection to the mock data
func (m *MockCollector) AddConnection(conn Connection) {
	m.connections = append(m.connections, conn)
}

// SetConnections replaces all connections with the provided slice
func (m *MockCollector) SetConnections(connections []Connection) {
	m.connections = make([]Connection, len(connections))
	copy(m.connections, connections)
}

// SaveToFile saves the current connections to a JSON file
func (m *MockCollector) SaveToFile(filename string) error {
	data, err := json.MarshalIndent(m.connections, "", "  ")
	if err != nil {
		return err
	}
	
	return os.WriteFile(filename, data, 0644)
}

// getDefaultTestConnections returns a set of deterministic test connections
func getDefaultTestConnections() []Connection {
	baseTime := time.Date(2025, 1, 15, 10, 30, 0, 0, time.UTC)
	
	return []Connection{
		{
			TS:        baseTime,
			PID:       1234,
			Process:   "nginx",
			User:      "www-data",
			UID:       33,
			Proto:     "tcp",
			IPVersion: "IPv4",
			State:     "LISTEN",
			Laddr:     "0.0.0.0",
			Lport:     80,
			Raddr:     "*",
			Rport:     0,
			Interface: "eth0",
			RxBytes:   0,
			TxBytes:   0,
			RttMs:     0,
			Mark:      "0x0",
			Namespace: "init",
			Inode:     12345,
		},
		{
			TS:        baseTime.Add(time.Second),
			PID:       1234,
			Process:   "nginx",
			User:      "www-data",
			UID:       33,
			Proto:     "tcp",
			IPVersion: "IPv4",
			State:     "ESTABLISHED",
			Laddr:     "10.0.0.1",
			Lport:     80,
			Raddr:     "203.0.113.10",
			Rport:     52344,
			Interface: "eth0",
			RxBytes:   10240,
			TxBytes:   2048,
			RttMs:     1.7,
			Mark:      "0x0",
			Namespace: "init",
			Inode:     12346,
		},
		{
			TS:        baseTime.Add(2 * time.Second),
			PID:       5678,
			Process:   "postgres",
			User:      "postgres",
			UID:       26,
			Proto:     "tcp",
			IPVersion: "IPv4",
			State:     "LISTEN",
			Laddr:     "127.0.0.1",
			Lport:     5432,
			Raddr:     "*",
			Rport:     0,
			Interface: "lo",
			RxBytes:   0,
			TxBytes:   0,
			RttMs:     0,
			Mark:      "0x0",
			Namespace: "init",
			Inode:     12347,
		},
		{
			TS:        baseTime.Add(3 * time.Second),
			PID:       5678,
			Process:   "postgres",
			User:      "postgres",
			UID:       26,
			Proto:     "tcp",
			IPVersion: "IPv4",
			State:     "ESTABLISHED",
			Laddr:     "127.0.0.1",
			Lport:     5432,
			Raddr:     "127.0.0.1",
			Rport:     45678,
			Interface: "lo",
			RxBytes:   8192,
			TxBytes:   4096,
			RttMs:     0.1,
			Mark:      "0x0",
			Namespace: "init",
			Inode:     12348,
		},
		{
			TS:        baseTime.Add(4 * time.Second),
			PID:       9999,
			Process:   "dns-server",
			User:      "named",
			UID:       25,
			Proto:     "udp",
			IPVersion: "IPv4",
			State:     "LISTEN",
			Laddr:     "0.0.0.0",
			Lport:     53,
			Raddr:     "*",
			Rport:     0,
			Interface: "eth0",
			RxBytes:   1024,
			TxBytes:   512,
			RttMs:     0,
			Mark:      "0x0",
			Namespace: "init",
			Inode:     12349,
		},
		{
			TS:        baseTime.Add(5 * time.Second),
			PID:       1111,
			Process:   "ssh",
			User:      "root",
			UID:       0,
			Proto:     "tcp",
			IPVersion: "IPv4",
			State:     "ESTABLISHED",
			Laddr:     "192.168.1.100",
			Lport:     22,
			Raddr:     "192.168.1.200",
			Rport:     54321,
			Interface: "eth0",
			RxBytes:   2048,
			TxBytes:   1024,
			RttMs:     2.3,
			Mark:      "0x0",
			Namespace: "init",
			Inode:     12350,
		},
		{
			TS:        baseTime.Add(6 * time.Second),
			PID:       2222,
			Process:   "app-server",
			User:      "app",
			UID:       1000,
			Proto:     "unix",
			IPVersion: "",
			State:     "CONNECTED",
			Laddr:     "/tmp/app.sock",
			Lport:     0,
			Raddr:     "",
			Rport:     0,
			Interface: "unix",
			RxBytes:   512,
			TxBytes:   256,
			RttMs:     0,
			Mark:      "",
			Namespace: "init",
			Inode:     12351,
		},
	}
}

// ConnectionBuilder provides a fluent interface for building test connections
type ConnectionBuilder struct {
	conn Connection
}

// NewConnectionBuilder creates a new connection builder with sensible defaults
func NewConnectionBuilder() *ConnectionBuilder {
	return &ConnectionBuilder{
		conn: Connection{
			TS:        time.Now(),
			PID:       1000,
			Process:   "test-process",
			User:      "test-user",
			UID:       1000,
			Proto:     "tcp",
			IPVersion: "IPv4",
			State:     "ESTABLISHED",
			Laddr:     "127.0.0.1",
			Lport:     8080,
			Raddr:     "127.0.0.1",
			Rport:     9090,
			Interface: "lo",
			RxBytes:   1024,
			TxBytes:   512,
			RttMs:     1.0,
			Mark:      "0x0",
			Namespace: "init",
			Inode:     99999,
		},
	}
}

// WithPID sets the PID
func (b *ConnectionBuilder) WithPID(pid int) *ConnectionBuilder {
	b.conn.PID = pid
	return b
}

// WithProcess sets the process name
func (b *ConnectionBuilder) WithProcess(process string) *ConnectionBuilder {
	b.conn.Process = process
	return b
}

// WithProto sets the protocol
func (b *ConnectionBuilder) WithProto(proto string) *ConnectionBuilder {
	b.conn.Proto = proto
	return b
}

// WithState sets the connection state
func (b *ConnectionBuilder) WithState(state string) *ConnectionBuilder {
	b.conn.State = state
	return b
}

// WithLocalAddr sets the local address and port
func (b *ConnectionBuilder) WithLocalAddr(addr string, port int) *ConnectionBuilder {
	b.conn.Laddr = addr
	b.conn.Lport = port
	return b
}

// WithRemoteAddr sets the remote address and port
func (b *ConnectionBuilder) WithRemoteAddr(addr string, port int) *ConnectionBuilder {
	b.conn.Raddr = addr
	b.conn.Rport = port
	return b
}

// WithInterface sets the network interface
func (b *ConnectionBuilder) WithInterface(iface string) *ConnectionBuilder {
	b.conn.Interface = iface
	return b
}

// WithBytes sets the rx and tx byte counts
func (b *ConnectionBuilder) WithBytes(rx, tx int64) *ConnectionBuilder {
	b.conn.RxBytes = rx
	b.conn.TxBytes = tx
	return b
}

// Build returns the constructed connection
func (b *ConnectionBuilder) Build() Connection {
	return b.conn
}

// TestFixture provides test scenarios for different use cases
type TestFixture struct {
	Name        string
	Description string
	Connections []Connection
}

// GetTestFixtures returns predefined test fixtures for different scenarios
func GetTestFixtures() []TestFixture {
	baseTime := time.Date(2025, 1, 15, 10, 30, 0, 0, time.UTC)
	
	return []TestFixture{
		{
			Name:        "empty",
			Description: "No connections",
			Connections: []Connection{},
		},
		{
			Name:        "single-tcp",
			Description: "Single TCP connection",
			Connections: []Connection{
				NewConnectionBuilder().
					WithPID(1234).
					WithProcess("test-app").
					WithProto("tcp").
					WithState("ESTABLISHED").
					WithLocalAddr("127.0.0.1", 8080).
					WithRemoteAddr("127.0.0.1", 9090).
					Build(),
			},
		},
		{
			Name:        "mixed-protocols",
			Description: "Mix of TCP, UDP, and Unix sockets",
			Connections: []Connection{
				{
					TS:        baseTime,
					PID:       1,
					Process:   "tcp-server",
					Proto:     "tcp",
					State:     "LISTEN",
					Laddr:     "0.0.0.0",
					Lport:     80,
					Interface: "eth0",
				},
				{
					TS:        baseTime.Add(time.Second),
					PID:       2,
					Process:   "udp-server",
					Proto:     "udp",
					State:     "LISTEN",
					Laddr:     "0.0.0.0",
					Lport:     53,
					Interface: "eth0",
				},
				{
					TS:        baseTime.Add(2 * time.Second),
					PID:       3,
					Process:   "unix-app",
					Proto:     "unix",
					State:     "CONNECTED",
					Laddr:     "/tmp/test.sock",
					Interface: "unix",
				},
			},
		},
		{
			Name:        "high-volume",
			Description: "Large number of connections for performance testing",
			Connections: generateHighVolumeConnections(1000),
		},
	}
}

// generateHighVolumeConnections creates a large number of test connections
func generateHighVolumeConnections(count int) []Connection {
	connections := make([]Connection, count)
	baseTime := time.Date(2025, 1, 15, 10, 30, 0, 0, time.UTC)
	
	for i := 0; i < count; i++ {
		connections[i] = NewConnectionBuilder().
			WithPID(1000 + i).
			WithProcess("worker-" + string(rune('a'+i%26))).
			WithProto([]string{"tcp", "udp"}[i%2]).
			WithState([]string{"ESTABLISHED", "LISTEN", "TIME_WAIT"}[i%3]).
			WithLocalAddr("127.0.0.1", 8000+i%1000).
			WithRemoteAddr("10.0.0."+string(rune('1'+i%10)), 9000+i%1000).
			Build()
		connections[i].TS = baseTime.Add(time.Duration(i) * time.Millisecond)
	}
	
	return connections
}