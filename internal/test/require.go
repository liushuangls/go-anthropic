package test

// Requirements provides assertion methods around the
// TestingT interface.
type Requirements struct {
	t TestingT
}

// New makes a new Requirements object for the specified TestingT.
func NewRequire(t TestingT) *Requirements {
	return &Requirements{
		t: t,
	}
}

// TestingT is an interface wrapper around *testing.T
type TestingT interface {
	Errorf(format string, args ...interface{})
	FailNow()
}

type tHelper interface {
	Helper()
}

func (a *Requirements) Equal(expected interface{}, actual interface{}, msgAndArgs ...interface{}) {
	if h, ok := a.t.(tHelper); ok {
		h.Helper()
	}
	Equal(a.t, expected, actual, msgAndArgs...)
}

func (a *Requirements) NoError(err error, msgAndArgs ...interface{}) {
	if h, ok := a.t.(tHelper); ok {
		h.Helper()
	}
	NoError(a.t, err, msgAndArgs...)
}

func (a *Requirements) Error(err error, msgAndArgs ...interface{}) {
	if h, ok := a.t.(tHelper); ok {
		h.Helper()
	}
	Error(a.t, err, msgAndArgs...)
}

func (a *Requirements) Contains(s interface{}, contains interface{}, msgAndArgs ...interface{}) {
	if h, ok := a.t.(tHelper); ok {
		h.Helper()
	}
	Contains(a.t, s, contains, msgAndArgs...)
}

func (a *Requirements) True(value bool, msgAndArgs ...interface{}) {
	if h, ok := a.t.(tHelper); ok {
		h.Helper()
	}
	True(a.t, value, msgAndArgs...)
}

func (a *Requirements) Nil(object interface{}, msgAndArgs ...interface{}) {
	if h, ok := a.t.(tHelper); ok {
		h.Helper()
	}
	Nil(a.t, object, msgAndArgs...)
}

func (a *Requirements) NotNil(object interface{}, msgAndArgs ...interface{}) {
	if h, ok := a.t.(tHelper); ok {
		h.Helper()
	}
	NotNil(a.t, object, msgAndArgs...)
}

func (a *Requirements) ErrorIs(err error, target error, msgAndArgs ...interface{}) {
	if h, ok := a.t.(tHelper); ok {
		h.Helper()
	}
	ErrorIs(a.t, err, target, msgAndArgs...)
}
