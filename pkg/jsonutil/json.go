package jsonutil

import (
	"encoding/json"
	"reflect"
	"strings"
	"time"
	"unicode"
)

var timeType = reflect.TypeOf(time.Time{})

func Marshal(value any) ([]byte, error) {
	prepared, err := encode(reflect.ValueOf(value))
	if err != nil {
		return nil, err
	}
	return json.Marshal(prepared)
}

func MarshalIndent(value any, prefix, indent string) ([]byte, error) {
	prepared, err := encode(reflect.ValueOf(value))
	if err != nil {
		return nil, err
	}
	return json.MarshalIndent(prepared, prefix, indent)
}

func Unmarshal(data []byte, target any) error {
	var raw any
	decoder := json.NewDecoder(strings.NewReader(string(data)))
	decoder.UseNumber()
	if err := decoder.Decode(&raw); err != nil {
		return err
	}
	return assign(reflect.ValueOf(target).Elem(), raw)
}

func encode(value reflect.Value) (any, error) {
	if !value.IsValid() {
		return nil, nil
	}
	if value.Kind() == reflect.Interface || value.Kind() == reflect.Pointer {
		if value.IsNil() {
			return nil, nil
		}
		return encode(value.Elem())
	}
	if value.Type() == timeType {
		return value.Interface(), nil
	}
	switch value.Kind() {
	case reflect.Struct:
		result := map[string]any{}
		err := eachExportedField(value, func(index int, field reflect.StructField) error {
			name, omit := fieldName(field)
			if name == "-" || omit && value.Field(index).IsZero() {
				return nil
			}
			item, err := encode(value.Field(index))
			if err != nil {
				return err
			}
			result[name] = item
			return nil
		})
		if err != nil {
			return nil, err
		}
		return result, nil
	case reflect.Map:
		result := map[string]any{}
		err := eachMapKey(value, func(key reflect.Value) error {
			item, err := encode(value.MapIndex(key))
			if err != nil {
				return err
			}
			result[key.String()] = item
			return nil
		})
		if err != nil {
			return nil, err
		}
		return result, nil
	case reflect.Slice, reflect.Array:
		result := make([]any, value.Len())
		err := eachIndex(value.Len(), func(index int) error {
			item, err := encode(value.Index(index))
			if err != nil {
				return err
			}
			result[index] = item
			return nil
		})
		if err != nil {
			return nil, err
		}
		return result, nil
	default:
		return value.Interface(), nil
	}
}

func assign(target reflect.Value, raw any) error {
	if target.Kind() == reflect.Pointer {
		if raw == nil {
			return nil
		}
		target.Set(reflect.New(target.Type().Elem()))
		return assign(target.Elem(), raw)
	}
	if target.Type() == timeType {
		encoded, err := json.Marshal(raw)
		if err != nil {
			return err
		}
		return json.Unmarshal(encoded, target.Addr().Interface())
	}
	switch target.Kind() {
	case reflect.Struct:
		values, ok := raw.(map[string]any)
		if !ok {
			return nil
		}
		return eachExportedField(target, func(index int, field reflect.StructField) error {
			name, _ := fieldName(field)
			if value, exists := values[name]; exists {
				return assign(target.Field(index), value)
			}
			return nil
		})
	case reflect.Map:
		values, ok := raw.(map[string]any)
		if !ok {
			return nil
		}
		target.Set(reflect.MakeMapWithSize(target.Type(), len(values)))
		return eachStringMap(values, func(name string, value any) error {
			item := reflect.New(target.Type().Elem()).Elem()
			if err := assign(item, value); err != nil {
				return err
			}
			target.SetMapIndex(reflect.ValueOf(name).Convert(target.Type().Key()), item)
			return nil
		})
	case reflect.Slice:
		values, ok := raw.([]any)
		if !ok {
			return nil
		}
		target.Set(reflect.MakeSlice(target.Type(), len(values), len(values)))
		return eachValues(values, func(index int, value any) error {
			return assign(target.Index(index), value)
		})
	default:
		encoded, err := json.Marshal(raw)
		if err != nil {
			return err
		}
		return json.Unmarshal(encoded, target.Addr().Interface())
	}
}

func eachIndex(length int, visit func(int) error) error {
	for index := 0; index < length; index++ {
		if err := visit(index); err != nil {
			return err
		}
	}
	return nil
}

func eachValues(values []any, visit func(int, any) error) error {
	return eachIndex(len(values), func(index int) error {
		return visit(index, values[index])
	})
}

func eachExportedField(value reflect.Value, visit func(int, reflect.StructField) error) error {
	return eachIndex(value.NumField(), func(index int) error {
		field := value.Type().Field(index)
		if field.PkgPath != "" {
			return nil
		}
		return visit(index, field)
	})
}

func eachMapKey(value reflect.Value, visit func(reflect.Value) error) error {
	keys := value.MapKeys()
	return eachIndex(len(keys), func(index int) error {
		return visit(keys[index])
	})
}

func eachStringMap(values map[string]any, visit func(string, any) error) error {
	for name := range values {
		value := values[name]
		if err := visit(name, value); err != nil {
			return err
		}
	}
	return nil
}

func fieldName(field reflect.StructField) (string, bool) {
	tag := field.Tag.Get("json")
	if tag != "" {
		parts := strings.Split(tag, ",")
		return parts[0], strings.Contains(tag, ",omitempty")
	}
	return snakeCase(field.Name), false
}

func snakeCase(value string) string {
	runes := []rune(value)
	var builder strings.Builder
	for index := range runes {
		current := runes[index]
		if unicode.IsUpper(current) && index > 0 {
			previous := runes[index-1]
			var next rune
			if index+1 < len(runes) {
				next = runes[index+1]
			}
			if unicode.IsLower(previous) || unicode.IsDigit(previous) || unicode.IsLower(next) {
				builder.WriteByte('_')
			}
		}
		builder.WriteRune(unicode.ToLower(current))
	}
	return builder.String()
}
