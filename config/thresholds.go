package config

type Thresholds struct {
	LogTailLines      int64
	MonitorIntervalMs int
	MaxFailuresCount  int
}

var DefaultThresholds = Thresholds{
	LogTailLines:      100,
	MonitorIntervalMs: 60000, // 1 minute
	MaxFailuresCount:  10,
}
