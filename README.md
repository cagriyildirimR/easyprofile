# EasyProfile

A convenience tool for Go profiles

## Quick Start

```go
package main

import (
	"github.com/cagriyildirimR/easyprofile"
	"time"
)

func main() {
	// Use default configuration
	go profile.RunProfile(nil)
	
	// Your application code here
	// ...
	
	// Give enough time for profiling to complete
	time.Sleep(30 * time.Second)
}
```

## Custom Configuration

```go
package main

import (
	"github.com/cagriyildirimR/easyprofile"
	"time"
)

func main() {
	config := &profile.ProfileConfig{
		Port:         7070,                // Custom port for pprof server
		OutputDir:    "custom_profiles",   // Custom output directory
		Rate:         200,                 // Custom sampling rate
		GracePeriod:  2 * time.Second,     // Wait 2 seconds before starting profiling
		OpenProfiles: true,                // Open profiles in browser after finished
		
		// CPU profile configuration
		CPU: &profile.CPUProfileConfig{
			Duration:    5 * time.Second,  // 5-second long CPU profiles
			SampleCount: 3,                // Collect 3 CPU profiles
		},
		
		// Heap profile configuration
		Heap: &profile.HeapProfileConfig{
			Interval:    3 * time.Second,  // 3 seconds between heap profiles
			SampleCount: 5,                // Collect 5 heap profiles
		},
	}
	
	go profile.RunProfile(config)
	
	// Your application code here
	// ...
	
	// Give enough time for profiling to complete
	time.Sleep(30 * time.Second)
}
```

## License

See the [LICENSE](LICENSE) file for details.