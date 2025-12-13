package service

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/chiwen/client/internal/api/mode"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/spf13/viper"
)

// HeartbeatLoop 负责周期性发送心跳：会在第一次发送时包含静态信息，随后只发送动态信息（如果静态变化会一并发送）
// agentSecret 是解密后的明文 secret（建议长度/格式由服务端定义），interval 单位秒
func HeartbeatLoop(id, agentSecret string, interval int, dataDir string) {
	ticker := time.NewTicker(time.Duration(interval) * time.Second)
	defer ticker.Stop()

	var lastStaticHash string

	// 立即发送一次心跳（启动时）
	if err := sendOneHeartbeat(id, agentSecret, &lastStaticHash); err != nil {
		// 记录但不退出
		_ = err
	}

	for range ticker.C {
		_ = sendOneHeartbeat(id, agentSecret, &lastStaticHash)
	}
}

// sendOneHeartbeat 收集信息、签名并发送到服务端
func sendOneHeartbeat(id, agentSecret string, lastStaticHash *string) error {
	timestamp := time.Now().Unix()

	staticInfo := CollectStaticInfo()
	dynamicInfo := CollectDynamicInfo()

	// 将 static+dynamic 合并到 metrics map
	metrics := map[string]interface{}{
		"static_info":  staticInfo,
		"dynamic_info": dynamicInfo,
	}

	// 检查静态信息是否变化（通过 JSON hash）
	staticBytes, err := json.Marshal(staticInfo)
	if err != nil {
		return err
	}
	staticHash := fmt.Sprintf("%x", sha256.Sum256(staticBytes))
	includeStatic := false
	if *lastStaticHash == "" || *lastStaticHash != staticHash {
		includeStatic = true
		*lastStaticHash = staticHash
	}

	if !includeStatic {
		// 当 static 不变时只保留 dynamic_info
		metrics = map[string]interface{}{
			"dynamic_info": dynamicInfo,
		}
	}

	// canonical payload: id|timestamp|metrics-json（metrics 压缩成紧凑 JSON）
	metricsBytes, err := json.Marshal(metrics)
	if err != nil {
		return err
	}
	payload := fmt.Sprintf("%s|%d|%s", id, timestamp, string(metricsBytes))

	// HMAC-SHA256 签名（agentSecret 作为 key）
	sig := hmacBase64Sign([]byte(agentSecret), []byte(payload))

	hb := &mode.HeartbeatRequest{
		ID:        id,
		Timestamp: timestamp,
		Metrics:   metricsToInterfaceMap(metrics),
		Signature: sig,
	}

	// 发送心跳
	_, err = SendHeartbeat(hb)
	if err != nil {
		// 可以在此增加重试/回退策略
		return err
	}
	return nil
}

// SendHeartbeat 负责把心跳 POST 到服务端（类似 SendRegister）
func SendHeartbeat(hb *mode.HeartbeatRequest) ([]byte, error) {
	serverHost := viper.GetString("server.host")
	serverPort := viper.GetInt("server.port")
	proto := viper.GetString("server.protocol")
	path := viper.GetString("server.heartbeat_path")
	if serverHost == "" {
		serverHost = "localhost"
	}
	if proto == "" {
		proto = "http"
	}
	url := fmt.Sprintf("%s://%s:%d%s", proto, serverHost, serverPort, path)

	body, _ := json.Marshal(hb)
	timeout := time.Duration(viper.GetInt("server.timeout")) * time.Second
	if timeout == 0 {
		timeout = 10 * time.Second
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: timeout}
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server returned %d: %s", resp.StatusCode, string(respBody))
	}
	return respBody, nil
}

// hmacBase64Sign 用 key 对 msg 做 HMAC-SHA256 并返回 base64
func hmacBase64Sign(key, msg []byte) string {
	m := hmac.New(sha256.New, key)
	m.Write(msg)
	return base64.StdEncoding.EncodeToString(m.Sum(nil))
}

// metricsToInterfaceMap 确保 metrics 为 map[string]interface{}（为 json.Marshal 兼容）
func metricsToInterfaceMap(m map[string]interface{}) map[string]interface{} {
	return m
}

// ----------------- 采集静态 / 动态 信息（可替换更强的实现） -----------------

// CollectStaticInfo 采集一组静态信息（仅示例，建议使用 gopsutil 进一步丰富）
func CollectStaticInfo() map[string]interface{} {
	info := map[string]interface{}{}
	info["hostname"], _ = os.Hostname()
	info["os"] = runtime.GOOS
	info["cpu_count"] = runtime.NumCPU()

	// 获取内网IP
	if addrs, err := net.InterfaceAddrs(); err == nil {
		var internalIPs []string
		for _, addr := range addrs {
			if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
				if ipnet.IP.To4() != nil {
					internalIPs = append(internalIPs, ipnet.IP.String())
				}
			}
		}
		if len(internalIPs) > 0 {
			info["internal_ips"] = internalIPs
		}
	}

	// 获取系统详细信息（如ubuntu或centos）
	if runtime.GOOS == "linux" {
		// 尝试读取/etc/os-release文件
		if data, err := os.ReadFile("/etc/os-release"); err == nil {
			lines := strings.Split(string(data), "\n")
			for _, line := range lines {
				if strings.HasPrefix(line, "PRETTY_NAME=") {
					info["os_detail"] = strings.Trim(strings.TrimPrefix(line, "PRETTY_NAME="), "\"")
					break
				} else if strings.HasPrefix(line, "NAME=") && info["os_detail"] == "" {
					info["os_detail"] = strings.Trim(strings.TrimPrefix(line, "NAME="), "\"")
				}
			}
		}
	} else if runtime.GOOS == "darwin" {
		info["os_detail"] = "macOS"
	} else if runtime.GOOS == "windows" {
		info["os_detail"] = "Windows"
	}

	// 格式化配置信息为易读格式（cpu:8c mem:2g）
	cpuCount := runtime.NumCPU()
	if vm, err := mem.VirtualMemory(); err == nil {
		memGB := vm.Total / (1024 * 1024 * 1024) // 转换为GB
		info["config"] = fmt.Sprintf("cpu:%dc mem:%dg", cpuCount, memGB)
	} else {
		info["config"] = fmt.Sprintf("cpu:%dc", cpuCount)
	}

	// 内存和磁盘等用 gopsutil 获取更准确信息
	if vm, err := mem.VirtualMemory(); err == nil {
		info["total_memory_bytes"] = vm.Total
		info["total_memory_gb"] = vm.Total / (1024 * 1024 * 1024)
	}
	if parts, err := disk.Partitions(true); err == nil {
		disks := []map[string]interface{}{}
		for _, p := range parts {
			di := map[string]interface{}{
				"device":     p.Device,
				"mountpoint": p.Mountpoint,
				"fstype":     p.Fstype,
				"opts":       p.Opts,
			}
			if usage, err := disk.Usage(p.Mountpoint); err == nil {
				di["total_bytes"] = usage.Total
				di["free_bytes"] = usage.Free
				di["total_gb"] = usage.Total / (1024 * 1024 * 1024)
			}
			disks = append(disks, di)
		}
		info["disks"] = disks
	}
	return info
}

// CollectDynamicInfo 采集 CPU/内存/磁盘 使用率等（示例）
func CollectDynamicInfo() map[string]interface{} {
	d := map[string]interface{}{}
	// memory
	if vm, err := mem.VirtualMemory(); err == nil {
		d["memory_usage_percent"] = vm.UsedPercent
		d["memory_used_bytes"] = vm.Used
	}
	// disk - 总体用 root 分区展示
	if usage, err := disk.Usage("/"); err == nil {
		d["disk_usage_percent"] = usage.UsedPercent
		d["disk_used_bytes"] = usage.Used
	}
	// cpu (sleep 100ms采集，平衡准确性和延迟)
	if percents, err := cpu.Percent(100*time.Millisecond, false); err == nil && len(percents) > 0 {
		d["cpu_usage_percent"] = percents[0]
	} else {
		d["cpu_usage_percent"] = 0.0
	}

	// 网络 / 网卡 / 公网 IP 需要更多实现：可调用外部接口或 parse net interfaces
	return d
}
