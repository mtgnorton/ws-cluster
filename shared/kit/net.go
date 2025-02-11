package kit

import (
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"time"
)

// GetServerIP 获取服务器主IP地址（IPv4优先）
func GetServerIP() (string, error) {
	// 获取所有网络接口
	interfaces, err := net.Interfaces()
	if err != nil {
		return "", fmt.Errorf("get interfaces failed: %v", err)
	}

	// 按优先级排序的候选IP列表
	var ipv4Candidates, ipv6Candidates []string

	for _, iface := range interfaces {
		// 排除无效接口的条件
		if iface.Flags&net.FlagUp == 0 { // 接口未启用
			continue
		}
		if iface.Flags&net.FlagLoopback != 0 { // 排除回环接口
			continue
		}
		if isVirtualInterface(iface.Name) { // 排除虚拟接口
			continue
		}

		// 获取接口地址
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}

		for _, addr := range addrs {
			ipNet, ok := addr.(*net.IPNet)
			if !ok {
				continue
			}

			ip := ipNet.IP
			if ip.IsLoopback() || ip.IsLinkLocalMulticast() || ip.IsLinkLocalUnicast() {
				continue
			}

			if ipv4 := ip.To4(); ipv4 != nil {
				ipv4Candidates = append(ipv4Candidates, ipv4.String())
			} else if ipv6 := ip.To16(); ipv6 != nil {
				ipv6Candidates = append(ipv6Candidates, ipv6.String())
			}
		}
	}

	// 选择最佳候选IP
	switch {
	case len(ipv4Candidates) > 0:
		return selectBestIP(ipv4Candidates), nil
	case len(ipv6Candidates) > 0:
		return selectBestIP(ipv6Candidates), nil
	default:
		return "", errors.New("no valid IP address found")
	}
}

// 判断是否为虚拟接口（根据命名模式）
func isVirtualInterface(name string) bool {
	virtualPrefixes := []string{"docker", "veth", "br-", "virbr", "lo", "tun", "kube"}
	for _, prefix := range virtualPrefixes {
		if strings.HasPrefix(name, prefix) {
			return true
		}
	}
	return false
}

// 选择最佳IP（根据RFC标准优先级）
func selectBestIP(ips []string) string {
	// 优先级排序：
	// 1. 私有地址优先（内网通信优先）
	// 2. 全局单播地址次之
	// 3. 其他地址
	for _, ip := range ips {
		if isPrivateIP(net.ParseIP(ip)) {
			return ip
		}
	}
	return ips[0] // 返回找到的第一个IP
}

// 判断是否为私有地址
func isPrivateIP(ip net.IP) bool {
	if ip4 := ip.To4(); ip4 != nil {
		// RFC 1918
		return ip4[0] == 10 ||
			(ip4[0] == 172 && ip4[1] >= 16 && ip4[1] <= 31) ||
			(ip4[0] == 192 && ip4[1] == 168)
	}
	// RFC 4193 (IPv6 ULA)
	return len(ip) == net.IPv6len && ip[0]&0xfe == 0xfc
}

// GetPublicIP 获取外部公网IP（通过第三方API查询）
func GetPublicIP() (string, error) {
	client := &http.Client{
		Timeout: 3 * time.Second,
	}

	resp, err := client.Get("https://api.ipify.org")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}
