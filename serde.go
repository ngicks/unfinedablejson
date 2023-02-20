package undefinedablejson

import (
	"reflect"
	"strings"
	"unsafe"

	jsoniter "github.com/json-iterator/go"
	"github.com/modern-go/reflect2"
)

var config = jsoniter.Config{
	EscapeHTML:             true,
	SortMapKeys:            true,
	ValidateJsonRawMessage: true,
}.Froze()

func init() {
	config.RegisterExtension(&UndefinedableExtension{})
}

type IsUndefineder interface {
	IsUndefined() bool
}

var undefinedableTy = reflect2.TypeOfPtr((*IsUndefineder)(nil)).Elem()

// UndefinedableEncoder fakes the Encoder so that
// undefined Undefinedable[T] fields are skipped.
type UndefinedableEncoder struct {
	ty  reflect2.Type
	org jsoniter.ValEncoder
}

func (e UndefinedableEncoder) IsEmpty(ptr unsafe.Pointer) bool {
	val := e.ty.UnsafeIndirect(ptr)
	return val.(IsUndefineder).IsUndefined()
}

func (e UndefinedableEncoder) Encode(ptr unsafe.Pointer, stream *jsoniter.Stream) {
	e.org.Encode(ptr, stream)
}

// FakedOmitemptyField implements reflect2.StructField interface,
// faking the struct tag to pretend it is always tagged with ,omitempty option.
type FakedOmitemptyField struct {
	reflect2.StructField
	fakedTag reflect.StructTag
}

func NewFakedOmitemptyField(f reflect2.StructField) FakedOmitemptyField {
	return FakedOmitemptyField{
		StructField: f,
		fakedTag:    FakeOmitempty(f.Tag()),
	}
}

func (f FakedOmitemptyField) Tag() reflect.StructTag {
	return f.fakedTag
}

func FakeOmitempty(t reflect.StructTag) reflect.StructTag {
	tags, err := ParseStructTag(t)
	if err != nil {
		panic(err)
	}

	found := false
	for i, tag := range tags {
		if found {
			break
		}
		if tag.Key != "json" {
			continue
		}

		found = true

		options := strings.Split(tag.Value, ",")
		if len(options) > 0 {
			// skip a first element since it is the field name.
			options = options[1:]
		}

		hasOmitempty := false
		for _, opt := range options {
			if opt == "omitempty" {
				hasOmitempty = true
			}
		}

		if !hasOmitempty {
			tags[i].Value += ",omitempty"
		}
	}

	if !found {
		tags = append(tags, Tag{Key: "json", Value: ",omitempty"})
	}

	return FlattenStructTag(tags)
}

// UndefinedableExtension is the extension for jsoniter.API.
// This forces jsoniter.API to skip undefined Undefinedable[T] when marshalling.
type UndefinedableExtension struct {
}

func (extension *UndefinedableExtension) UpdateStructDescriptor(structDescriptor *jsoniter.StructDescriptor) {
	if structDescriptor.Type.Implements(undefinedableTy) {
		return
	}

	for _, binding := range structDescriptor.Fields {
		if binding.Field.Type().Implements(undefinedableTy) {
			enc := binding.Encoder
			binding.Field = NewFakedOmitemptyField(binding.Field)
			binding.Encoder = UndefinedableEncoder{ty: binding.Field.Type(), org: enc}
		}
	}
}

func (extension *UndefinedableExtension) CreateMapKeyDecoder(typ reflect2.Type) jsoniter.ValDecoder {
	return nil
}

func (extension *UndefinedableExtension) CreateMapKeyEncoder(typ reflect2.Type) jsoniter.ValEncoder {
	return nil
}

func (extension *UndefinedableExtension) CreateDecoder(typ reflect2.Type) jsoniter.ValDecoder {
	return nil
}

func (extension *UndefinedableExtension) CreateEncoder(typ reflect2.Type) jsoniter.ValEncoder {
	return nil
}

func (extension *UndefinedableExtension) DecorateDecoder(typ reflect2.Type, decoder jsoniter.ValDecoder) jsoniter.ValDecoder {
	return decoder
}

func (extension *UndefinedableExtension) DecorateEncoder(typ reflect2.Type, encoder jsoniter.ValEncoder) jsoniter.ValEncoder {
	return encoder
}

// MarshalFieldsJSON encodes v into JSON.
// It skips fields if those are undefined Undefinedable[T].
//
// v can be any type.
func MarshalFieldsJSON(v any) ([]byte, error) {
	return config.Marshal(v)
}

// UnmarshalFieldsJSON decodes data into v.
// v must be pointer type, return error otherwise.
//
// Currently this is almost same as json.Unmarshal.
// Future releases may change behavior of this function.
// It is safe to unmarshal data through this if v has at least an Undefinedable[T] field.
func UnmarshalFieldsJSON(data []byte, v any) error {
	return config.Unmarshal(data, v)
}
