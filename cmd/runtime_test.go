package cmd

import (
	"testing"

	"github.com/karol-broda/snitch/internal/collector"
)

func TestParseFilterArgs_Empty(t *testing.T) {
	filters, err := ParseFilterArgs([]string{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if filters.Proto != "" {
		t.Errorf("expected empty proto, got %q", filters.Proto)
	}
}

func TestParseFilterArgs_Proto(t *testing.T) {
	filters, err := ParseFilterArgs([]string{"proto=tcp"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if filters.Proto != "tcp" {
		t.Errorf("expected proto 'tcp', got %q", filters.Proto)
	}
}

func TestParseFilterArgs_State(t *testing.T) {
	filters, err := ParseFilterArgs([]string{"state=LISTEN"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if filters.State != "LISTEN" {
		t.Errorf("expected state 'LISTEN', got %q", filters.State)
	}
}

func TestParseFilterArgs_PID(t *testing.T) {
	filters, err := ParseFilterArgs([]string{"pid=1234"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if filters.Pid != 1234 {
		t.Errorf("expected pid 1234, got %d", filters.Pid)
	}
}

func TestParseFilterArgs_InvalidPID(t *testing.T) {
	_, err := ParseFilterArgs([]string{"pid=notanumber"})
	if err == nil {
		t.Error("expected error for invalid pid")
	}
}

func TestParseFilterArgs_Proc(t *testing.T) {
	filters, err := ParseFilterArgs([]string{"proc=nginx"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if filters.Proc != "nginx" {
		t.Errorf("expected proc 'nginx', got %q", filters.Proc)
	}
}

func TestParseFilterArgs_Lport(t *testing.T) {
	filters, err := ParseFilterArgs([]string{"lport=80"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if filters.Lport != 80 {
		t.Errorf("expected lport 80, got %d", filters.Lport)
	}
}

func TestParseFilterArgs_InvalidLport(t *testing.T) {
	_, err := ParseFilterArgs([]string{"lport=notaport"})
	if err == nil {
		t.Error("expected error for invalid lport")
	}
}

func TestParseFilterArgs_Rport(t *testing.T) {
	filters, err := ParseFilterArgs([]string{"rport=443"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if filters.Rport != 443 {
		t.Errorf("expected rport 443, got %d", filters.Rport)
	}
}

func TestParseFilterArgs_InvalidRport(t *testing.T) {
	_, err := ParseFilterArgs([]string{"rport=invalid"})
	if err == nil {
		t.Error("expected error for invalid rport")
	}
}

func TestParseFilterArgs_UserByName(t *testing.T) {
	filters, err := ParseFilterArgs([]string{"user=root"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if filters.User != "root" {
		t.Errorf("expected user 'root', got %q", filters.User)
	}
}

func TestParseFilterArgs_UserByUID(t *testing.T) {
	filters, err := ParseFilterArgs([]string{"user=1000"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if filters.UID != 1000 {
		t.Errorf("expected uid 1000, got %d", filters.UID)
	}
}

func TestParseFilterArgs_Laddr(t *testing.T) {
	filters, err := ParseFilterArgs([]string{"laddr=127.0.0.1"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if filters.Laddr != "127.0.0.1" {
		t.Errorf("expected laddr '127.0.0.1', got %q", filters.Laddr)
	}
}

func TestParseFilterArgs_Raddr(t *testing.T) {
	filters, err := ParseFilterArgs([]string{"raddr=8.8.8.8"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if filters.Raddr != "8.8.8.8" {
		t.Errorf("expected raddr '8.8.8.8', got %q", filters.Raddr)
	}
}

func TestParseFilterArgs_Contains(t *testing.T) {
	filters, err := ParseFilterArgs([]string{"contains=google"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if filters.Contains != "google" {
		t.Errorf("expected contains 'google', got %q", filters.Contains)
	}
}

func TestParseFilterArgs_Interface(t *testing.T) {
	filters, err := ParseFilterArgs([]string{"if=eth0"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if filters.Interface != "eth0" {
		t.Errorf("expected interface 'eth0', got %q", filters.Interface)
	}

	// test alternative syntax
	filters2, err := ParseFilterArgs([]string{"interface=lo"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if filters2.Interface != "lo" {
		t.Errorf("expected interface 'lo', got %q", filters2.Interface)
	}
}

func TestParseFilterArgs_Mark(t *testing.T) {
	filters, err := ParseFilterArgs([]string{"mark=0x1234"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if filters.Mark != "0x1234" {
		t.Errorf("expected mark '0x1234', got %q", filters.Mark)
	}
}

func TestParseFilterArgs_Namespace(t *testing.T) {
	filters, err := ParseFilterArgs([]string{"namespace=default"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if filters.Namespace != "default" {
		t.Errorf("expected namespace 'default', got %q", filters.Namespace)
	}
}

func TestParseFilterArgs_Inode(t *testing.T) {
	filters, err := ParseFilterArgs([]string{"inode=123456"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if filters.Inode != 123456 {
		t.Errorf("expected inode 123456, got %d", filters.Inode)
	}
}

func TestParseFilterArgs_InvalidInode(t *testing.T) {
	_, err := ParseFilterArgs([]string{"inode=notanumber"})
	if err == nil {
		t.Error("expected error for invalid inode")
	}
}

func TestParseFilterArgs_Multiple(t *testing.T) {
	filters, err := ParseFilterArgs([]string{"proto=tcp", "state=LISTEN", "lport=80"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if filters.Proto != "tcp" {
		t.Errorf("expected proto 'tcp', got %q", filters.Proto)
	}
	if filters.State != "LISTEN" {
		t.Errorf("expected state 'LISTEN', got %q", filters.State)
	}
	if filters.Lport != 80 {
		t.Errorf("expected lport 80, got %d", filters.Lport)
	}
}

func TestParseFilterArgs_InvalidFormat(t *testing.T) {
	_, err := ParseFilterArgs([]string{"invalidformat"})
	if err == nil {
		t.Error("expected error for invalid format")
	}
}

func TestParseFilterArgs_UnknownKey(t *testing.T) {
	_, err := ParseFilterArgs([]string{"unknownkey=value"})
	if err == nil {
		t.Error("expected error for unknown key")
	}
}

func TestParseFilterArgs_CaseInsensitiveKeys(t *testing.T) {
	filters, err := ParseFilterArgs([]string{"PROTO=tcp", "State=LISTEN"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if filters.Proto != "tcp" {
		t.Errorf("expected proto 'tcp', got %q", filters.Proto)
	}
	if filters.State != "LISTEN" {
		t.Errorf("expected state 'LISTEN', got %q", filters.State)
	}
}

func TestBuildFilters_TCPOnly(t *testing.T) {
	// save and restore global flags
	oldTCP, oldUDP := filterTCP, filterUDP
	defer func() {
		filterTCP, filterUDP = oldTCP, oldUDP
	}()

	filterTCP = true
	filterUDP = false

	filters, err := BuildFilters([]string{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if filters.Proto != "tcp" {
		t.Errorf("expected proto 'tcp', got %q", filters.Proto)
	}
}

func TestBuildFilters_UDPOnly(t *testing.T) {
	oldTCP, oldUDP := filterTCP, filterUDP
	defer func() {
		filterTCP, filterUDP = oldTCP, oldUDP
	}()

	filterTCP = false
	filterUDP = true

	filters, err := BuildFilters([]string{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if filters.Proto != "udp" {
		t.Errorf("expected proto 'udp', got %q", filters.Proto)
	}
}

func TestBuildFilters_ListenOnly(t *testing.T) {
	oldListen, oldEstab := filterListen, filterEstab
	defer func() {
		filterListen, filterEstab = oldListen, oldEstab
	}()

	filterListen = true
	filterEstab = false

	filters, err := BuildFilters([]string{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if filters.State != "LISTEN" {
		t.Errorf("expected state 'LISTEN', got %q", filters.State)
	}
}

func TestBuildFilters_EstablishedOnly(t *testing.T) {
	oldListen, oldEstab := filterListen, filterEstab
	defer func() {
		filterListen, filterEstab = oldListen, oldEstab
	}()

	filterListen = false
	filterEstab = true

	filters, err := BuildFilters([]string{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if filters.State != "ESTABLISHED" {
		t.Errorf("expected state 'ESTABLISHED', got %q", filters.State)
	}
}

func TestBuildFilters_IPv4Flag(t *testing.T) {
	oldIPv4 := filterIPv4
	defer func() {
		filterIPv4 = oldIPv4
	}()

	filterIPv4 = true

	filters, err := BuildFilters([]string{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !filters.IPv4 {
		t.Error("expected IPv4 to be true")
	}
}

func TestBuildFilters_IPv6Flag(t *testing.T) {
	oldIPv6 := filterIPv6
	defer func() {
		filterIPv6 = oldIPv6
	}()

	filterIPv6 = true

	filters, err := BuildFilters([]string{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !filters.IPv6 {
		t.Error("expected IPv6 to be true")
	}
}

func TestBuildFilters_CombinedArgsAndFlags(t *testing.T) {
	oldTCP := filterTCP
	defer func() {
		filterTCP = oldTCP
	}()

	filterTCP = true

	filters, err := BuildFilters([]string{"lport=80"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if filters.Proto != "tcp" {
		t.Errorf("expected proto 'tcp', got %q", filters.Proto)
	}
	if filters.Lport != 80 {
		t.Errorf("expected lport 80, got %d", filters.Lport)
	}
}

func TestRuntime_PreWarmDNS(t *testing.T) {
	rt := &Runtime{
		Connections: []collector.Connection{
			{Laddr: "127.0.0.1", Raddr: "192.168.1.1"},
			{Laddr: "127.0.0.1", Raddr: "10.0.0.1"},
		},
	}

	// should not panic
	rt.PreWarmDNS()
}

func TestRuntime_PreWarmDNS_Empty(t *testing.T) {
	rt := &Runtime{
		Connections: []collector.Connection{},
	}

	// should not panic with empty connections
	rt.PreWarmDNS()
}

func TestRuntime_SortConnections(t *testing.T) {
	rt := &Runtime{
		Connections: []collector.Connection{
			{Lport: 443},
			{Lport: 80},
			{Lport: 8080},
		},
	}

	rt.SortConnections(collector.SortOptions{
		Field:     collector.SortByLport,
		Direction: collector.SortAsc,
	})

	if rt.Connections[0].Lport != 80 {
		t.Errorf("expected first connection to have lport 80, got %d", rt.Connections[0].Lport)
	}
	if rt.Connections[1].Lport != 443 {
		t.Errorf("expected second connection to have lport 443, got %d", rt.Connections[1].Lport)
	}
	if rt.Connections[2].Lport != 8080 {
		t.Errorf("expected third connection to have lport 8080, got %d", rt.Connections[2].Lport)
	}
}

func TestRuntime_SortConnections_Desc(t *testing.T) {
	rt := &Runtime{
		Connections: []collector.Connection{
			{Lport: 80},
			{Lport: 443},
			{Lport: 8080},
		},
	}

	rt.SortConnections(collector.SortOptions{
		Field:     collector.SortByLport,
		Direction: collector.SortDesc,
	})

	if rt.Connections[0].Lport != 8080 {
		t.Errorf("expected first connection to have lport 8080, got %d", rt.Connections[0].Lport)
	}
}

func TestApplyFilter_AllKeys(t *testing.T) {
	tests := []struct {
		key      string
		value    string
		validate func(t *testing.T, f *collector.FilterOptions)
	}{
		{"proto", "tcp", func(t *testing.T, f *collector.FilterOptions) {
			if f.Proto != "tcp" {
				t.Errorf("proto: expected 'tcp', got %q", f.Proto)
			}
		}},
		{"state", "LISTEN", func(t *testing.T, f *collector.FilterOptions) {
			if f.State != "LISTEN" {
				t.Errorf("state: expected 'LISTEN', got %q", f.State)
			}
		}},
		{"pid", "100", func(t *testing.T, f *collector.FilterOptions) {
			if f.Pid != 100 {
				t.Errorf("pid: expected 100, got %d", f.Pid)
			}
		}},
		{"proc", "nginx", func(t *testing.T, f *collector.FilterOptions) {
			if f.Proc != "nginx" {
				t.Errorf("proc: expected 'nginx', got %q", f.Proc)
			}
		}},
		{"lport", "80", func(t *testing.T, f *collector.FilterOptions) {
			if f.Lport != 80 {
				t.Errorf("lport: expected 80, got %d", f.Lport)
			}
		}},
		{"rport", "443", func(t *testing.T, f *collector.FilterOptions) {
			if f.Rport != 443 {
				t.Errorf("rport: expected 443, got %d", f.Rport)
			}
		}},
		{"laddr", "127.0.0.1", func(t *testing.T, f *collector.FilterOptions) {
			if f.Laddr != "127.0.0.1" {
				t.Errorf("laddr: expected '127.0.0.1', got %q", f.Laddr)
			}
		}},
		{"raddr", "8.8.8.8", func(t *testing.T, f *collector.FilterOptions) {
			if f.Raddr != "8.8.8.8" {
				t.Errorf("raddr: expected '8.8.8.8', got %q", f.Raddr)
			}
		}},
		{"contains", "test", func(t *testing.T, f *collector.FilterOptions) {
			if f.Contains != "test" {
				t.Errorf("contains: expected 'test', got %q", f.Contains)
			}
		}},
		{"if", "eth0", func(t *testing.T, f *collector.FilterOptions) {
			if f.Interface != "eth0" {
				t.Errorf("interface: expected 'eth0', got %q", f.Interface)
			}
		}},
		{"mark", "0xff", func(t *testing.T, f *collector.FilterOptions) {
			if f.Mark != "0xff" {
				t.Errorf("mark: expected '0xff', got %q", f.Mark)
			}
		}},
		{"namespace", "ns1", func(t *testing.T, f *collector.FilterOptions) {
			if f.Namespace != "ns1" {
				t.Errorf("namespace: expected 'ns1', got %q", f.Namespace)
			}
		}},
		{"inode", "12345", func(t *testing.T, f *collector.FilterOptions) {
			if f.Inode != 12345 {
				t.Errorf("inode: expected 12345, got %d", f.Inode)
			}
		}},
	}

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			filters := &collector.FilterOptions{}
			err := applyFilter(filters, tt.key, tt.value)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			tt.validate(t, filters)
		})
	}
}

