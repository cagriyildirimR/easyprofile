package profile

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"time"

	"github.com/cagriyildirimR/easyprofile/internal/portmanager"
)

type Config struct {
	// Server configuration
	Port         int           // Port for pprof server, default 6060
	OutputDir    string        // Base directory for profiles, default "profile/profiles_<timestamp>"
	Rate         int           // Rate of sampling, default 100
	GracePeriod  time.Duration // Grace period is the time to wait before starting to collect profiles, default 1 seconds
	OpenProfiles bool          // Whether to open profiles in browser after collection, default true

	// Profile configurations
	Heap *HeapProfileConfig
	CPU  *CPUProfileConfig

	// Internal state
	portManager *portmanager.PortManager
}

// HeapProfileConfig defines configuration for heap profiling
type HeapProfileConfig struct {
	Interval    time.Duration // Time between samples
	SampleCount int           // Number of samples to collect
}

// CPUProfileConfig defines configuration for CPU profiling
type CPUProfileConfig struct {
	Duration    time.Duration // Duration of each CPU profile
	SampleCount int           // Number of profiles to collect
}

func RunProfile(config *Config) {
	if config == nil {
		config = defaultProfileConfig()
	}

	if config.Port == 0 {
		config.Port = 6060
	}

	if config.OutputDir == "" {
		currentTime := time.Now()
		config.OutputDir = fmt.Sprintf("profile/profiles_%s", currentTime.Format("2006-01-02_15-04-05"))
	}

	if err := os.MkdirAll(config.OutputDir, 0755); err != nil {
		fmt.Printf("ERROR: Failed to create profile directory: %v\n", err)
		return
	}

	if config.Rate <= 0 {
		runtime.SetCPUProfileRate(100)
	} else {
		runtime.SetCPUProfileRate(config.Rate)
	}

	if config.portManager == nil {
		config.portManager = portmanager.New(21000)
	}

	go func() {
		fmt.Printf("INFO: Starting pprof server on address: %s\n", fmt.Sprintf(":%d", config.Port))
		if err := http.ListenAndServe(fmt.Sprintf(":%d", config.Port), nil); err != nil {
			fmt.Printf("ERROR: pprof server error: %v\n", err)
		}
	}()

	if config.Heap != nil {
		go collectHeapProfiles(config)
	}

	if config.CPU != nil {
		go collectCPUProfile(config)
	}
}

// defaultProfileConfig runs heap and cpu profiling with default configuration
// heap profiling is enabled with 10 second interval and 10 samples
// cpu profiling is enabled with 10 second duration and 10 samples
func defaultProfileConfig() *Config {
	return &Config{
		Port:         6060,
		OutputDir:    "profile/profiles",
		Rate:         100,
		GracePeriod:  time.Second,
		OpenProfiles: true,
		Heap: &HeapProfileConfig{
			Interval:    time.Second * 10,
			SampleCount: 10,
		},
		CPU: &CPUProfileConfig{
			Duration:    time.Second * 10,
			SampleCount: 10,
		},
	}
}

func collectHeapProfiles(config *Config) {
	heapDir := fmt.Sprintf("%s/heap", config.OutputDir)
	if err := os.MkdirAll(heapDir, 0755); err != nil {
		fmt.Printf("ERROR: Failed to create heap profile directory: %v\n", err)
		return
	}

	time.Sleep(config.GracePeriod)

	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	for i := 0; i < config.Heap.SampleCount; i++ {
		profileTime := time.Now().Format("2006-01-02_15-04-05")
		filename := fmt.Sprintf("%s/heap_%s.prof", heapDir, profileTime)

		resp, err := client.Get(fmt.Sprintf("http://localhost:%d/debug/pprof/heap", config.Port))
		if err != nil {
			fmt.Printf("ERROR: Failed to collect heap profile: %v\n", err)
			time.Sleep(config.Heap.Interval)
			continue
		}

		data, err := io.ReadAll(resp.Body)
		if err != nil {
			fmt.Printf("ERROR: Failed to read heap profile data: %v\n", err)
			resp.Body.Close()
			time.Sleep(config.Heap.Interval)
			continue
		}
		resp.Body.Close()

		if err := os.WriteFile(filename, data, 0644); err != nil {
			fmt.Printf("ERROR: Failed to write heap profile to file: %v\n", err)
		} else {
			fmt.Printf("INFO: Saved heap profile to %s\n", filename)
		}

		time.Sleep(config.Heap.Interval)
	}

	if config.OpenProfiles {
		openProfilesInBrowser(config, heapDir, "heap")
	}
}

func collectCPUProfile(config *Config) {
	cpuDir := fmt.Sprintf("%s/cpu", config.OutputDir)
	if err := os.MkdirAll(cpuDir, 0755); err != nil {
		fmt.Printf("ERROR: Failed to create CPU profile directory: %v\n", err)
		return
	}

	time.Sleep(config.GracePeriod)

	client := &http.Client{
		Timeout: config.CPU.Duration + 3*time.Second,
	}

	for i := 0; i < config.CPU.SampleCount; i++ {
		profileTime := time.Now().Format("2006-01-02_15-04-05")
		filename := fmt.Sprintf("%s/cpu_%s.prof", cpuDir, profileTime)

		resp, err := client.Get(fmt.Sprintf("http://localhost:%d/debug/pprof/profile?seconds=%d", config.Port, int(config.CPU.Duration.Seconds())))
		if err != nil {
			fmt.Printf("ERROR: Failed to collect CPU profile: %v\n", err)
			continue
		}

		data, err := io.ReadAll(resp.Body)
		if err != nil {
			fmt.Printf("ERROR: Failed to read CPU profile data: %v\n", err)
			resp.Body.Close()
			continue
		}
		resp.Body.Close()

		if err := os.WriteFile(filename, data, 0644); err != nil {
			fmt.Printf("ERROR: Failed to write CPU profile to file: %v\n", err)
		} else {
			fmt.Printf("INFO: Saved CPU profile to %s\n", filename)
		}
	}

	if config.OpenProfiles {
		openProfilesInBrowser(config, cpuDir, "cpu")
	}
}

func openProfilesInBrowser(config *Config, profileDir string, profileType string) {
	port, err := config.portManager.GetPort()
	if err != nil {
		fmt.Printf("ERROR: Failed to get available port for pprof web UI: %v\n", err)
		return
	}

	cmd := exec.Command("go", "tool", "pprof", "-http", fmt.Sprintf(":%d", port), fmt.Sprintf("%s/*", profileDir))

	if err := cmd.Start(); err != nil {
		fmt.Printf("ERROR: Failed to open %s profiles in browser: %v\n", profileType, err)
		return
	}

	fmt.Printf("INFO: Opening %s profiles in browser at http://localhost:%d\n", profileType, port)

	go func() {
		cmd.Wait()
	}()
}
