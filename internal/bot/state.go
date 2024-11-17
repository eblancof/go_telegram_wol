package bot

type AddDeviceState struct {
	Name  string
	Stage string
}

type ModifyDeviceState struct {
	DeviceName string
	Field      string
}

var (
	addDeviceStates    = make(map[int64]*AddDeviceState)
	modifyDeviceStates = make(map[int64]*ModifyDeviceState)
	buttonMessages     = make(map[int64][]int)
)
