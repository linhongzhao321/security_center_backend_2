package monitor

import (
	"testing"
)

func TestNewConfig(t *testing.T) {
	config, err := NewConfig()
	if err != nil {
		t.Error(err)
	}
	t.Log(`projects count:`, len(config.Projects))
}
