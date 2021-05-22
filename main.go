package main

import (
	"fmt"
	"math"
	"runtime"
	"os/exec"
	"bytes"
	"strings"
	"strconv"
	"github.com/kbinani/screenshot"
	"github.com/shirou/gopsutil/host"
	"github.com/shirou/gopsutil/mem"
	"github.com/ricochet2200/go-disk-usage/du"
	"github.com/xyproto/cdrom"
)

const float64EqualityThreshold = 1e-9
const KB = uint64(1024)
const GB = KB * KB * KB

func areFloatsEqual(a, b float64) bool {
	return math.Abs(a - b) <= float64EqualityThreshold
}

func executeBashCommand(command string) (error, string, string) {
	var stdout, stderr bytes.Buffer
	cmd := exec.Command("bash", "-c", command)
    cmd.Stdout = &stdout
    cmd.Stderr = &stderr
    err := cmd.Run()
    return err, strings.TrimSpace(stdout.String()), stderr.String()
}

type testFunctionReturn struct {
	definitelyVM, definitelyUser bool // default is always false
	percentageVM, percentageUser float64 // default is 0
}

func compineTestFunctionReturnResults(a, b testFunctionReturn) testFunctionReturn {
	return testFunctionReturn{
		definitelyVM: (a.definitelyVM || b.definitelyVM),
		definitelyUser: (a.definitelyUser || b.definitelyUser),
		percentageVM: (a.percentageVM + b.percentageVM),
		percentageUser: (a.percentageUser + b.percentageUser)}
}

func testForMultipleScreens() testFunctionReturn {
	if screenshot.NumActiveDisplays() > 1 {
		return testFunctionReturn{definitelyUser: true}
	}

	return testFunctionReturn{}
}

func testAspectRatio() testFunctionReturn {
	bounds := screenshot.GetDisplayBounds(0) // we only test for screen 0
	ratio := float64(bounds.Dx()) / float64(bounds.Dy())

	if areFloatsEqual(ratio, float64(16) / float64(10)) ||
		areFloatsEqual(ratio, float64(16) / float64(9)) ||
		areFloatsEqual(ratio, float64(4) / float64(3)) ||
		areFloatsEqual(ratio, float64(3) / float64(2)) ||
		areFloatsEqual(ratio, float64(32) / float64(9)) {
			return testFunctionReturn{percentageUser: 0.9}
	}

	return testFunctionReturn{definitelyVM: true}
}

// low and ultra high uptimes are sus
func testUptime() testFunctionReturn {
	uptime,_ := host.Uptime()

	// if uptime is lower than 5 min
	if uptime < 10*60 {
		return testFunctionReturn{percentageVM: 0.7}
	}

	// if uptime longer than 2 days liekly suspicious
	if uptime > 2*24*60*60 {
		return testFunctionReturn{percentageVM: 0.4}
	}

	// within limits normal user behavior
	return testFunctionReturn{percentageUser: 0.6}
}

func getDiskUsage() *du.DiskUsage {
	usage := du.NewDiskUsage("/")

	// For windows we have to do something different
	if runtime.GOOS == "windows" {
		usage = du.NewDiskUsage("C:\\")
	}

	return usage
}

func testAvialableDiskSpace() testFunctionReturn {
	usage := getDiskUsage()

	// if it's less than around 200 GB definitly a VM
	if usage.Available() / (KB*KB) < uint64(200000) {
		return testFunctionReturn{definitelyVM: true}
	}

	return testFunctionReturn{percentageUser: 0.7}
}

func testRamVsDiskSpace() testFunctionReturn {
	usage := getDiskUsage()
	v, _ := mem.VirtualMemory()

	// if there is more than 7.5% ram then it's most liekly a VM
	if float64(v.Total) / float64(usage.Available()) > 0.075 {
		return testFunctionReturn{definitelyVM: true}
	}

	return testFunctionReturn{percentageUser: 0.5}
}

func testRamShouldBeWithEvenGBRamSlotsOnly() testFunctionReturn {
	v, _ := mem.VirtualMemory()

	ramGb := v.Total / GB

	if ramGb % 2 != 0 {
		return testFunctionReturn{definitelyVM: true}
	}
	
	return testFunctionReturn{percentageUser: 0.2}
}

func testRamShouldBeWithGBRamSlotsOnly() testFunctionReturn {
	v, _ := mem.VirtualMemory()
	
	ramGb := v.Total / GB

	// if the ram is not Divisible by GB then it's a wired configured VM
	// will trip if only 512 MB ram
	if (ramGb * GB) - v.Total != 0 {
		return testFunctionReturn{definitelyVM: true}
	}

	return testFunctionReturn{percentageUser: 0.2}
}

// only for linux, test for vbox/VMWare kernelModules
func testForCommonKernelModules() testFunctionReturn {
	err, out, _ := executeBashCommand("find /lib/modules/$(uname -r) -type f -name '*.ko*' 2>/dev/null | grep \"vboxguest\\|/vme/\" | wc -l")
	suspectKernelModules, _ := strconv.Atoi(out)
	if err == nil  {
		if suspectKernelModules > 0 {
			return testFunctionReturn{definitelyVM: true}
		} else {
			return testFunctionReturn{percentageUser: 0.8}
		}

	}

	// windows & mac
	return testFunctionReturn{}
}

func testForCDRomDrive() testFunctionReturn {
	_, err := cdrom.New()

	// if no cdrom drive is present, probably
	if err != nil && err.Error() == "no such file or directory" {
		return testFunctionReturn{percentageUser: 0.8, percentageVM: 0.2}
	}

	return testFunctionReturn{percentageUser: 0.2, percentageVM: 0.8}
}

func isEnvironementAVM() bool {
	testFunctions := []func () testFunctionReturn {testForMultipleScreens, testAspectRatio, testForCommonKernelModules, testUptime, testAvialableDiskSpace, testRamShouldBeWithGBRamSlotsOnly, testRamVsDiskSpace, testForCDRomDrive}

	combined := testFunctionReturn{}

	for i := 0; i < len(testFunctions); i++ {
		current := testFunctions[i]()

		if current.definitelyUser {
			return false
		}

		if current.definitelyVM {
			return true
		}

		combined = compineTestFunctionReturnResults(combined, current)
	}

	return combined.percentageVM > combined.percentageUser
}

func main() {

	fmt.Println(isEnvironementAVM())

}