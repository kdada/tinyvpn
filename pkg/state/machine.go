package state

// Event is the name of transition event. A valid event should not be a empty string.
type Event string

const (
	// EventNone is the original event before state machine starting
	EventNone Event = ""
)

// State is the name of state. A valid state should not be a empty string.
type State string

const (
	// StateNone is the original state before state machine starting
	StateNone State = ""
)

// Handler is a state handler for handling transition events
type Handler interface {
	// State returns the state of current handler
	State() State
	// EnterState is called when state machine transform from old state to current state.
	// If current state is the first state, event is EventNone.
	EnterState(m Machine, event Event, data interface{}) error
	// ExitState is called when state machine transform from current state to another state
	ExitState(m Machine, event Event, data interface{}) error
}

// Machine is a state machine to manage states and transitions
type Machine interface {
	// AddStateHandler adds state handler
	AddStateHandler(handlers ...Handler)
	// AddTransition adds a transition between two states. If a state does not have
	// handler, it does nothing when transform to the state.
	AddTransition(event Event, from State, to State)
	// Start starts the machine and transform to the state
	Start(state State, data interface{}) error
	// Trigger triggers an event and performs a transition
	Trigger(event Event, data interface{}) error
}
