//go:build linux

package collector

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// set SNITCH_DEBUG_TIMING=1 to enable timing diagnostics
var debugTiming = os.Getenv("SNITCH_DEBUG_TIMING") != ""

func logTiming(label string, start time.Time, extra ...string) {
	if !debugTiming {
		return
	}
	elapsed := time.Since(start)
	if len(extra) > 0 {
		fmt.Fprintf(os.Stderr, "[timing] %s: %v (%s)\n", label, elapsed, extra[0])
	} else {
		fmt.Fprintf(os.Stderr, "[timing] %s: %v\n", label, elapsed)
	}
}

// userCache caches uid to username mappings to avoid repeated lookups
var userCache = struct {
	sync.RWMutex
	m map[int]string
}{m: make(map[int]string)}

func lookupUsername(uid int) string {
	userCache.RLock()
	if username, exists := userCache.m[uid]; exists {
		userCache.RUnlock()
		return username
	}
	userCache.RUnlock()

	start := time.Now()
	username := strconv.Itoa(uid)
	u, err := user.LookupId(strconv.Itoa(uid))
	if err == nil && u != nil {
		username = u.Username
	}
	elapsed := time.Since(start)
	if debugTiming && elapsed > 10*time.Millisecond {
		fmt.Fprintf(os.Stderr, "[timing] user.LookupId(%d) slow: %v\n", uid, elapsed)
	}

	userCache.Lock()
	userCache.m[uid] = username
	userCache.Unlock()

	return username
}

// DefaultCollector implements the Collector interface using /proc filesystem
type DefaultCollector struct{}

// GetConnections fetches all network connections by parsing /proc files
func (dc *DefaultCollector) GetConnections() ([]Connection, error) {
	totalStart := time.Now()
	defer func() { logTiming("GetConnections total", totalStart) }()

	inodeStart := time.Now()
	inodeMap, err := buildInodeToProcessMap()
	logTiming("buildInodeToProcessMap", inodeStart, fmt.Sprintf("%d inodes", len(inodeMap)))
	if err != nil {
		return nil, fmt.Errorf("failed to build inode map: %w", err)
	}

	var connections []Connection

	parseStart := time.Now()
	tcpConns, err := parseProcNet("/proc/net/tcp", "tcp", 4, inodeMap)
	if err == nil {
		connections = append(connections, tcpConns...)
	}

	tcpConns6, err := parseProcNet("/proc/net/tcp6", "tcp6", 6, inodeMap)
	if err == nil {
		connections = append(connections, tcpConns6...)
	}

	udpConns, err := parseProcNet("/proc/net/udp", "udp", 4, inodeMap)
	if err == nil {
		connections = append(connections, udpConns...)
	}

	udpConns6, err := parseProcNet("/proc/net/udp6", "udp6", 6, inodeMap)
	if err == nil {
		connections = append(connections, udpConns6...)
	}
	logTiming("parseProcNet (all)", parseStart, fmt.Sprintf("%d connections", len(connections)))

	return connections, nil
}

// GetAllConnections returns both network and Unix domain socket connections
func GetAllConnections() ([]Connection, error) {
	networkConns, err := GetConnections()
	if err != nil {
		return nil, err
	}

	unixConns, err := GetUnixSockets()
	if err == nil {
		networkConns = append(networkConns, unixConns...)
	}

	return networkConns, nil
}

type processInfo struct {
	pid     int
	command string
	uid     int
	user    string
}

type inodeEntry struct {
	inode int64
	info  *processInfo
}

func buildInodeToProcessMap() (map[int64]*processInfo, error) {
	readDirStart := time.Now()
	procDir, err := os.Open("/proc")
	if err != nil {
		return nil, err
	}
	defer procDir.Close()

	entries, err := procDir.Readdir(-1)
	if err != nil {
		return nil, err
	}

	// collect pids first
	pids := make([]int, 0, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		pid, err := strconv.Atoi(entry.Name())
		if err != nil {
			continue
		}
		pids = append(pids, pid)
	}
	logTiming("  readdir /proc", readDirStart, fmt.Sprintf("%d pids", len(pids)))

	// process pids in parallel with limited concurrency
	scanStart := time.Now()
	const numWorkers = 8
	pidChan := make(chan int, len(pids))
	resultChan := make(chan []inodeEntry, len(pids))

	var totalFDs atomic.Int64
	var wg sync.WaitGroup
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for pid := range pidChan {
				entries := scanProcessSockets(pid)
				if len(entries) > 0 {
					totalFDs.Add(int64(len(entries)))
					resultChan <- entries
				}
			}
		}()
	}

	for _, pid := range pids {
		pidChan <- pid
	}
	close(pidChan)

	go func() {
		wg.Wait()
		close(resultChan)
	}()

	inodeMap := make(map[int64]*processInfo)
	for entries := range resultChan {
		for _, e := range entries {
			inodeMap[e.inode] = e.info
		}
	}
	logTiming("  scan all processes", scanStart, fmt.Sprintf("%d socket fds scanned", totalFDs.Load()))

	return inodeMap, nil
}

func scanProcessSockets(pid int) []inodeEntry {
	start := time.Now()

	procInfo, err := getProcessInfo(pid)
	if err != nil {
		return nil
	}

	pidStr := strconv.Itoa(pid)
	fdDir := filepath.Join("/proc", pidStr, "fd")
	fdEntries, err := os.ReadDir(fdDir)
	if err != nil {
		return nil
	}

	var results []inodeEntry
	for _, fdEntry := range fdEntries {
		fdPath := filepath.Join(fdDir, fdEntry.Name())
		link, err := os.Readlink(fdPath)
		if err != nil {
			continue
		}

		if strings.HasPrefix(link, "socket:[") && strings.HasSuffix(link, "]") {
			inodeStr := link[8 : len(link)-1]
			inode, err := strconv.ParseInt(inodeStr, 10, 64)
			if err != nil {
				continue
			}
			results = append(results, inodeEntry{inode: inode, info: procInfo})
		}
	}

	elapsed := time.Since(start)
	if debugTiming && elapsed > 20*time.Millisecond {
		fmt.Fprintf(os.Stderr, "[timing] slow process scan: pid=%d (%s) fds=%d time=%v\n",
			pid, procInfo.command, len(fdEntries), elapsed)
	}

	return results
}

func getProcessInfo(pid int) (*processInfo, error) {
	info := &processInfo{pid: pid}

	commPath := filepath.Join("/proc", strconv.Itoa(pid), "comm")
	commData, err := os.ReadFile(commPath)
	if err == nil && len(commData) > 0 {
		info.command = strings.TrimSpace(string(commData))
	}

	if info.command == "" {
		cmdlinePath := filepath.Join("/proc", strconv.Itoa(pid), "cmdline")
		cmdlineData, err := os.ReadFile(cmdlinePath)
		if err != nil {
			return nil, err
		}

		if len(cmdlineData) > 0 {
			parts := bytes.Split(cmdlineData, []byte{0})
			if len(parts) > 0 && len(parts[0]) > 0 {
				fullPath := string(parts[0])
				baseName := filepath.Base(fullPath)
				if strings.Contains(baseName, " ") {
					baseName = strings.Fields(baseName)[0]
				}
				info.command = baseName
			}
		}
	}

	statusPath := filepath.Join("/proc", strconv.Itoa(pid), "status")
	statusFile, err := os.Open(statusPath)
	if err != nil {
		return info, nil
	}
	defer statusFile.Close()

	scanner := bufio.NewScanner(statusFile)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "Uid:") {
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				uid, err := strconv.Atoi(fields[1])
				if err == nil {
					info.uid = uid
					info.user = lookupUsername(uid)
				}
			}
			break
		}
	}

	return info, nil
}

func parseProcNet(path, proto string, ipVersion int, inodeMap map[int64]*processInfo) ([]Connection, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var connections []Connection
	scanner := bufio.NewScanner(file)

	scanner.Scan()

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 10 {
			continue
		}

		localAddr, localPort, err := parseHexAddr(fields[1])
		if err != nil {
			continue
		}

		remoteAddr, remotePort, err := parseHexAddr(fields[2])
		if err != nil {
			continue
		}

		stateHex := fields[3]
		state := parseState(stateHex, proto)

		inode, _ := strconv.ParseInt(fields[9], 10, 64)

		// refine udp state: if unconnected and remote is wildcard, it's listening
		if strings.HasPrefix(proto, "udp") && state == "UNCONNECTED" {
			if remoteAddr == "*" && remotePort == 0 {
				state = "LISTEN"
			}
		}

		conn := Connection{
			TS:        time.Now(),
			Proto:     proto,
			IPVersion: fmt.Sprintf("IPv%d", ipVersion),
			State:     state,
			Laddr:     localAddr,
			Lport:     localPort,
			Raddr:     remoteAddr,
			Rport:     remotePort,
			Inode:     inode,
		}

		if procInfo, exists := inodeMap[inode]; exists {
			conn.PID = procInfo.pid
			conn.Process = procInfo.command
			conn.UID = procInfo.uid
			conn.User = procInfo.user
		}

		conn.Interface = guessNetworkInterface(localAddr)

		connections = append(connections, conn)
	}

	return connections, scanner.Err()
}

func parseState(hexState, proto string) string {
	state, err := strconv.ParseInt(hexState, 16, 32)
	if err != nil {
		return ""
	}

	tcpStates := map[int64]string{
		0x01: "ESTABLISHED",
		0x02: "SYN_SENT",
		0x03: "SYN_RECV",
		0x04: "FIN_WAIT1",
		0x05: "FIN_WAIT2",
		0x06: "TIME_WAIT",
		0x07: "CLOSE",
		0x08: "CLOSE_WAIT",
		0x09: "LAST_ACK",
		0x0A: "LISTEN",
		0x0B: "CLOSING",
	}

	if strings.HasPrefix(proto, "tcp") {
		if s, exists := tcpStates[state]; exists {
			return s
		}
		return ""
	}

	// udp states - udp is connectionless so the kernel reuses tcp state values
	// with different meanings:
	// 0x07 (TCP_CLOSE) = unconnected socket, typically bound and listening
	// 0x01 (TCP_ESTABLISHED) = "connected" socket (connect() was called)
	udpStates := map[int64]string{
		0x01: "ESTABLISHED",
		0x07: "UNCONNECTED",
	}

	if s, exists := udpStates[state]; exists {
		return s
	}

	return ""
}

func parseHexAddr(hexAddr string) (string, int, error) {
	parts := strings.Split(hexAddr, ":")
	if len(parts) != 2 {
		return "", 0, fmt.Errorf("invalid address format")
	}

	hexIP := parts[0]

	port, err := strconv.ParseInt(parts[1], 16, 32)
	if err != nil {
		return "", 0, err
	}

	if len(hexIP) == 8 {
		ip1, _ := strconv.ParseInt(hexIP[6:8], 16, 32)
		ip2, _ := strconv.ParseInt(hexIP[4:6], 16, 32)
		ip3, _ := strconv.ParseInt(hexIP[2:4], 16, 32)
		ip4, _ := strconv.ParseInt(hexIP[0:2], 16, 32)
		addr := fmt.Sprintf("%d.%d.%d.%d", ip1, ip2, ip3, ip4)

		if addr == "0.0.0.0" {
			addr = "*"
		}

		return addr, int(port), nil
	} else if len(hexIP) == 32 {
		var ipv6Parts []string
		for i := 0; i < 32; i += 8 {
			word := hexIP[i : i+8]
			p1 := word[6:8] + word[4:6] + word[2:4] + word[0:2]
			ipv6Parts = append(ipv6Parts, p1)
		}

		fullAddr := strings.Join(ipv6Parts, "")
		var formatted []string
		for i := 0; i < len(fullAddr); i += 4 {
			formatted = append(formatted, fullAddr[i:i+4])
		}
		addr := strings.Join(formatted, ":")

		addr = simplifyIPv6(addr)

		if addr == "::" || addr == "0:0:0:0:0:0:0:0" {
			addr = "*"
		}

		return addr, int(port), nil
	}

	return "", 0, fmt.Errorf("unsupported address format")
}

func GetUnixSockets() ([]Connection, error) {
	connections := []Connection{}

	file, err := os.Open("/proc/net/unix")
	if err != nil {
		return connections, nil
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	scanner.Scan()

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		fields := strings.Fields(line)
		if len(fields) < 7 {
			continue
		}

		inode, _ := strconv.ParseInt(fields[6], 10, 64)
		path := ""
		if len(fields) > 7 {
			path = fields[7]
		}

		conn := Connection{
			TS:        time.Now(),
			Proto:     "unix",
			Laddr:     path,
			Raddr:     "",
			State:     "CONNECTED",
			Inode:     inode,
			Interface: "unix",
		}

		connections = append(connections, conn)
	}

	return connections, nil
}

