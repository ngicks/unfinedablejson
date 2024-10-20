package sliceund

import (
	"database/sql"
	"testing"

	"github.com/ngicks/und/internal/testcase"
	"github.com/ngicks/und/option"
	"gotest.tools/v3/assert"
)

func TestUnd(t *testing.T) {
	testcase.TestUnd_non_addressable(
		t,
		Defined[int](155),
		Null[int](),
		Undefined[int](),
		155,
		"155",
	)
}

// Tests for New-like function, e.g. FromPointer, WrapPointer
func TestUnd_new_functions(t *testing.T) {
	num := 15
	t.Run("FromPointer", func(t *testing.T) {
		fromNonNil := FromPointer(&num)
		assert.Equal(t, 15, fromNonNil.Value())
		assert.Equal(t, false, fromNonNil.IsUndefined())
		assert.Equal(t, false, fromNonNil.IsNull())
		assert.Equal(t, true, fromNonNil.IsDefined())
		fromNil := FromPointer((*int)(nil))
		assert.Equal(t, 0, fromNil.Value())
		assert.Equal(t, true, fromNil.IsUndefined())
		assert.Equal(t, false, fromNil.IsNull())
		assert.Equal(t, false, fromNil.IsDefined())
	})
	t.Run("WrapPointer", func(t *testing.T) {
		fromNonNil := WrapPointer(&num)
		assert.Equal(t, &num, fromNonNil.Value())
		assert.Equal(t, false, fromNonNil.IsUndefined())
		assert.Equal(t, false, fromNonNil.IsNull())
		assert.Equal(t, true, fromNonNil.IsDefined())
		fromNil := WrapPointer((*int)(nil))
		assert.Equal(t, (*int)(nil), fromNil.Value())
		assert.Equal(t, true, fromNil.IsUndefined())
		assert.Equal(t, false, fromNil.IsNull())
		assert.Equal(t, false, fromNil.IsDefined())
	})
	t.Run("FromOption", func(t *testing.T) {
		undefined := FromOption(option.None[option.Option[int]]())
		assert.Equal(t, 0, undefined.Value())
		assert.Equal(t, true, undefined.IsUndefined())
		assert.Equal(t, false, undefined.IsNull())
		assert.Equal(t, false, undefined.IsDefined())
		null := FromOption(option.Some(option.None[int]()))
		assert.Equal(t, 0, null.Value())
		assert.Equal(t, false, null.IsUndefined())
		assert.Equal(t, true, null.IsNull())
		assert.Equal(t, false, null.IsDefined())
		defined := FromOption(option.Some(option.Some(num)))
		assert.Equal(t, num, defined.Value())
		assert.Equal(t, false, defined.IsUndefined())
		assert.Equal(t, false, defined.IsNull())
		assert.Equal(t, true, defined.IsDefined())
	})
	t.Run("FromSqlNull", func(t *testing.T) {
		null := FromSqlNull(sql.Null[int]{Valid: false, V: 15})
		assert.Equal(t, 0, null.Value())
		assert.Equal(t, false, null.IsUndefined())
		assert.Equal(t, true, null.IsNull())
		assert.Equal(t, false, null.IsDefined())

		defined := FromSqlNull(sql.Null[int]{Valid: true, V: 15})
		assert.Equal(t, 15, defined.Value())
		assert.Equal(t, false, defined.IsUndefined())
		assert.Equal(t, false, defined.IsNull())
		assert.Equal(t, true, defined.IsDefined())
	})
}

// Tests for.
//
// - Equal
// - EqualFunc
// - Map
// - Unwrap()
func TestUnd_Methods(t *testing.T) {
	definedFoo := Defined("foo")
	definedFoo2 := Defined("foo")
	definedBar := Defined("bar")
	definedNull := Null[string]()
	undefined := Undefined[string]()

	t.Run("Equal", func(t *testing.T) {
		for _, combo := range [][2]Und[string]{
			{definedFoo, definedFoo2},
			{definedBar, definedBar},
			{definedNull, definedNull},
			{undefined, undefined},
		} {
			assert.Assert(t, combo[0].Equal(combo[1]))
		}

		for _, combo := range [][2]Und[string]{
			{definedBar, definedNull},
			{definedBar, undefined},
			{definedNull, undefined},
		} {
			assert.Assert(t, !combo[0].Equal(combo[1]))
		}
	})

	t.Run("EqualFunc", func(t *testing.T) {
		for _, combo := range [][2]Und[string]{
			{definedFoo, definedFoo2},
			{definedBar, definedBar},
			{definedNull, definedNull},
			{undefined, undefined},
		} {
			assert.Assert(t, combo[0].EqualFunc(combo[1], func(i, j string) bool { return i == j }))
		}

		for _, combo := range [][2]Und[string]{
			{definedFoo, definedFoo2},
			{definedBar, definedBar},
		} {
			assert.Assert(t, !combo[0].EqualFunc(combo[1], func(i, j string) bool { return i != j }))
		}

		for _, combo := range [][2]Und[string]{
			{definedBar, definedNull},
			{definedBar, undefined},
			{definedNull, undefined},
		} {
			assert.Assert(t, !combo[0].EqualFunc(combo[1], func(i, j string) bool { return true }))
		}
	})

	t.Run("Map", func(t *testing.T) {
		assert.Assert(
			t,
			definedBar.Map(func(o option.Option[option.Option[string]]) option.Option[option.Option[string]] {
				return Defined(o.Value().Value() + o.Value().Value()).Unwrap()
			}).Equal(Defined("barbar")),
		)
		assert.Assert(
			t,
			definedNull.Map(func(o option.Option[option.Option[string]]) option.Option[option.Option[string]] {
				return Defined("aa").Unwrap()
			}).Equal(Defined("aa")),
		)
		assert.Assert(
			t,
			undefined.Map(func(o option.Option[option.Option[string]]) option.Option[option.Option[string]] {
				return Defined("bb").Unwrap()
			}).Equal(Defined("bb")),
		)
	})

	t.Run("Unwrap", func(t *testing.T) {
		assert.Equal(t, definedBar.Unwrap(), option.Some(option.Some("bar")))
		assert.Equal(t, definedNull.Unwrap(), option.Some(option.None[string]()))
		assert.Equal(t, undefined.Unwrap(), option.None[option.Option[string]]())
	})
}
