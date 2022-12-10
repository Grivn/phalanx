package config

import "time"

type PhalanxConf struct {
	OligarchID  uint64        `json:"oligarch_id" yaml:"oligarch_id"`
	IsByzantine bool          `json:"is_byzantine" yaml:"is_byzantine"`
	IsSnapping  bool          `json:"is_snapping" yaml:"is_snapping"`
	NodeID      uint64        `json:"node_id" yaml:"node_id"`
	NodeCount   int           `json:"node_count" yaml:"node_count"`
	Timeout     time.Duration `json:"timeout" yaml:"timeout"`
	Multi       int           `json:"multi" yaml:"multi"`
	MemSize     int           `json:"mem_size" yaml:"mem_size"`
	CommandSize int           `yaml:"command_size" yaml:"command_size"`
	Selected    uint64        `json:"selected" yaml:"selected"`
}
