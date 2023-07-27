package daemon

type HardwareRequirements struct {
	minCPUCores                 int
	minRAM                      int
	minFreeSpace                int
	stopIfRequirementsAreNotMet bool
}

func (hr *HardwareRequirements) MinCPUCores() int {
	return hr.minCPUCores
}

func (hr *HardwareRequirements) MinRAM() int {
	return hr.minRAM
}

func (hr *HardwareRequirements) MinFreeSpace() int {
	return hr.minFreeSpace
}

func (hr *HardwareRequirements) StopIfRequirementsAreNotMet() bool {
	return hr.stopIfRequirementsAreNotMet
}
