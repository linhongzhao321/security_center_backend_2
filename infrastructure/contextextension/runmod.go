package contextextension

import "fmt"

type RunMod string

const RunModLocal RunMod = `local`
const RunModRelease RunMod = `release`

var EnabledRunMods = []RunMod{RunModLocal, RunModRelease}

func (mod *RunMod) Set(v string) error {
	m := RunMod(v)
	switch m {
	case RunModLocal, RunModRelease:
		*mod = m
		return nil
	default:
		return fmt.Errorf(`must be one of %v`, EnabledRunMods)
	}
}

func (mod *RunMod) String() string {
	return string(*mod)
}
func (mod *RunMod) Type() string {
	return "RunMod"
}
