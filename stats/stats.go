package stats

import (
	"log"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/load"
	"github.com/shirou/gopsutil/v3/mem"
)

type Stats struct {
	MemStats  *mem.VirtualMemoryStat
	DiskStats *disk.UsageStat
	CpuStats  []cpu.TimesStat // Typically contains one element when perCPU=false
	LoadStats *load.AvgStat
	TaskCount int
}

// Memory

func (s *Stats) MemUsedKb() uint64 {
	return s.MemStats.Used / 1024
}

func (s *Stats) MemUsedPercent() float64 {
	return s.MemStats.UsedPercent
}

func (s *Stats) MemAvailableKb() uint64 {
	return s.MemStats.Available / 1024
}

func (s *Stats) MemTotalKb() uint64 {
	return s.MemStats.Total / 1024
}

// Disk

func (s *Stats) DiskTotal() uint64 {
	return s.DiskStats.Total / 1024
}

func (s *Stats) DiskFree() uint64 {
	return s.DiskStats.Free / 1024
}

func (s *Stats) DiskUsed() uint64 {
	return s.DiskStats.Used / 1024
}

// CPU

func (s *Stats) CpuUsage() float64 {
	if len(s.CpuStats) == 0 {
		return 0.0
	}

	cpuStat := s.CpuStats[0] // aggregate values
	idle := cpuStat.Idle + cpuStat.Iowait
	nonIdle := cpuStat.User + cpuStat.Nice + cpuStat.System + cpuStat.Irq + cpuStat.Softirq + cpuStat.Steal
	total := idle + nonIdle

	if total == 0 {
		return 0.0
	}

	usage := (float64(nonIdle) / float64(total)) * 100
	return usage
}

// Aggregator

func GetStats() *Stats {
	mem := GetMemoryInfo()
	disk := GetDiskInfo()
	cpu := GetCpuStats()
	load := GetLoadAvg()

	return &Stats{
		MemStats:  mem,
		DiskStats: disk,
		CpuStats:  cpu,
		LoadStats: load,
	}
}

// Getters

func GetMemoryInfo() *mem.VirtualMemoryStat {
	memstats, err := mem.VirtualMemory()
	if err != nil {
		log.Printf("Error getting memory stats: %v", err)
		return &mem.VirtualMemoryStat{}
	}
	return memstats
}

func GetDiskInfo() *disk.UsageStat {
	usage, err := disk.Usage("/")
	if err != nil {
		log.Printf("Error getting disk usage: %v", err)
		return &disk.UsageStat{}
	}
	return usage
}

func GetCpuStats() []cpu.TimesStat {
	stats, err := cpu.Times(false) // aggregate across all CPUs
	if err != nil {
		log.Printf("Error getting CPU stats: %v", err)
		return []cpu.TimesStat{}
	}
	return stats
}

func GetLoadAvg() *load.AvgStat {
	loadavg, err := load.Avg()
	if err != nil {
		log.Printf("Error getting load average: %v", err)
		return &load.AvgStat{}
	}
	return loadavg
}

func GetCpuUsagePercent() float64 {
	first, _ := cpu.Times(false)
	time.Sleep(1 * time.Second)
	second, _ := cpu.Times(false)

	if len(first) == 0 || len(second) == 0 {
		return 0
	}

	a := first[0]
	b := second[0]

	idleDelta := (b.Idle + b.Iowait) - (a.Idle + a.Iowait)
	totalDelta := (b.User + b.Nice + b.System + b.Irq + b.Softirq + b.Steal + b.Idle + b.Iowait) -
		(a.User + a.Nice + a.System + a.Irq + a.Softirq + a.Steal + a.Idle + a.Iowait)

	if totalDelta == 0 {
		return 0
	}

	return (1.0 - float64(idleDelta)/float64(totalDelta)) * 100
}
