package collector

import (
	"testing"
	"time"
)

func TestSortConnections(t *testing.T) {
	conns := []Connection{
		{PID: 3, Process: "nginx", Lport: 80, State: "ESTABLISHED"},
		{PID: 1, Process: "sshd", Lport: 22, State: "LISTEN"},
		{PID: 2, Process: "postgres", Lport: 5432, State: "LISTEN"},
	}

	t.Run("sort by PID ascending", func(t *testing.T) {
		c := make([]Connection, len(conns))
		copy(c, conns)

		SortConnections(c, SortOptions{Field: SortByPID, Direction: SortAsc})

		if c[0].PID != 1 || c[1].PID != 2 || c[2].PID != 3 {
			t.Errorf("expected PIDs [1,2,3], got [%d,%d,%d]", c[0].PID, c[1].PID, c[2].PID)
		}
	})

	t.Run("sort by PID descending", func(t *testing.T) {
		c := make([]Connection, len(conns))
		copy(c, conns)

		SortConnections(c, SortOptions{Field: SortByPID, Direction: SortDesc})

		if c[0].PID != 3 || c[1].PID != 2 || c[2].PID != 1 {
			t.Errorf("expected PIDs [3,2,1], got [%d,%d,%d]", c[0].PID, c[1].PID, c[2].PID)
		}
	})

	t.Run("sort by port ascending", func(t *testing.T) {
		c := make([]Connection, len(conns))
		copy(c, conns)

		SortConnections(c, SortOptions{Field: SortByLport, Direction: SortAsc})

		if c[0].Lport != 22 || c[1].Lport != 80 || c[2].Lport != 5432 {
			t.Errorf("expected ports [22,80,5432], got [%d,%d,%d]", c[0].Lport, c[1].Lport, c[2].Lport)
		}
	})

	t.Run("sort by state puts LISTEN first", func(t *testing.T) {
		c := make([]Connection, len(conns))
		copy(c, conns)

		SortConnections(c, SortOptions{Field: SortByState, Direction: SortAsc})

		if c[0].State != "LISTEN" || c[1].State != "LISTEN" || c[2].State != "ESTABLISHED" {
			t.Errorf("expected LISTEN states first, got [%s,%s,%s]", c[0].State, c[1].State, c[2].State)
		}
	})

	t.Run("sort by process case insensitive", func(t *testing.T) {
		c := []Connection{
			{Process: "Nginx"},
			{Process: "apache"},
			{Process: "SSHD"},
		}

		SortConnections(c, SortOptions{Field: SortByProcess, Direction: SortAsc})

		if c[0].Process != "apache" {
			t.Errorf("expected apache first, got %s", c[0].Process)
		}
	})
}

func TestParseSortOptions(t *testing.T) {
	tests := []struct {
		input     string
		wantField SortField
		wantDir   SortDirection
	}{
		{"pid", SortByPID, SortAsc},
		{"pid:asc", SortByPID, SortAsc},
		{"pid:desc", SortByPID, SortDesc},
		{"lport", SortByLport, SortAsc},
		{"LPORT:DESC", SortByLport, SortDesc},
		{"", SortByLport, SortAsc}, // default
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			opts := ParseSortOptions(tt.input)
			if opts.Field != tt.wantField {
				t.Errorf("field: got %v, want %v", opts.Field, tt.wantField)
			}
			if opts.Direction != tt.wantDir {
				t.Errorf("direction: got %v, want %v", opts.Direction, tt.wantDir)
			}
		})
	}
}

func TestStateOrder(t *testing.T) {
	if stateOrder("LISTEN") >= stateOrder("ESTABLISHED") {
		t.Error("LISTEN should come before ESTABLISHED")
	}
	if stateOrder("ESTABLISHED") >= stateOrder("TIME_WAIT") {
		t.Error("ESTABLISHED should come before TIME_WAIT")
	}
	if stateOrder("UNKNOWN") != 99 {
		t.Error("unknown states should return 99")
	}
}

func TestSortByTimestamp(t *testing.T) {
	now := time.Now()
	conns := []Connection{
		{TS: now.Add(2 * time.Second)},
		{TS: now},
		{TS: now.Add(1 * time.Second)},
	}

	SortConnections(conns, SortOptions{Field: SortByTimestamp, Direction: SortAsc})

	if !conns[0].TS.Equal(now) {
		t.Error("oldest timestamp should be first")
	}
	if !conns[2].TS.Equal(now.Add(2 * time.Second)) {
		t.Error("newest timestamp should be last")
	}
}

func TestSortByRemoteAddr(t *testing.T) {
	conns := []Connection{
		{Raddr: "192.168.1.100", Rport: 443},
		{Raddr: "10.0.0.1", Rport: 80},
		{Raddr: "172.16.0.50", Rport: 8080},
	}

	t.Run("sort by raddr ascending", func(t *testing.T) {
		c := make([]Connection, len(conns))
		copy(c, conns)

		SortConnections(c, SortOptions{Field: SortByRaddr, Direction: SortAsc})

		if c[0].Raddr != "10.0.0.1" {
			t.Errorf("expected '10.0.0.1' first, got '%s'", c[0].Raddr)
		}
		if c[1].Raddr != "172.16.0.50" {
			t.Errorf("expected '172.16.0.50' second, got '%s'", c[1].Raddr)
		}
		if c[2].Raddr != "192.168.1.100" {
			t.Errorf("expected '192.168.1.100' last, got '%s'", c[2].Raddr)
		}
	})

	t.Run("sort by raddr descending", func(t *testing.T) {
		c := make([]Connection, len(conns))
		copy(c, conns)

		SortConnections(c, SortOptions{Field: SortByRaddr, Direction: SortDesc})

		if c[0].Raddr != "192.168.1.100" {
			t.Errorf("expected '192.168.1.100' first, got '%s'", c[0].Raddr)
		}
	})
}

func TestSortByRemotePort(t *testing.T) {
	conns := []Connection{
		{Raddr: "192.168.1.1", Rport: 443},
		{Raddr: "192.168.1.2", Rport: 80},
		{Raddr: "192.168.1.3", Rport: 8080},
	}

	t.Run("sort by rport ascending", func(t *testing.T) {
		c := make([]Connection, len(conns))
		copy(c, conns)

		SortConnections(c, SortOptions{Field: SortByRport, Direction: SortAsc})

		if c[0].Rport != 80 {
			t.Errorf("expected port 80 first, got %d", c[0].Rport)
		}
		if c[1].Rport != 443 {
			t.Errorf("expected port 443 second, got %d", c[1].Rport)
		}
		if c[2].Rport != 8080 {
			t.Errorf("expected port 8080 last, got %d", c[2].Rport)
		}
	})

	t.Run("sort by rport descending", func(t *testing.T) {
		c := make([]Connection, len(conns))
		copy(c, conns)

		SortConnections(c, SortOptions{Field: SortByRport, Direction: SortDesc})

		if c[0].Rport != 8080 {
			t.Errorf("expected port 8080 first, got %d", c[0].Rport)
		}
	})
}

