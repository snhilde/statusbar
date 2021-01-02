package statusbar

import (
	"testing"
	"github.com/snhilde/statusbar/v5/sbbattery"
	"github.com/snhilde/statusbar/v5/sbcputemp"
	"github.com/snhilde/statusbar/v5/sbcpuusage"
	"github.com/snhilde/statusbar/v5/sbdisk"
	"github.com/snhilde/statusbar/v5/sbfan"
	"github.com/snhilde/statusbar/v5/sbload"
	"github.com/snhilde/statusbar/v5/sbnetwork"
	"github.com/snhilde/statusbar/v5/sbram"
	"github.com/snhilde/statusbar/v5/sbtime"
	"github.com/snhilde/statusbar/v5/sbtodo"
	"github.com/snhilde/statusbar/v5/sbvolume"
	"github.com/snhilde/statusbar/v5/sbweather"
)

func TestStatusbar(t *testing.T) {
	// Build and run a new statusbar to make sure everything builds as expected.
	bar := New()

	bar.Append(sbbattery.New([3]string{"#17A130", "#BB4F2E", "#A1273E"}), 30)
	bar.Append(sbcputemp.New([3]string{"#8FFFFF", "#BB4F2E", "#A1273E"}), 1)
	bar.Append(sbcpuusage.New([3]string{"#FFFFFF", "#BB4F2E", "#A1273E"}), 1)
	bar.Append(sbdisk.New([]string{"/"}, [3]string{"#FFFFFF", "#BB4F2E", "#A1273E"}), 5)
	bar.Append(sbfan.New([3]string{"#FFFFFF", "#BB4F2E", "#A1273E"}), 1)
	bar.Append(sbload.New([3]string{"#434852", "#BB4F2E", "#A1273E"}), 1)

	bar.Split()

	bar.Append(sbnetwork.New([]string{"interface"}, [3]string{"#FFFFFF", "#BB4F2E", "#A1273E"}), 1)
	bar.Append(sbram.New([3]string{"#FFFFFF", "#BB4F2E", "#A1273E"}), 5)
	bar.Append(sbtime.New("Jan 2 - 03:04", [3]string{"#FFFFFF", "#BB4F2E", "#A1273E"}), 1)
	bar.Append(sbtodo.New("/home/user/.TODO", [3]string{"#F1EA6B", "#BB4F2E", "#A1273E"}), 5)
	bar.Append(sbvolume.New("Master", [3]string{"#FFFFFF", "#BB4F2E", "#A1273E"}), 1)
	bar.Append(sbweather.New("90210", [3]string{"#FFFFFF", "#BB4F2E", "#A1273E"}), 30*60)

	bar.EnableRESTAPI(1234)

	bar.Run()
}
