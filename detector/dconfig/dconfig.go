package dconfig

// Detector config is used for parsing configuration data from the config file
type DetectorConfig struct {
	Name          string `json:"name"`
	Type          string `json:"type"`
	ModelFile     string `json:"model_file"`
	LabelFile     string `json:"label_file"`
	NumThreads    int    `json:"num_threads"`
	NumConcurrent int    `json:"num_concurrent"`
	HWAccel       bool   `json:"hw_accel"`
}
