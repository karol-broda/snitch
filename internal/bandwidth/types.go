package bandwidth

import "time"

// InterfaceStats holds raw byte counters for a network interface
type InterfaceStats struct {
	Name      string
	RxBytes   uint64
	TxBytes   uint64
	RxPackets uint64
	TxPackets uint64
	Timestamp time.Time
}

// InterfaceRate holds computed rates for a network interface
type InterfaceRate struct {
	Name         string
	RxBytesPerSec float64
	TxBytesPerSec float64
	RxPacketsPerSec float64
	TxPacketsPerSec float64
	RxBytes      uint64
	TxBytes      uint64
}

// ProcessBandwidth holds bandwidth info for a process
type ProcessBandwidth struct {
	PID           int
	Process       string
	RxBytesPerSec float64
	TxBytesPerSec float64
	Connections   int
}

// Snapshot holds a point-in-time view of all interface stats
type Snapshot struct {
	Interfaces map[string]InterfaceStats
	Timestamp  time.Time
}

// RateResult holds computed rates between two snapshots
type RateResult struct {
	Interfaces []InterfaceRate
	Duration   time.Duration
	Timestamp  time.Time
}

