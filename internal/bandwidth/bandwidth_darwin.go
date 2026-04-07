//go:build darwin

package bandwidth

/*
#include <sys/sysctl.h>
#include <net/if.h>
#include <net/route.h>
#include <string.h>
#include <stdlib.h>

struct if_stats {
    char name[32];
    unsigned long long rx_bytes;
    unsigned long long tx_bytes;
    unsigned long long rx_packets;
    unsigned long long tx_packets;
};

static int get_interface_stats(struct if_stats *stats, int max_count) {
    int mib[6] = {CTL_NET, PF_ROUTE, 0, 0, NET_RT_IFLIST2, 0};
    size_t len;
    
    if (sysctl(mib, 6, NULL, &len, NULL, 0) < 0) {
        return -1;
    }
    
    char *buf = malloc(len);
    if (buf == NULL) {
        return -1;
    }
    
    if (sysctl(mib, 6, buf, &len, NULL, 0) < 0) {
        free(buf);
        return -1;
    }
    
    int count = 0;
    char *next = buf;
    char *end = buf + len;
    
    while (next < end && count < max_count) {
        struct if_msghdr *ifm = (struct if_msghdr *)next;
        next += ifm->ifm_msglen;
        
        if (ifm->ifm_type == RTM_IFINFO2) {
            struct if_msghdr2 *ifm2 = (struct if_msghdr2 *)ifm;
            struct sockaddr_dl *sdl = (struct sockaddr_dl *)(ifm2 + 1);
            
            if (sdl->sdl_nlen > 0 && sdl->sdl_nlen < 32) {
                strncpy(stats[count].name, sdl->sdl_data, sdl->sdl_nlen);
                stats[count].name[sdl->sdl_nlen] = '\0';
                stats[count].rx_bytes = ifm2->ifm_data.ifi_ibytes;
                stats[count].tx_bytes = ifm2->ifm_data.ifi_obytes;
                stats[count].rx_packets = ifm2->ifm_data.ifi_ipackets;
                stats[count].tx_packets = ifm2->ifm_data.ifi_opackets;
                count++;
            }
        }
    }
    
    free(buf);
    return count;
}
*/
import "C"

import (
	"time"
)

// GetInterfaceStats reads network interface statistics using sysctl
func GetInterfaceStats() (*Snapshot, error) {
	const maxInterfaces = 64
	var cStats [maxInterfaces]C.struct_if_stats

	count := C.get_interface_stats(&cStats[0], maxInterfaces)
	if count < 0 {
		return &Snapshot{
			Interfaces: make(map[string]InterfaceStats),
			Timestamp:  time.Now(),
		}, nil
	}

	snapshot := &Snapshot{
		Interfaces: make(map[string]InterfaceStats),
		Timestamp:  time.Now(),
	}

	for i := 0; i < int(count); i++ {
		name := C.GoString(&cStats[i].name[0])

		if name == "lo0" && cStats[i].rx_bytes == 0 && cStats[i].tx_bytes == 0 {
			continue
		}

		stats := InterfaceStats{
			Name:      name,
			RxBytes:   uint64(cStats[i].rx_bytes),
			TxBytes:   uint64(cStats[i].tx_bytes),
			RxPackets: uint64(cStats[i].rx_packets),
			TxPackets: uint64(cStats[i].tx_packets),
			Timestamp: snapshot.Timestamp,
		}

		snapshot.Interfaces[name] = stats
	}

	return snapshot, nil
}

// GetProcessIOStats is unavailable on macOS without dtrace
func GetProcessIOStats(pid int) (readBytes, writeBytes uint64, err error) {
	return 0, 0, nil
}
