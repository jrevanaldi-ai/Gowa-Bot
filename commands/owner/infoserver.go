package owner

import (
	"fmt"
	"os"
	"os/exec"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"time"

	"google.golang.org/protobuf/proto"

	"github.com/jrevanaldi-ai/gowa/proto/waE2E"
	"github.com/jrevanaldi-ai/gowa-bot/helper"
	"github.com/jrevanaldi-ai/gowa-bot/lib"
)


var InfoserverMetadata = &lib.CommandMetadata{
	Cmd:       "infoserver",
	Tag:       "owner",
	Desc:      "Tampilkan informasi lengkap server/system",
	Example:   ".infoserver",
	Hidden:    false,
	OwnerOnly: true,
	Alias:     []string{"sysinfo", "serverinfo", "info"},
}


func InfoserverHandler(ctx *lib.CommandContext) error {

	loadingMsg := "🔄 *Fetching server information...*\n\n_Mohon tunggu sebentar..._"


	sentResp, err := ctx.SendMessage(helper.CreateSimpleReply(loadingMsg, ctx.MessageID, ctx.Sender.String(), ctx.Chat.String()))
	if err != nil {
		return fmt.Errorf("failed to send loading message: %w", err)
	}


	var sentMsgID string
	respValue := reflect.ValueOf(sentResp)
	if respValue.Kind() == reflect.Struct {
		idField := respValue.FieldByName("ID")
		if idField.IsValid() {
			sentMsgID = idField.String()
		}
	}


	if sentMsgID == "" {
		info := collectServerInfo()
		_, err = ctx.SendMessage(helper.CreateSimpleReply(info, ctx.MessageID, ctx.Sender.String(), ctx.Chat.String()))
		return err
	}


	info := collectServerInfo()


	editMsg := ctx.Client.BuildEdit(ctx.Chat, sentMsgID, &waE2E.Message{
		ExtendedTextMessage: &waE2E.ExtendedTextMessage{
			Text: proto.String(info),
			ContextInfo: &waE2E.ContextInfo{
				StanzaID:    proto.String(ctx.MessageID),
				Participant: proto.String(ctx.Sender.String()),
			},
		},
	})
	_, err = ctx.Client.SendMessage(ctx.Ctx, ctx.Chat, editMsg)
	if err != nil {
		return fmt.Errorf("failed to send server info: %w", err)
	}

	return nil
}


func collectServerInfo() string {
	var info strings.Builder

	info.WriteString("*╭───⦿ SERVER INFO ⦿───*\n")
	info.WriteString("│\n")


	info.WriteString("┌─⦿ *System Information*\n")
	info.WriteString(fmt.Sprintf("│ • *OS:* %s\n", getOSInfo()))
	info.WriteString(fmt.Sprintf("│ • *Architecture:* %s\n", runtime.GOARCH))
	info.WriteString(fmt.Sprintf("│ • *Hostname:* %s\n", getHostname()))
	info.WriteString(fmt.Sprintf("│ • *Uptime:* %s\n", getSystemUptime()))
	info.WriteString("└──────────────\n\n")


	info.WriteString("┌─⦿ *CPU Information*\n")
	info.WriteString(fmt.Sprintf("│ • *Model:* %s\n", getCPUModel()))
	info.WriteString(fmt.Sprintf("│ • *Cores:* %d\n", runtime.NumCPU()))
	info.WriteString(fmt.Sprintf("│ • *Usage:* %s\n", getCPUUsage()))
	info.WriteString("└──────────────\n\n")


	info.WriteString("┌─⦿ *Memory (RAM)*\n")
	info.WriteString(fmt.Sprintf("│ • *Total:* %s\n", getTotalMemory()))
	info.WriteString(fmt.Sprintf("│ • *Used:* %s\n", getUsedMemory()))
	info.WriteString(fmt.Sprintf("│ • *Free:* %s\n", getFreeMemory()))
	info.WriteString(fmt.Sprintf("│ • *Usage:* %s\n", getMemoryUsagePercent()))
	info.WriteString("└──────────────\n\n")


	info.WriteString("┌─⦿ *Go Runtime*\n")
	info.WriteString(fmt.Sprintf("│ • *Version:* %s\n", runtime.Version()))
	info.WriteString(fmt.Sprintf("│ • *Goroutines:* %d\n", runtime.NumGoroutine()))
	info.WriteString(fmt.Sprintf("│ • *GC Count:* %d\n", getGCCount()))
	info.WriteString(fmt.Sprintf("│ • *Memory Used:* %s\n", getGoMemory()))
	info.WriteString(fmt.Sprintf("│ • *Memory Total:* %s\n", getGoTotalMemory()))
	info.WriteString("└──────────────\n\n")


	info.WriteString("┌─⦿ *Disk Usage*\n")
	info.WriteString(fmt.Sprintf("│ • *Total:* %s\n", getDiskTotal()))
	info.WriteString(fmt.Sprintf("│ • *Used:* %s\n", getDiskUsed()))
	info.WriteString(fmt.Sprintf("│ • *Free:* %s\n", getDiskFree()))
	info.WriteString(fmt.Sprintf("│ • *Usage:* %s\n", getDiskUsagePercent()))
	info.WriteString("└──────────────\n\n")


	info.WriteString("┌─⦿ *Network*\n")
	info.WriteString(fmt.Sprintf("│ • *IP Public:* %s\n", getPublicIP()))
	info.WriteString(fmt.Sprintf("│ • *IP Local:* %s\n", getLocalIP()))
	info.WriteString("└──────────────\n\n")


	info.WriteString("┌─⦿ *Bot Information*\n")
	info.WriteString(fmt.Sprintf("│ • *Uptime:* %s\n", getBotUptime()))
	info.WriteString(fmt.Sprintf("│ • *Start Time:* %s\n", getBotStartTime()))
	info.WriteString(fmt.Sprintf("│ • *GOMAXPROCS:* %d\n", runtime.GOMAXPROCS(0)))
	info.WriteString(fmt.Sprintf("│ • *Go Version:* %s\n", runtime.Version()))
	info.WriteString("└──────────────\n\n")


	info.WriteString("┌─⦿ *Process*\n")
	info.WriteString(fmt.Sprintf("│ • *PID:* %d\n", os.Getpid()))
	info.WriteString(fmt.Sprintf("│ • *Path:* %s\n", getExecutablePath()))
	info.WriteString(fmt.Sprintf("│ • *Threads:* %d\n", getThreadCount()))
	info.WriteString("└──────────────\n\n")

	info.WriteString("╰────────────────\n")

	return info.String()
}



func getOSInfo() string {
	switch runtime.GOOS {
	case "linux":
		return "🐧 Linux"
	case "windows":
		return "🪟 Windows"
	case "darwin":
		return "🍎 macOS"
	default:
		return runtime.GOOS
	}
}

func getHostname() string {
	hostname, err := os.Hostname()
	if err != nil {
		return "Unknown"
	}
	return hostname
}

func getSystemUptime() string {
	cmd := exec.Command("uptime", "-s")
	output, err := cmd.Output()
	if err != nil {

		return "N/A"
	}

	startTime, err := time.Parse("2006-01-02 15:04:05", strings.TrimSpace(string(output)))
	if err != nil {
		return "N/A"
	}

	uptime := time.Since(startTime)
	return formatDuration(uptime)
}

func getCPUModel() string {

	if runtime.GOOS == "linux" {
		cmd := exec.Command("grep", "-m", "1", "model name", "/proc/cpuinfo")
		output, err := cmd.Output()
		if err == nil {
			parts := strings.SplitN(string(output), ":", 2)
			if len(parts) == 2 {
				model := strings.TrimSpace(parts[1])

				if len(model) > 40 {
					model = model[:37] + "..."
				}
				return model
			}
		}
	}


	if runtime.GOOS == "darwin" {
		cmd := exec.Command("sysctl", "-n", "machdep.cpu.brand_string")
		output, err := cmd.Output()
		if err == nil {
			model := strings.TrimSpace(string(output))
			if len(model) > 40 {
				model = model[:37] + "..."
			}
			return model
		}
	}

	return fmt.Sprintf("%s (%s)", runtime.GOOS, runtime.GOARCH)
}

func getCPUUsage() string {
	if runtime.GOOS == "linux" {

		cmd := exec.Command("top", "-bn1")
		output, err := cmd.Output()
		if err == nil {
			lines := strings.Split(string(output), "\n")
			for _, line := range lines {
				if strings.Contains(line, "%Cpu") || strings.Contains(line, "%cpu") {

					parts := strings.Fields(line)
					for i, part := range parts {
						if part == "id," || part == "id" {
							if i > 0 {
								idle, _ := strconv.ParseFloat(parts[i-1], 64)
								usage := 100.0 - idle
								return fmt.Sprintf("%.1f%%", usage)
							}
						}
					}
				}
			}
		}
	}

	return "N/A"
}

func getTotalMemory() string {
	if runtime.GOOS == "linux" {
		cmd := exec.Command("grep", "MemTotal", "/proc/meminfo")
		output, err := cmd.Output()
		if err == nil {
			parts := strings.Fields(string(output))
			if len(parts) >= 2 {
				kb, _ := strconv.ParseInt(parts[1], 10, 64)
				return formatBytes(kb * 1024)
			}
		}
	}


	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)
	return formatBytes(int64(mem.TotalAlloc))
}

func getUsedMemory() string {
	if runtime.GOOS == "linux" {
		cmd := exec.Command("grep", "MemAvailable", "/proc/meminfo")
		output, err := cmd.Output()
		if err == nil {
			parts := strings.Fields(string(output))
			if len(parts) >= 2 {
				availableKb, _ := strconv.ParseInt(parts[1], 10, 64)


				cmd2 := exec.Command("grep", "MemTotal", "/proc/meminfo")
				output2, err2 := cmd2.Output()
				if err2 == nil {
					parts2 := strings.Fields(string(output2))
					if len(parts2) >= 2 {
						totalKb, _ := strconv.ParseInt(parts2[1], 10, 64)
						usedKb := totalKb - availableKb
						return formatBytes(usedKb * 1024)
					}
				}
			}
		}
	}

	return "N/A"
}

func getFreeMemory() string {
	if runtime.GOOS == "linux" {
		cmd := exec.Command("grep", "MemAvailable", "/proc/meminfo")
		output, err := cmd.Output()
		if err == nil {
			parts := strings.Fields(string(output))
			if len(parts) >= 2 {
				kb, _ := strconv.ParseInt(parts[1], 10, 64)
				return formatBytes(kb * 1024)
			}
		}
	}

	return "N/A"
}

func getMemoryUsagePercent() string {
	if runtime.GOOS == "linux" {
		cmd := exec.Command("free")
		output, err := cmd.Output()
		if err == nil {
			lines := strings.Split(string(output), "\n")
			for _, line := range lines {
				if strings.HasPrefix(line, "Mem:") {
					parts := strings.Fields(line)
					if len(parts) >= 3 {
						total, _ := strconv.ParseFloat(parts[1], 64)
						used, _ := strconv.ParseFloat(parts[2], 64)
						if total > 0 {
							percent := (used / total) * 100
							return fmt.Sprintf("%.1f%%", percent)
						}
					}
				}
			}
		}
	}

	return "N/A"
}

func getGCCount() uint32 {
	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)
	return mem.NumGC
}

func getGoMemory() string {
	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)
	return formatBytes(int64(mem.Alloc))
}

func getGoTotalMemory() string {
	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)
	return formatBytes(int64(mem.TotalAlloc))
}

func getDiskTotal() string {
	return getDiskInfo("total")
}

func getDiskUsed() string {
	return getDiskInfo("used")
}

func getDiskFree() string {
	return getDiskInfo("avail")
}

func getDiskUsagePercent() string {
	return getDiskInfo("percent")
}

func getDiskInfo(field string) string {
	cmd := exec.Command("df", "-h", "/")
	output, err := cmd.Output()
	if err != nil {
		return "N/A"
	}

	lines := strings.Split(string(output), "\n")
	if len(lines) >= 2 {

		fields := strings.Fields(lines[1])
		if len(fields) >= 5 {
			switch field {
			case "total":
				return fields[1]
			case "used":
				return fields[2]
			case "avail":
				return fields[3]
			case "percent":
				return strings.TrimPrefix(fields[4], "Use%") + " used"
			}
		}
	}

	return "N/A"
}

func getPublicIP() string {

	services := []string{
		"curl -s ifconfig.me",
		"curl -s api.ipify.org",
		"curl -s icanhazip.com",
	}

	for _, service := range services {
		cmd := exec.Command("sh", "-c", service)
		output, err := cmd.Output()
		if err == nil && len(output) > 0 {
			return strings.TrimSpace(string(output))
		}
	}

	return "N/A"
}

func getLocalIP() string {

	cmd := exec.Command("hostname", "-I")
	output, err := cmd.Output()
	if err == nil && len(output) > 0 {
		ips := strings.Fields(string(output))
		if len(ips) > 0 {
			return ips[0]
		}
	}

	return "127.0.0.1"
}

func getBotUptime() string {
	uptime := time.Since(startTime)
	return formatDuration(uptime)
}

func getBotStartTime() string {
	return startTime.Format("2006-01-02 15:04:05")
}

func getExecutablePath() string {
	path, err := os.Executable()
	if err != nil {
		return "Unknown"
	}


	if len(path) > 50 {
		path = "..." + path[len(path)-47:]
	}
	return path
}

func getThreadCount() int {
	return runtime.NumGoroutine()
}


func formatBytes(bytes int64) string {
	const (
		KB = 1024
		MB = 1024 * KB
		GB = 1024 * MB
		TB = 1024 * GB
	)

	switch {
	case bytes >= TB:
		return fmt.Sprintf("%.2f TB", float64(bytes)/TB)
	case bytes >= GB:
		return fmt.Sprintf("%.2f GB", float64(bytes)/GB)
	case bytes >= MB:
		return fmt.Sprintf("%.2f MB", float64(bytes)/MB)
	case bytes >= KB:
		return fmt.Sprintf("%.2f KB", float64(bytes)/KB)
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}

func formatDuration(d time.Duration) string {
	hours := int(d.Hours())
	minutes := int(d.Minutes()) % 60
	seconds := int(d.Seconds()) % 60

	if hours > 24 {
		days := hours / 24
		hours = hours % 24
		return fmt.Sprintf("%d hari %02d:%02d:%02d", days, hours, minutes, seconds)
	}

	return fmt.Sprintf("%02d:%02d:%02d", hours, minutes, seconds)
}


var startTime = time.Now()
