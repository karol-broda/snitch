package bandwidth

import (
	"sync"
	"time"
)

// Tracker maintains state for computing bandwidth rates
type Tracker struct {
	mu       sync.RWMutex
	previous *Snapshot
	current  *Snapshot
}

// NewTracker creates a new bandwidth tracker
func NewTracker() *Tracker {
	return &Tracker{}
}

// Update takes a new snapshot and computes rates
func (t *Tracker) Update() (*RateResult, error) {
	snapshot, err := GetInterfaceStats()
	if err != nil {
		return nil, err
	}

	t.mu.Lock()
	defer t.mu.Unlock()

	t.previous = t.current
	t.current = snapshot

	if t.previous == nil {
		return &RateResult{
			Interfaces: make([]InterfaceRate, 0),
			Duration:   0,
			Timestamp:  snapshot.Timestamp,
		}, nil
	}

	return t.computeRates(), nil
}

func (t *Tracker) computeRates() *RateResult {
	duration := t.current.Timestamp.Sub(t.previous.Timestamp)
	if duration <= 0 {
		duration = time.Millisecond
	}

	seconds := duration.Seconds()
	rates := make([]InterfaceRate, 0, len(t.current.Interfaces))

	for name, curr := range t.current.Interfaces {
		prev, exists := t.previous.Interfaces[name]
		if !exists {
			continue
		}

		rxDelta := curr.RxBytes - prev.RxBytes
		txDelta := curr.TxBytes - prev.TxBytes
		rxPktDelta := curr.RxPackets - prev.RxPackets
		txPktDelta := curr.TxPackets - prev.TxPackets

		if curr.RxBytes < prev.RxBytes {
			rxDelta = curr.RxBytes
		}
		if curr.TxBytes < prev.TxBytes {
			txDelta = curr.TxBytes
		}

		rate := InterfaceRate{
			Name:            name,
			RxBytesPerSec:   float64(rxDelta) / seconds,
			TxBytesPerSec:   float64(txDelta) / seconds,
			RxPacketsPerSec: float64(rxPktDelta) / seconds,
			TxPacketsPerSec: float64(txPktDelta) / seconds,
			RxBytes:         curr.RxBytes,
			TxBytes:         curr.TxBytes,
		}
		rates = append(rates, rate)
	}

	return &RateResult{
		Interfaces: rates,
		Duration:   duration,
		Timestamp:  t.current.Timestamp,
	}
}

// GetRate returns the current computed rate without updating
func (t *Tracker) GetRate() *RateResult {
	t.mu.RLock()
	defer t.mu.RUnlock()

	if t.previous == nil || t.current == nil {
		return nil
	}

	return t.computeRates()
}

// FormatBytes formats bytes into human-readable string
func FormatBytes(bytes float64) string {
	const unit = 1024
	if bytes < unit {
		return formatFloat(bytes) + " B"
	}
	div, exp := float64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return formatFloat(bytes/div) + " " + string("KMGTPE"[exp]) + "iB"
}

// FormatRate formats bytes per second into human-readable rate
func FormatRate(bytesPerSec float64) string {
	return FormatBytes(bytesPerSec) + "/s"
}

func formatFloat(f float64) string {
	if f < 10 {
		return floatToString(f, 2)
	}
	if f < 100 {
		return floatToString(f, 1)
	}
	return floatToString(f, 0)
}

func floatToString(f float64, precision int) string {
	if precision == 0 {
		return intToString(int64(f))
	}
	
	intPart := int64(f)
	fracPart := f - float64(intPart)

	result := intToString(intPart) + "."

	for i := 0; i < precision; i++ {
		fracPart *= 10
		digit := int(fracPart)
		result += string(rune('0' + digit))
		fracPart -= float64(digit)
	}
	
	return result
}

func intToString(n int64) string {
	if n == 0 {
		return "0"
	}
	negative := n < 0
	if negative {
		n = -n
	}
	var digits []byte
	for n > 0 {
		digits = append([]byte{byte('0' + n%10)}, digits...)
		n /= 10
	}

	if negative {
		return "-" + string(digits)
	}
	return string(digits)
}
