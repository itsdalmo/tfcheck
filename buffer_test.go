package tfcheck_test

import (
	"reflect"
	"testing"

	tfcheck "github.com/itsdalmo/tfcheck"
)

func TestBuffer(t *testing.T) {
	b := &tfcheck.Buffer{}

	_, err := b.Write([]byte("a\n"))
	eq(t, nil, err)
	eq(t, "a\n", b.String())
	eq(t, []string{"a\n"}, b.Lines())
	eq(t, []string{"a\n"}, b.Tail(1))

	_, err = b.Write([]byte("b"))
	eq(t, nil, err)
	eq(t, "a\nb", b.String())
	eq(t, []string{"a\n", "b"}, b.Lines())
	eq(t, []string{"b"}, b.Tail(1))

	_, err = b.Write([]byte("\n"))
	eq(t, nil, err)
	eq(t, "a\nb\n", b.String())
	eq(t, []string{"a\n", "b\n"}, b.Lines())
	eq(t, []string{"b\n"}, b.Tail(1))
	eq(t, []string{"a\n", "b\n"}, b.Tail(2))

	// Does not break when N > number of lines
	eq(t, []string{"a\n", "b\n"}, b.Tail(100))
}

func eq[T any](t *testing.T, want, got T) {
	t.Helper()

	if !reflect.DeepEqual(want, got) {
		t.Errorf("objects are not equal:\nwant: %#v\ngot:  %#v\n", want, got)
	}
}
