package task

import "slices"

type StateTransitionMap map[State][]State

var StateTransitions = StateTransitionMap{
	Pending: {
		Scheduled,
	},
	Scheduled: {
		Scheduled,
		Running,
		Failed,
	},
	Running: {
		Running,
		Completed,
		Failed,
	},
	Completed: {},
	Failed:    {},
}

func ValidStateTransition(src State, dest State) bool {
	if destStates, ok := StateTransitions[src]; ok {
		return slices.Contains(destStates, dest)
	}
	return false
}
