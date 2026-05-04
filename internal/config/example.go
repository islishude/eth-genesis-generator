package config

// ExampleYAML returns the starter genesis.yaml written by the init command.
func ExampleYAML() string {
	return RenderInitYAML(DefaultInitConfig())
}
