package iterx

import "iter"

// ErrorChildren does a shallow unwrapping of multierrors.
func ErrorChildren(err error) iter.Seq[error] {
	return func(yield func(error) bool) {
		if err == nil {
			return
		}
		type Unwrapper interface {
			Unwrap() []error
		}
		if unwrapper, ok := err.(Unwrapper); ok {
			for _, suberr := range unwrapper.Unwrap() {
				if suberr != nil && !yield(suberr) {
					return
				}
			}
		} else {
			if !yield(err) {
				return
			}
		}
	}
}
