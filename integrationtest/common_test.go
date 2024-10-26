package integrationtest

func toPtr[T any](s T) *T {
	return &s
}
