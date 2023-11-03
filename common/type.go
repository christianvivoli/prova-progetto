package common

import (
	"bytes"
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"
)

// NullBytes is a JSON null literal
var NullBytes = []byte("null")

var _ json.Unmarshaler = (*Patch[any])(nil)
var _ json.Marshaler = (*Patch[any])(nil)

// JsonTime definisce un tipo per la gestione di null type per le operazioni di PATCH, dove è necessario sapere se NULL è un valore scelto dall'utente o per mancanza del campo passato.
type Patch[T any] struct {
	Value T
	Set   bool
}

// MarshalJSON implements json.Marshaler
func (v Patch[T]) MarshalJSON() ([]byte, error) {
	if !v.Set {
		return NullBytes, nil
	}
	return json.Marshal(v.Value)
}

// UnmarshalJSON implements json.Unmarshaler
func (v *Patch[T]) UnmarshalJSON(data []byte) error {
	// If this method was called, the value was set.
	v.Set = true

	if bytes.Equal(data, NullBytes) {
		// The key was set to null, leave zero value of T.
		return nil
	}

	var temp T
	if err := json.Unmarshal(data, &temp); err != nil {
		return err
	}
	v.Value = temp
	return nil
}

// NewPatch restituisce una nuova istanza di NewPatch.
func NewPatch[T any](val T) Patch[T] {
	return Patch[T]{
		Value: val,
		Set:   true,
	}
}

var _ sql.Scanner = (*Sex)(nil)
var _ driver.Valuer = (Sex)(Sex(0))
var _ json.Marshaler = (*Sex)(nil)
var _ json.Unmarshaler = (*Sex)(nil)

// Sex definisce il tipo per i valori del sesso di un cliente
type Sex byte

// UnmarshalJSON implements json.Unmarshaler
func (s *Sex) UnmarshalJSON(data []byte) error {

	if bytes.Equal(data, NullBytes) {
		return nil
	}

	switch string(data) {
	case "\"" + string(Male) + "\"":
		*s = Male
	case "\"" + string(Female) + "\"":
		*s = Female
	}

	return nil
}

// MarshalJSON implements json.Marshaler
func (s Sex) MarshalJSON() ([]byte, error) {
	type Alias Sex
	return json.Marshal(string((Alias)(s)))
}

// Costanti per definire i valori di Sex accettati
const (
	Male   Sex = 'm'
	Female Sex = 'f'
)

// Scan implements postgres.Scanner
func (s *Sex) Scan(src any) error {
	if v, ok := src.([]uint8); ok {
		if len(v) > 0 {
			*s = Sex(v[0])
		} else {
			*s = 0
		}
		return nil
	}
	return errors.New("error during scan type: Sex")
}

// Value implements driver.Valuer
func (s Sex) Value() (driver.Value, error) {
	return string(s), nil
}

var (
	// ErrNoneValueTaken represents the error that is raised when None value is taken.
	ErrNoneValueTaken = errors.New("none value taken")
)

// Option is a data type that must be Some (i.e. having a value) or None (i.e. doesn't have a value).
type Option[T any] []T

const (
	value = iota
)

// Some is a function to make an Option type instance with the actual value.
func Some[T any](v T) Option[T] {
	return Option[T]{
		value: v,
	}
}

// None is a function to make an Option type that doesn't have a value.
func None[T any]() Option[T] {
	return nil
}

// IsNone returns whether the Option *doesn't* have a value or not.
func (o Option[T]) IsNone() bool {
	return o == nil
}

// IsSome returns whether the Option has a value or not.
func (o Option[T]) IsSome() bool {
	return o != nil
}

// Unwrap returns the value regardless of Some/None status.
// If the Option value is Some, this method returns the actual value.
// On the other hand, if the Option value is None, this method returns the *default* value according to the type.
func (o Option[T]) Unwrap() T {
	if o.IsNone() {
		var defaultValue T
		return defaultValue
	}
	return o[value]
}

// Take takes the contained value in Option.
// If Option value is Some, this returns the value that is contained in Option.
// On the other hand, this returns an ErrNoneValueTaken as the second return value.
func (o Option[T]) Take() (T, error) {
	if o.IsNone() {
		var defaultValue T
		return defaultValue, ErrNoneValueTaken
	}
	return o[value], nil
}

// MarshalJSON implements json.Marshaler
func (o Option[T]) MarshalJSON() ([]byte, error) {
	if o.IsNone() {
		return NullBytes, nil
	}

	marshal, err := json.Marshal(o.Unwrap())
	if err != nil {
		return nil, err
	}
	return marshal, nil
}

// UnmarshalJSON implements json.Unmarshaler
func (o *Option[T]) UnmarshalJSON(data []byte) error {
	if len(data) <= 0 || bytes.Equal(data, NullBytes) {
		*o = None[T]()
		return nil
	}

	var v T
	err := json.Unmarshal(data, &v)
	if err != nil {
		return err
	}
	*o = Some(v)

	return nil
}

var _ sql.Scanner = (*Times)(nil)
var _ driver.Valuer = (Times)([]time.Time{})

// Times definisce un tipo slice di Time utile per lo storage su POSTGRES.
type Times []time.Time

// Scan implements postgres.Scanner
func (t *Times) Scan(src any) error {
	if v, ok := src.([]byte); ok {
		return json.Unmarshal(v, t)
	}
	return errors.New("error during scan type: Times")
}

// Value implements driver.Valuer
func (t Times) Value() (driver.Value, error) {
	return json.Marshal(t)
}

// Number defines a generic number type.
type Number interface {
	~int64 | ~int32 | ~int8 | ~int |
		~uint64 | ~uint32 | ~uint8 | ~uint |
		~float64 | ~float32
}
