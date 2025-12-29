package collector

import "time"

type Connection struct {
	TS         time.Time `json:"ts"`
	PID        int       `json:"pid"`
	Process    string    `json:"process"`
	Cmdline    string    `json:"cmdline,omitempty"`
	Cwd        string    `json:"cwd,omitempty"`
	User       string    `json:"user"`
	UID        int       `json:"uid"`
	Proto      string    `json:"proto"`
	IPVersion  string    `json:"ipversion"`
	State      string    `json:"state"`
	Laddr      string    `json:"laddr"`
	Lport      int       `json:"lport"`
	Raddr      string    `json:"raddr"`
	Rport      int       `json:"rport"`
	Interface  string    `json:"interface"`
	RxBytes    int64     `json:"rx_bytes"`
	TxBytes    int64     `json:"tx_bytes"`
	RttMs      float64   `json:"rtt_ms"`
	Mark       string    `json:"mark"`
	Namespace  string    `json:"namespace"`
	Inode      int64     `json:"inode"`
}
