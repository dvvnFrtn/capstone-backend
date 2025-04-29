package errs

import (
	"errors"
	"fmt"
	"runtime"
	"sort"
)

type (
	Op   string
	Code uint8
	Msg  string
)

const (
	Unexpected Code = iota
	RateLimit
	OTPExpired
	BadRequest
	NotFound
	Conflict
	Forbidden
	Unauthorize
	Internal
)

func (c Code) String() string {
	switch c {
	case Conflict:
		return "resource_already_exists"
	case OTPExpired:
		return "token_expired"
	case Unexpected:
		return "unexpected"
	case Internal:
		return "internal_error"
	case Unauthorize:
		return "unauthorize"
	case Forbidden:
		return "forbidden_action"
	case BadRequest:
		return "invalid_request"
	case NotFound:
		return "resource_not_found"
	case RateLimit:
		return "too_many_request"
	default:
		return "unknown_error"
	}
}

type Error struct {
	Op   Op
	Code Code
	Msg  Msg
	Err  error
}

func (e *Error) Error() string {
	return e.Err.Error()
}

func (e *Error) Unwrap() error {
	return e.Err
}

func CodeIs(err error, code Code) bool {
	var ce *Error
	if errors.As(err, &ce) {
		if ce.Code == code {
			return true
		}
	}
	return false
}

func New(args ...any) error {
	if len(args) == 0 {
		panic("call to errors.New with no arguments")
	}

	err := &Error{}
	for _, arg := range args {
		switch arg := arg.(type) {
		case Op:
			err.Op = arg
		case Code:
			err.Code = arg
		case Msg:
			err.Msg = arg
		case string:
			err.Err = errors.New(arg)
		case *Error:
			errcopy := *arg
			err.Err = &errcopy
		case error:
			err.Err = arg
		default:
			_, file, line, _ := runtime.Caller(1)
			return fmt.Errorf("_error.E: bad call from %s:%d %v, unknown type %T, value %v in error call", file, line, args, arg, arg)
		}
	}

	prev, ok := err.Err.(*Error)
	if !ok {
		return err
	}

	if err.Code == Unexpected {
		err.Code = prev.Code
		prev.Code = Unexpected
	}

	if prev.Msg != "" {
		err.Msg = prev.Msg
	}

	return err
}

func OpStack(err error) []string {
	type o struct {
		Op    string
		Order int
	}

	e := err
	i := 0
	var os []o

	for errors.Unwrap(e) != nil {
		var ce *Error
		if errors.As(e, &ce) {
			if ce.Op != "" {
				op := o{Op: string(ce.Op), Order: i}
				os = append(os, op)
			}
		}
		e = errors.Unwrap(e)
		i++
	}

	sort.Slice(os, func(i, j int) bool {
		return os[i].Order > os[j].Order
	})

	var ops []string
	for _, op := range os {
		ops = append(ops, op.Op)
	}

	return ops
}
