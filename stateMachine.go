package stateful

type (
	// StateMachine handles the state of the StatefulObject
	StateMachine struct {
		StatefulObject  Stateful
		transitionRules TransitionRules
	}
)

// AddTransition adds a transition to the state machine.
func (sm *StateMachine) AddTransition(
	transition Transition,
	sourceStates States,
	destinationStates States,
) {
	sm.transitionRules = append(
		sm.transitionRules,
		TransitionRule{
			SourceStates:      sourceStates,
			Transition:        transition,
			DestinationStates: destinationStates,
		},
	)
}

// GetTransitionRules returns all transitionRules in the state machine
func (sm StateMachine) GetTransitionRules() TransitionRules {
	return sm.transitionRules
}

// GetAllStates returns all known and possible states by the state machine
func (sm StateMachine) GetAllStates() States {
	states := States{}
	keys := make(map[State]bool)

	for _, transitionRule := range sm.transitionRules {
		for _, state := range append(transitionRule.SourceStates, transitionRule.DestinationStates...) {
			if _, ok := keys[state]; !ok {
				keys[state] = true
				if !state.IsWildCard() {
					states = append(states, state)
				}
			}
		}
	}
	return states
}

// Run runs the state machine with the given transition.
// If the transition
func (sm StateMachine) Run(
	transition Transition,
	transitionArguments TransitionArguments,
) error {
	transitionRule := sm.transitionRules.Find(transition)
	if transitionRule == nil {
		return NewTransitionRuleNotFoundError(transition)
	}

	if !transitionRule.IsAllowedToRun(sm.StatefulObject.State()) {
		return NewCannotRunFromStateError(sm, *transitionRule)
	}

	newState, err := transition(transitionArguments)
	if err != nil {
		return err
	}

	if !transitionRule.IsAllowedToTransfer(newState) {
		return NewCannotTransferToStateError(newState)
	}

	err = sm.StatefulObject.SetState(newState)
	if err != nil {
		return err
	}
	return nil
}

func (sm StateMachine) GetAvailableTransitions() Transitions {
	transitions := Transitions{}
	for _, transitionRule := range sm.transitionRules {
		if transitionRule.IsAllowedToRun(sm.StatefulObject.State()) {
			transitions = append(transitions, transitionRule.Transition)
		}
	}
	return transitions
}
