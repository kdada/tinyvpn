package state

import "fmt"

// BaseHandler is a state handler for handling transition events
type BaseHandler struct {
	state State
}

// NewBaseHandler creates a state handler
func NewBaseHandler(state State) *BaseHandler {
	return &BaseHandler{state}
}

// State returns the state of current handler
func (h *BaseHandler) State() State {
	return h.state
}

// EnterState is called when state machine transform from old state to current state.
// If current state is the first state, event is empty.
func (h *BaseHandler) EnterState(m Machine, event Event, data interface{}) error {
	return nil
}

// ExitState is called when state machine transform from current state to another state
func (h *BaseHandler) ExitState(m Machine, event Event, data interface{}) error {
	return nil
}

// Transition stores transitions from a state
type Transition struct {
	// State is the current state
	State State
	// Events map a event to next state
	Events map[Event]State
}

// NewTransition Create transition
func NewTransition(state State) *Transition {
	return &Transition{
		State:  state,
		Events: make(map[Event]State),
	}
}

// AddEvent adds a event to current State
func (t *Transition) AddEvent(event Event, to State) {
	t.Events[event] = to
}

// NextState gets the next state by event
func (t *Transition) NextState(event Event) (State, error) {
	state, ok := t.Events[event]
	if !ok {
		return StateNone, fmt.Errorf("state %s has no event %s", t.State, event)
	}
	return state, nil
}

// BaseMachine is a state machine to manage states and transitions
type BaseMachine struct {
	State       State
	Handlers    map[State]Handler
	Transitions map[State]*Transition
}

// NewBaseMachine creates a state machine
func NewBaseMachine() *BaseMachine {
	return &BaseMachine{
		State:       StateNone,
		Handlers:    make(map[State]Handler),
		Transitions: make(map[State]*Transition),
	}
}

// AddStateHandler adds state handler
func (m *BaseMachine) AddStateHandler(handlers ...Handler) {
	for _, h := range handlers {
		m.Handlers[h.State()] = h
	}
}

// AddTransition adds a transition between two states. If a state does not have
// handler, it does nothing when transform to the state.
func (m *BaseMachine) AddTransition(event Event, from State, to State) {
	transition, ok := m.Transitions[from]
	if !ok {
		transition = NewTransition(from)
	}
	transition.AddEvent(event, to)
	m.Transitions[from] = transition
}

// Start start the machine and transform to the state
func (m *BaseMachine) Start(state State, data interface{}) error {
	if m.State != StateNone {
		return fmt.Errorf("state machine has already started")
	}
	m.State = state
	handler, ok := m.Handlers[state]
	if ok {
		return handler.EnterState(m, EventNone, data)
	}
	return nil
}

// Trigger triggers an event and performs a transition
func (m *BaseMachine) Trigger(event Event, data interface{}) error {
	if m.State == StateNone {
		return fmt.Errorf("state machine does not start")
	}
	transition, ok := m.Transitions[m.State]
	if !ok {
		return fmt.Errorf("state %s has no transition", m.State)
	}
	next, err := transition.NextState(event)
	if err != nil {
		return err
	}
	// transform
	handler, ok := m.Handlers[m.State]
	if ok {
		if err = handler.ExitState(m, event, data); err != nil {
			return err
		}
	}
	handler, ok = m.Handlers[next]
	if ok {
		if err = handler.EnterState(m, event, data); err != nil {
			return err
		}
	}
	m.State = next
	return nil
}
