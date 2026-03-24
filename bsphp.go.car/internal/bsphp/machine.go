package bsphp

import (
	"crypto/rand"
	"encoding/hex"
	"os"
	"path/filepath"
	"sync"

	"github.com/shirou/gopsutil/v3/host"
)

var (
	machineMu     sync.Mutex
	fallbackSaved string
)

// MachineCode 本机特征串：优先 HostID，否则本地持久化随机 ID。
func MachineCode() string {
	if id, err := host.HostID(); err == nil && id != "" {
		return id
	}
	machineMu.Lock()
	defer machineMu.Unlock()
	if fallbackSaved != "" {
		return fallbackSaved
	}
	dir, _ := os.UserConfigDir()
	p := filepath.Join(dir, "bsphp_go_car", "machine_id")
	if b, err := os.ReadFile(p); err == nil && len(b) > 0 {
		fallbackSaved = string(b)
		return fallbackSaved
	}
	var buf [16]byte
	_, _ = rand.Read(buf[:])
	fallbackSaved = hex.EncodeToString(buf[:])
	_ = os.MkdirAll(filepath.Dir(p), 0700)
	_ = os.WriteFile(p, []byte(fallbackSaved), 0600)
	return fallbackSaved
}
