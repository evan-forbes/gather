package manage

import "fmt"

// SignalWrap uses ErrorSignal to attach information
// to the error provided.
func SignalWrap(err error, act Action) ErrorSignal {
	return ErrorSignal{err, act}
}

// Error fulfills the error interface while
// representing the added information in ErrorSignal
func (r ErrorSignal) Error() string {
	return fmt.Sprintf("%s worthy= %s", r.Action.String(), r.error.Error())
}

// ErrorSignal is a wrapper around errors that contain
// extra information encoded by an Action const
type ErrorSignal struct {
	error
	Action Action
}

// Action is used to add information to errors
// to let the Controller type know when to reboot
// the rebootable system
type Action int

const (
	SHUTDOWN   Action = 1
	REBOOT     Action = 2
	WRITELOCAL Action = 3
)

// String fulfills the Stringer interface for
// OfferAction
func (o Action) String() string {
	names := [...]string{
		"na",
		"SHUTDOWN", "REBOOT",
	}
	if o > REBOOT || o < SHUTDOWN {
		return "Unregistered OfferAction"
	}

	return names[o]
}
