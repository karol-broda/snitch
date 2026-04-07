//go:build linux

package bandwidth

import (
	"bufio"
	"os"
	"strconv"
	"strings"
	"time"
)

// GetInterfaceStats reads network interface statistics from /proc/net/dev
func GetInterfaceStats() (*Snapshot, error) {
	file, err := os.Open("/proc/net/dev")
	if err != nil {
		return nil, err
	}
	defer file.Close()

	snapshot := &Snapshot{
		Interfaces: make(map[string]InterfaceStats),
		Timestamp:  time.Now(),
	}

	scanner := bufio.NewScanner(file)

	scanner.Scan()
	scanner.Scan()

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		stats, err := parseDevLine(line)
		if err != nil {
			continue
		}

		if stats.Name == "lo" && stats.RxBytes == 0 && stats.TxBytes == 0 {
			continue
		}

		stats.Timestamp = snapshot.Timestamp
		snapshot.Interfaces[stats.Name] = stats
	}

	return snapshot, scanner.Err()
}

// fields: iface: rx_bytes rx_packets rx_errs rx_drop rx_fifo rx_frame rx_compressed rx_multicast | tx_bytes tx_packets ...
func parseDevLine(line string) (InterfaceStats, error) {
	var stats InterfaceStats

	colonIdx := strings.Index(line, ":")
	if colonIdx == -1 {
		return stats, nil
	}

	stats.Name = strings.TrimSpace(line[:colonIdx])
	fields := strings.Fields(line[colonIdx+1:])

	if len(fields) < 16 {
		return stats, nil
	}

	rxBytes, err := strconv.ParseUint(fields[0], 10, 64)
	if err != nil {
		return stats, err
	}
	stats.RxBytes = rxBytes

	rxPackets, err := strconv.ParseUint(fields[1], 10, 64)
	if err != nil {
		return stats, err
	}
	stats.RxPackets = rxPackets

	txBytes, err := strconv.ParseUint(fields[8], 10, 64)
	if err != nil {
		return stats, err
	}
	stats.TxBytes = txBytes

	txPackets, err := strconv.ParseUint(fields[9], 10, 64)
	if err != nil {
		return stats, err
	}
	stats.TxPackets = txPackets

	return stats, nil
}

// GetProcessIOStats reads I/O statistics for a process from /proc/<pid>/io
func GetProcessIOStats(pid int) (readBytes, writeBytes uint64, err error) {
	path := "/proc/" + strconv.Itoa(pid) + "/io"
	file, err := os.Open(path)
	if err != nil {
		return 0, 0, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "read_bytes:") {
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				readBytes, _ = strconv.ParseUint(fields[1], 10, 64)
			}
		} else if strings.HasPrefix(line, "write_bytes:") {
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				writeBytes, _ = strconv.ParseUint(fields[1], 10, 64)
			}
		}
	}

	return readBytes, writeBytes, scanner.Err()
}
