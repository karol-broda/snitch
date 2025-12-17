//go:build darwin

package collector

/*
#include <libproc.h>
#include <sys/proc_info.h>
#include <sys/socket.h>
#include <netinet/in.h>
#include <netinet/tcp_fsm.h>
#include <arpa/inet.h>
#include <pwd.h>
#include <stdlib.h>
#include <string.h>

// get process name by pid
static int get_proc_name(int pid, char *name, int namelen) {
    return proc_name(pid, name, namelen);
}

// get uid for a process
static int get_proc_uid(int pid) {
    struct proc_bsdinfo info;
    int ret = proc_pidinfo(pid, PROC_PIDTBSDINFO, 0, &info, sizeof(info));
    if (ret <= 0) {
        return -1;
    }
    return info.pbi_uid;
}

// get username from uid
static const char* get_username(int uid) {
    struct passwd *pw = getpwuid(uid);
    if (pw == NULL) {
        return NULL;
    }
    return pw->pw_name;
}

// socket info extraction - handles the union properly in C
typedef struct {
    int family;
    int sock_type;
    int protocol;
    int state;
    uint32_t laddr4;
    uint32_t raddr4;
    uint8_t laddr6[16];
    uint8_t raddr6[16];
    int lport;
    int rport;
} socket_info_t;

static int get_socket_info(int pid, int fd, socket_info_t *info) {
    struct socket_fdinfo si;
    int ret = proc_pidfdinfo(pid, fd, PROC_PIDFDSOCKETINFO, &si, sizeof(si));
    if (ret <= 0) {
        return -1;
    }

    info->family = si.psi.soi_family;
    info->sock_type = si.psi.soi_type;
    info->protocol = si.psi.soi_protocol;

    if (info->family == AF_INET) {
        if (info->sock_type == SOCK_STREAM) {
            // TCP
            info->state = si.psi.soi_proto.pri_tcp.tcpsi_state;
            info->laddr4 = si.psi.soi_proto.pri_tcp.tcpsi_ini.insi_laddr.ina_46.i46a_addr4.s_addr;
            info->raddr4 = si.psi.soi_proto.pri_tcp.tcpsi_ini.insi_faddr.ina_46.i46a_addr4.s_addr;
            info->lport = ntohs(si.psi.soi_proto.pri_tcp.tcpsi_ini.insi_lport);
            info->rport = ntohs(si.psi.soi_proto.pri_tcp.tcpsi_ini.insi_fport);
        } else if (info->sock_type == SOCK_DGRAM) {
            // UDP
            info->state = 0;
            info->laddr4 = si.psi.soi_proto.pri_in.insi_laddr.ina_46.i46a_addr4.s_addr;
            info->raddr4 = si.psi.soi_proto.pri_in.insi_faddr.ina_46.i46a_addr4.s_addr;
            info->lport = ntohs(si.psi.soi_proto.pri_in.insi_lport);
            info->rport = ntohs(si.psi.soi_proto.pri_in.insi_fport);
        }
    } else if (info->family == AF_INET6) {
        if (info->sock_type == SOCK_STREAM) {
            // TCP6
            info->state = si.psi.soi_proto.pri_tcp.tcpsi_state;
            memcpy(info->laddr6, &si.psi.soi_proto.pri_tcp.tcpsi_ini.insi_laddr.ina_6, 16);
            memcpy(info->raddr6, &si.psi.soi_proto.pri_tcp.tcpsi_ini.insi_faddr.ina_6, 16);
            info->lport = ntohs(si.psi.soi_proto.pri_tcp.tcpsi_ini.insi_lport);
            info->rport = ntohs(si.psi.soi_proto.pri_tcp.tcpsi_ini.insi_fport);
        } else if (info->sock_type == SOCK_DGRAM) {
            // UDP6
            info->state = 0;
            memcpy(info->laddr6, &si.psi.soi_proto.pri_in.insi_laddr.ina_6, 16);
            memcpy(info->raddr6, &si.psi.soi_proto.pri_in.insi_faddr.ina_6, 16);
            info->lport = ntohs(si.psi.soi_proto.pri_in.insi_lport);
            info->rport = ntohs(si.psi.soi_proto.pri_in.insi_fport);
        }
    }

    return 0;
}
*/
import "C"

import (
	"fmt"
	"net"
	"strconv"
	"time"
	"unsafe"
)

// DefaultCollector implements the Collector interface using libproc on macOS
type DefaultCollector struct{}

// GetConnections fetches all network connections using libproc
func (dc *DefaultCollector) GetConnections() ([]Connection, error) {
	pids, err := listAllPids()
	if err != nil {
		return nil, fmt.Errorf("failed to list pids: %w", err)
	}

	var connections []Connection

	for _, pid := range pids {
		procConns, err := getConnectionsForPid(pid)
		if err != nil {
			continue
		}
		connections = append(connections, procConns...)
	}

	return connections, nil
}

// GetAllConnections returns network connections
func GetAllConnections() ([]Connection, error) {
	return GetConnections()
}

func listAllPids() ([]int, error) {
	numPids := C.proc_listpids(C.PROC_ALL_PIDS, 0, nil, 0)
	if numPids <= 0 {
		return nil, fmt.Errorf("proc_listpids failed")
	}

	bufSize := C.int(numPids) * C.int(unsafe.Sizeof(C.int(0)))
	buf := make([]C.int, numPids)

	numPids = C.proc_listpids(C.PROC_ALL_PIDS, 0, unsafe.Pointer(&buf[0]), bufSize)
	if numPids <= 0 {
		return nil, fmt.Errorf("proc_listpids failed")
	}

	count := int(numPids) / int(unsafe.Sizeof(C.int(0)))
	pids := make([]int, 0, count)
	for i := 0; i < count; i++ {
		if buf[i] > 0 {
			pids = append(pids, int(buf[i]))
		}
	}

	return pids, nil
}

func getConnectionsForPid(pid int) ([]Connection, error) {
	procName := getProcessName(pid)
	uid := int(C.get_proc_uid(C.int(pid)))
	user := ""
	if uid >= 0 {
		cUser := C.get_username(C.int(uid))
		if cUser != nil {
			user = C.GoString(cUser)
		} else {
			user = strconv.Itoa(uid)
		}
	}

	bufSize := C.proc_pidinfo(C.int(pid), C.PROC_PIDLISTFDS, 0, nil, 0)
	if bufSize <= 0 {
		return nil, fmt.Errorf("failed to get fd list size")
	}

	buf := make([]byte, bufSize)
	ret := C.proc_pidinfo(C.int(pid), C.PROC_PIDLISTFDS, 0, unsafe.Pointer(&buf[0]), bufSize)
	if ret <= 0 {
		return nil, fmt.Errorf("failed to get fd list")
	}

	fdInfoSize := int(unsafe.Sizeof(C.struct_proc_fdinfo{}))
	numFds := int(ret) / fdInfoSize

	var connections []Connection

	for i := 0; i < numFds; i++ {
		fdInfo := (*C.struct_proc_fdinfo)(unsafe.Pointer(&buf[i*fdInfoSize]))

		if fdInfo.proc_fdtype != C.PROX_FDTYPE_SOCKET {
			continue
		}

		conn, ok := getSocketInfo(pid, int(fdInfo.proc_fd), procName, uid, user)
		if ok {
			connections = append(connections, conn)
		}
	}

	return connections, nil
}

func getSocketInfo(pid, fd int, procName string, uid int, user string) (Connection, bool) {
	var info C.socket_info_t

	ret := C.get_socket_info(C.int(pid), C.int(fd), &info)
	if ret != 0 {
		return Connection{}, false
	}

	// only interested in IPv4 and IPv6
	if info.family != C.AF_INET && info.family != C.AF_INET6 {
		return Connection{}, false
	}

	// only TCP and UDP
	if info.sock_type != C.SOCK_STREAM && info.sock_type != C.SOCK_DGRAM {
		return Connection{}, false
	}

	proto := "tcp"
	if info.sock_type == C.SOCK_DGRAM {
		proto = "udp"
	}

	ipVersion := "IPv4"
	if info.family == C.AF_INET6 {
		ipVersion = "IPv6"
		proto = proto + "6"
	}

	var laddr, raddr string

	if info.family == C.AF_INET {
		laddr = ipv4ToString(uint32(info.laddr4))
		raddr = ipv4ToString(uint32(info.raddr4))
	} else {
		laddr = ipv6ToString(info.laddr6)
		raddr = ipv6ToString(info.raddr6)
	}

	if laddr == "0.0.0.0" || laddr == "::" {
		laddr = "*"
	}
	if raddr == "0.0.0.0" || raddr == "::" {
		raddr = "*"
	}

	state := ""
	if info.sock_type == C.SOCK_STREAM {
		state = tcpStateToString(int(info.state))
	} else if info.sock_type == C.SOCK_DGRAM {
		// udp is connectionless - infer state from remote address
		if raddr == "*" && int(info.rport) == 0 {
			state = "LISTEN"
		} else {
			state = "ESTABLISHED"
		}
	}

	conn := Connection{
		TS:        time.Now(),
		Proto:     proto,
		IPVersion: ipVersion,
		State:     state,
		Laddr:     laddr,
		Lport:     int(info.lport),
		Raddr:     raddr,
		Rport:     int(info.rport),
		PID:       pid,
		Process:   procName,
		UID:       uid,
		User:      user,
		Interface: guessNetworkInterface(laddr),
	}

	return conn, true
}

func getProcessName(pid int) string {
	var name [256]C.char
	ret := C.get_proc_name(C.int(pid), &name[0], 256)
	if ret <= 0 {
		return ""
	}
	return C.GoString(&name[0])
}

func ipv4ToString(addr uint32) string {
	ip := make(net.IP, 4)
	ip[0] = byte(addr)
	ip[1] = byte(addr >> 8)
	ip[2] = byte(addr >> 16)
	ip[3] = byte(addr >> 24)
	return ip.String()
}

func ipv6ToString(addr [16]C.uint8_t) string {
	ip := make(net.IP, 16)
	for i := 0; i < 16; i++ {
		ip[i] = byte(addr[i])
	}

	if ip.To4() != nil {
		return ip.To4().String()
	}

	return ip.String()
}

func tcpStateToString(state int) string {
	// macOS TCP states from netinet/tcp_fsm.h
	states := map[int]string{
		0:  "CLOSED",
		1:  "LISTEN",
		2:  "SYN_SENT",
		3:  "SYN_RECV",
		4:  "ESTABLISHED",
		5:  "CLOSE_WAIT",
		6:  "FIN_WAIT1",
		7:  "CLOSING",
		8:  "LAST_ACK",
		9:  "FIN_WAIT2",
		10: "TIME_WAIT",
	}

	if s, exists := states[state]; exists {
		return s
	}
	return ""
}
