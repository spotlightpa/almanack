package httpx

import (
	"slices"
	"net/http"
)

// Controller implements the pattern in
// https://choly.ca/post/go-experiments-with-handler/
type Controller func(w http.ResponseWriter, r *http.Request) http.Handler

func (c Controller) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if h := c(w, r); h != nil {
		h.ServeHTTP(w, r)
	}
}

// Middleware is any function that wraps an http.Handler returning a new http.Handler.
type Middleware = func(h http.Handler) http.Handler

// Stack is a slice of Middleware for use with Router.
type Stack []Middleware

func (stack Stack) Clone() Stack {
	return slices.Clone(stack)
}

// Push adds Middleware to end of the stack.
func (stack *Stack) Push(mw ...Middleware) {
	*stack = append(*stack, mw...)
}

// Push adds Middleware to end of the stack if cond is true.
func (stack *Stack) PushIf(cond bool, mw ...Middleware) {
	if cond {
		*stack = append(*stack, mw...)
	}
}

// AsMiddleware returns a Middleware
// which applies each of the members of the stack to its handlers.
func (stack Stack) AsMiddleware() Middleware {
	return func(h http.Handler) http.Handler {
		for i := len(stack) - 1; i >= 0; i-- {
			m := (stack)[i]
			h = m(h)
		}
		return h
	}
}

// Handler builds an http.Handler
// that applies all the middleware in the stack
// before calling h.
func (stack Stack) Handler(h http.Handler) http.Handler {
	h = stack.AsMiddleware()(h)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h.ServeHTTP(w, r)
	})
}

// HandlerFunc builds an http.Handler
// that applies all the middleware in the stack
// before calling f.
func (stack Stack) HandlerFunc(f http.HandlerFunc) http.Handler {
	return stack.Handler(f)
}

// Controller builds an http.Handler
// that applies all the middleware in the stack
// before calling c.
func (stack Stack) Controller(c Controller) http.Handler {
	return stack.Handler(c)
}
