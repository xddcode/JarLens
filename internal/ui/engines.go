package ui

import (
	"github.com/xddcode/jarlens/internal/decompile"
	"github.com/xddcode/jarlens/internal/engine/cfr"
	"github.com/xddcode/jarlens/internal/engine/procyon"
	"github.com/xddcode/jarlens/internal/engine/vineflower"
)

var (
	decompilerEngineIDs    = []string{"vineflower", "cfr", "procyon"}
	decompilerEngineLabels = []string{"Vineflower 1.12.0", "CFR 0.152", "Procyon 0.6.0"}
)

func (m *Model) initializeEngines() {
	if m.engine == nil {
		m.engines = []decompile.Engine{vineflower.New(), cfr.New(), procyon.New()}
		m.engine = m.engines[0]
	}
	if len(m.engines) == 0 {
		m.engines = []decompile.Engine{m.engine}
	}
	m.selectEngine(m.settings.Engine)
}

func (m *Model) selectEngine(id string) {
	index := decompilerEngineIndex(id)
	if index >= 0 && index < len(m.engines) {
		m.engine = m.engines[index]
	}
}

func (m *Model) closeEngines() {
	for _, engine := range m.engines {
		if closer, ok := engine.(interface{ Close() error }); ok {
			_ = closer.Close()
		}
	}
}

func decompilerEngineIndex(id string) int {
	for index, candidate := range decompilerEngineIDs {
		if candidate == id {
			return index
		}
	}
	return -1
}
