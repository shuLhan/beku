package beku

type packageState uint

const (
	packageStateNew packageState = 1 << iota
	packageStateLoad
	packageStateChange
	packageStateDirty
)
