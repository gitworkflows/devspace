// Copyright 2020 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package json

import (
	"encoding"
	"errors"
	"reflect"

	"github.com/go-json-experiment/json/internal"
	"github.com/go-json-experiment/json/internal/jsonflags"
	"github.com/go-json-experiment/json/internal/jsonopts"
	"github.com/go-json-experiment/json/internal/jsonwire"
	"github.com/go-json-experiment/json/jsontext"
)

var errNonStringValue = errors.New("JSON value must be string type")

// Interfaces for custom serialization.
var (
	jsonMarshalerV1Type   = reflect.TypeFor[MarshalerV1]()
	jsonMarshalerV2Type   = reflect.TypeFor[MarshalerV2]()
	jsonUnmarshalerV1Type = reflect.TypeFor[UnmarshalerV1]()
	jsonUnmarshalerV2Type = reflect.TypeFor[UnmarshalerV2]()
	textAppenderType      = reflect.TypeFor[encodingTextAppender]()
	textMarshalerType     = reflect.TypeFor[encoding.TextMarshaler]()
	textUnmarshalerType   = reflect.TypeFor[encoding.TextUnmarshaler]()

	// TODO(https://go.dev/issue/62384): Use encoding.TextAppender instead of this hack.
	// This exists for now to provide performance benefits to netip types.
	// There is no semantic difference with this change.
	appenderToType = reflect.TypeFor[interface{ AppendTo([]byte) []byte }]()

	allMarshalerTypes   = []reflect.Type{jsonMarshalerV2Type, jsonMarshalerV1Type, textAppenderType, textMarshalerType}
	allUnmarshalerTypes = []reflect.Type{jsonUnmarshalerV2Type, jsonUnmarshalerV1Type, textUnmarshalerType}
	allMethodTypes      = append(allMarshalerTypes, allUnmarshalerTypes...)
)

// TODO(https://go.dev/issue/62384): Use encoding.TextAppender instead
// and document public support for this method in json.Marshal.
type encodingTextAppender interface {
	AppendText(b []byte) ([]byte, error)
}

// MarshalerV1 is implemented by types that can marshal themselves.
// It is recommended that types implement [MarshalerV2] unless the implementation
// is trying to avoid a hard dependency on the "jsontext" package.
//
// It is recommended that implementations return a buffer that is safe
// for the caller to retain and potentially mutate.
type MarshalerV1 interface {
	MarshalJSON() ([]byte, error)
}

// MarshalerV2 is implemented by types that can marshal themselves.
// It is recommended that types implement MarshalerV2 instead of [MarshalerV1]
// since this is both more performant and flexible.
// If a type implements both MarshalerV1 and MarshalerV2,
// then MarshalerV2 takes precedence. In such a case, both implementations
// should aim to have equivalent behavior for the default marshal options.
//
// The implementation must write only one JSON value to the Encoder and
// must not retain the pointer to [jsontext.Encoder] or the [Options] value.
type MarshalerV2 interface {
	MarshalJSONV2(*jsontext.Encoder, Options) error

	// TODO: Should users call the MarshalEncode function or
	// should/can they call this method directly? Does it matter?
}

// UnmarshalerV1 is implemented by types that can unmarshal themselves.
// It is recommended that types implement [UnmarshalerV2] unless the implementation
// is trying to avoid a hard dependency on the "jsontext" package.
//
// The input can be assumed to be a valid encoding of a JSON value
// if called from unmarshal functionality in this package.
// UnmarshalJSON must copy the JSON data if it is retained after returning.
// It is recommended that UnmarshalJSON implement merge semantics when
// unmarshaling into a pre-populated value.
//
// Implementations must not retain or mutate the input []byte.
type UnmarshalerV1 interface {
	UnmarshalJSON([]byte) error
}

// UnmarshalerV2 is implemented by types that can unmarshal themselves.
// It is recommended that types implement UnmarshalerV2 instead of [UnmarshalerV1]
// since this is both more performant and flexible.
// If a type implements both UnmarshalerV1 and UnmarshalerV2,
// then UnmarshalerV2 takes precedence. In such a case, both implementations
// should aim to have equivalent behavior for the default unmarshal options.
//
// The implementation must read only one JSON value from the Decoder.
// It is recommended that UnmarshalJSONV2 implement merge semantics when
// unmarshaling into a pre-populated value.
//
// Implementations must not retain the pointer to [jsontext.Decoder] or
// the [Options] value.
type UnmarshalerV2 interface {
	UnmarshalJSONV2(*jsontext.Decoder, Options) error

	// TODO: Should users call the UnmarshalDecode function or
	// should/can they call this method directly? Does it matter?
}

func makeMethodArshaler(fncs *arshaler, t reflect.Type) *arshaler {
	// Avoid injecting method arshaler on the pointer or interface version
	// to avoid ever calling the method on a nil pointer or interface receiver.
	// Let it be injected on the value receiver (which is always addressable).
	if t.Kind() == reflect.Pointer || t.Kind() == reflect.Interface {
		return fncs
	}

	if needAddr, ok := implements(t, textMarshalerType); ok {
		fncs.nonDefault = true
		prevMarshal := fncs.marshal
		fncs.marshal = func(enc *jsontext.Encoder, va addressableValue, mo *jsonopts.Struct) error {
			if mo.Flags.Get(jsonflags.CallMethodsWithLegacySemantics) &&
				(needAddr && va.forcedAddr) {
				return prevMarshal(enc, va, mo)
			}
			marshaler := va.Addr().Interface().(encoding.TextMarshaler)
			if err := export.Encoder(enc).AppendRaw('"', false, func(b []byte) ([]byte, error) {
				b2, err := marshaler.MarshalText()
				return append(b, b2...), err
			}); err != nil {
				err = wrapSkipFunc(err, "marshal method")
				if mo.Flags.Get(jsonflags.ReportErrorsWithLegacySemantics) {
					return internal.NewMarshalerError(va.Addr().Interface(), err, "MarshalText") // unlike unmarshal, always wrapped
				}
				if !isSemanticError(err) && !export.IsIOError(err) {
					err = newMarshalErrorBefore(enc, t, err)
				}
				return err
			}
			return nil
		}
		// TODO(https://go.dev/issue/62384): Rely on encoding.TextAppender instead.
		if implementsAny(t, appenderToType) && t.PkgPath() == "net/netip" {
			fncs.marshal = func(enc *jsontext.Encoder, va addressableValue, mo *jsonopts.Struct) error {
				appender := va.Addr().Interface().(interface{ AppendTo([]byte) []byte })
				if err := export.Encoder(enc).AppendRaw('"', false, func(b []byte) ([]byte, error) {
					return appender.AppendTo(b), nil
				}); err != nil {
					if !isSemanticError(err) && !export.IsIOError(err) {
						err = newMarshalErrorBefore(enc, t, err)
					}
					return err
				}
				return nil
			}
		}
	}

	if needAddr, ok := implements(t, textAppenderType); ok {
		fncs.nonDefault = true
		prevMarshal := fncs.marshal
		fncs.marshal = func(enc *jsontext.Encoder, va addressableValue, mo *jsonopts.Struct) (err error) {
			if mo.Flags.Get(jsonflags.CallMethodsWithLegacySemantics) &&
				(needAddr && va.forcedAddr) {
				return prevMarshal(enc, va, mo)
			}
			appender := va.Addr().Interface().(encodingTextAppender)
			if err := export.Encoder(enc).AppendRaw('"', false, appender.AppendText); err != nil {
				err = wrapSkipFunc(err, "append method")
				if mo.Flags.Get(jsonflags.ReportErrorsWithLegacySemantics) {
					return internal.NewMarshalerError(va.Addr().Interface(), err, "AppendText") // unlike unmarshal, always wrapped
				}
				if !isSemanticError(err) && !export.IsIOError(err) {
					err = newMarshalErrorBefore(enc, t, err)
				}
				return err
			}
			return nil
		}
	}

	if needAddr, ok := implements(t, jsonMarshalerV1Type); ok {
		fncs.nonDefault = true
		prevMarshal := fncs.marshal
		fncs.marshal = func(enc *jsontext.Encoder, va addressableValue, mo *jsonopts.Struct) error {
			if mo.Flags.Get(jsonflags.CallMethodsWithLegacySemantics) &&
				((needAddr && va.forcedAddr) || export.Encoder(enc).Tokens.Last.NeedObjectName()) {
				return prevMarshal(enc, va, mo)
			}
			marshaler := va.Addr().Interface().(MarshalerV1)
			val, err := marshaler.MarshalJSON()
			if err != nil {
				err = wrapSkipFunc(err, "marshal method")
				if mo.Flags.Get(jsonflags.ReportErrorsWithLegacySemantics) {
					return internal.NewMarshalerError(va.Addr().Interface(), err, "MarshalJSON") // unlike unmarshal, always wrapped
				}
				err = newMarshalErrorBefore(enc, t, err)
				return collapseSemanticErrors(err)
			}
			if err := enc.WriteValue(val); err != nil {
				if mo.Flags.Get(jsonflags.ReportErrorsWithLegacySemantics) {
					return internal.NewMarshalerError(va.Addr().Interface(), err, "MarshalJSON") // unlike unmarshal, always wrapped
				}
				if isSyntacticError(err) {
					err = newMarshalErrorBefore(enc, t, err)
				}
				return err
			}
			return nil
		}
	}

	if needAddr, ok := implements(t, jsonMarshalerV2Type); ok {
		fncs.nonDefault = true
		prevMarshal := fncs.marshal
		fncs.marshal = func(enc *jsontext.Encoder, va addressableValue, mo *jsonopts.Struct) error {
			if mo.Flags.Get(jsonflags.CallMethodsWithLegacySemantics) &&
				((needAddr && va.forcedAddr) || export.Encoder(enc).Tokens.Last.NeedObjectName()) {
				return prevMarshal(enc, va, mo)
			}
			xe := export.Encoder(enc)
			prevDepth, prevLength := xe.Tokens.DepthLength()
			xe.Flags.Set(jsonflags.WithinArshalCall | 1)
			err := va.Addr().Interface().(MarshalerV2).MarshalJSONV2(enc, mo)
			xe.Flags.Set(jsonflags.WithinArshalCall | 0)
			currDepth, currLength := xe.Tokens.DepthLength()
			if (prevDepth != currDepth || prevLength+1 != currLength) && err == nil {
				err = errNonSingularValue
			}
			if err != nil {
				err = wrapSkipFunc(err, "marshal method")
				if mo.Flags.Get(jsonflags.ReportErrorsWithLegacySemantics) {
					return internal.NewMarshalerError(va.Addr().Interface(), err, "MarshalJSONV2") // unlike unmarshal, always wrapped
				}
				if !export.IsIOError(err) {
					err = newSemanticErrorWithPosition(enc, t, prevDepth, prevLength, err)
				}
				return err
			}
			return nil
		}
	}

	if _, ok := implements(t, textUnmarshalerType); ok {
		fncs.nonDefault = true
		fncs.unmarshal = func(dec *jsontext.Decoder, va addressableValue, uo *jsonopts.Struct) error {
			xd := export.Decoder(dec)
			var flags jsonwire.ValueFlags
			val, err := xd.ReadValue(&flags)
			if err != nil {
				return err // must be a syntactic or I/O error
			}
			if val.Kind() == 'n' {
				if !uo.Flags.Get(jsonflags.MergeWithLegacySemantics) {
					va.SetZero()
				}
				return nil
			}
			if val.Kind() != '"' {
				return newUnmarshalErrorAfter(dec, t, errNonStringValue)
			}
			s := jsonwire.UnquoteMayCopy(val, flags.IsVerbatim())
			unmarshaler := va.Addr().Interface().(encoding.TextUnmarshaler)
			if err := unmarshaler.UnmarshalText(s); err != nil {
				err = wrapSkipFunc(err, "unmarshal method")
				if uo.Flags.Get(jsonflags.ReportErrorsWithLegacySemantics) {
					return err // unlike marshal, never wrapped
				}
				if !isSemanticError(err) && !isSyntacticError(err) && !export.IsIOError(err) {
					err = newUnmarshalErrorAfter(dec, t, err)
				}
				return err
			}
			return nil
		}
	}

	if _, ok := implements(t, jsonUnmarshalerV1Type); ok {
		fncs.nonDefault = true
		prevUnmarshal := fncs.unmarshal
		fncs.unmarshal = func(dec *jsontext.Decoder, va addressableValue, uo *jsonopts.Struct) error {
			if uo.Flags.Get(jsonflags.CallMethodsWithLegacySemantics) &&
				export.Decoder(dec).Tokens.Last.NeedObjectName() {
				return prevUnmarshal(dec, va, uo)
			}
			val, err := dec.ReadValue()
			if err != nil {
				return err // must be a syntactic or I/O error
			}
			unmarshaler := va.Addr().Interface().(UnmarshalerV1)
			if err := unmarshaler.UnmarshalJSON(val); err != nil {
				err = wrapSkipFunc(err, "unmarshal method")
				if uo.Flags.Get(jsonflags.ReportErrorsWithLegacySemantics) {
					return err // unlike marshal, never wrapped
				}
				err = newUnmarshalErrorAfter(dec, t, err)
				return collapseSemanticErrors(err)
			}
			return nil
		}
	}

	if _, ok := implements(t, jsonUnmarshalerV2Type); ok {
		fncs.nonDefault = true
		prevUnmarshal := fncs.unmarshal
		fncs.unmarshal = func(dec *jsontext.Decoder, va addressableValue, uo *jsonopts.Struct) error {
			if uo.Flags.Get(jsonflags.CallMethodsWithLegacySemantics) &&
				export.Decoder(dec).Tokens.Last.NeedObjectName() {
				return prevUnmarshal(dec, va, uo)
			}
			xd := export.Decoder(dec)
			prevDepth, prevLength := xd.Tokens.DepthLength()
			xd.Flags.Set(jsonflags.WithinArshalCall | 1)
			err := va.Addr().Interface().(UnmarshalerV2).UnmarshalJSONV2(dec, uo)
			xd.Flags.Set(jsonflags.WithinArshalCall | 0)
			currDepth, currLength := xd.Tokens.DepthLength()
			if (prevDepth != currDepth || prevLength+1 != currLength) && err == nil {
				err = errNonSingularValue
			}
			if err != nil {
				err = wrapSkipFunc(err, "unmarshal method")
				if uo.Flags.Get(jsonflags.ReportErrorsWithLegacySemantics) {
					if err2 := xd.SkipUntil(prevDepth, prevLength+1); err2 != nil {
						return err2
					}
					return err // unlike marshal, never wrapped
				}
				if !isSyntacticError(err) && !export.IsIOError(err) {
					err = newSemanticErrorWithPosition(dec, t, prevDepth, prevLength, err)
				}
				return err
			}
			return nil
		}
	}

	return fncs
}

// implementsAny is like t.Implements(ifaceType) for a list of interfaces,
// but checks whether either t or reflect.PointerTo(t) implements the interface.
func implementsAny(t reflect.Type, ifaceTypes ...reflect.Type) bool {
	for _, ifaceType := range ifaceTypes {
		if _, ok := implements(t, ifaceType); ok {
			return true
		}
	}
	return false
}

// implements is like t.Implements(ifaceType) but checks whether
// either t or reflect.PointerTo(t) implements the interface.
// It also reports whether the value needs to be addressed
// in order to satisfy the interface.
func implements(t, ifaceType reflect.Type) (needAddr, ok bool) {
	switch {
	case t.Implements(ifaceType):
		return false, true
	case reflect.PointerTo(t).Implements(ifaceType):
		return true, true
	default:
		return false, false
	}
}
