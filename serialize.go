package gody

import (
	"fmt"
	"reflect"
	"strings"
)

// ErrInvalidBody represents all invalid body report
type ErrInvalidBody struct {
	Kind reflect.Kind
}

func (e *ErrInvalidBody) Error() string {
	return fmt.Sprintln("invalid body:", e.Kind)
}

// ErrInvalidTag represents all invalid tag report
type ErrInvalidTag struct {
	Format string
}

func (e *ErrInvalidTag) Error() string {
	return fmt.Sprintln("invalid tag:", e.Format)
}

// Field is a struct to represents the domain about a field inner gody lib
type Field struct {
	Name  string
	Value string
	Tags  map[string]string
}

// Serialize is a func to serialize/parse all content about the struct input
func Serialize(b interface{}) ([]Field, error) {
	if b == nil {
		return nil, &ErrInvalidBody{}
	}

	valueOf := reflect.ValueOf(b)
	typeOf := reflect.TypeOf(b)

	if kindOfBody := typeOf.Kind(); kindOfBody != reflect.Struct {
		return nil, &ErrInvalidBody{Kind: kindOfBody}
	}

	fields := make([]Field, 0)
	for i := 0; i < typeOf.NumField(); i++ {
		field := typeOf.Field(i)
		tagString := field.Tag.Get("validate")
		if tagString == "" {
			continue
		}

		tagFormats := strings.Fields(tagString)
		tags := make(map[string]string)
		for i := 0; i < len(tagFormats); i++ {
			tagFormatSplitted := strings.Split(tagFormats[i], "=")
			if len(tagFormatSplitted) != 2 {
				return nil, &ErrInvalidTag{Format: tagFormats[i]}
			}
			tags[tagFormatSplitted[0]] = tagFormatSplitted[1]
		}

		fieldValue := valueOf.FieldByName(field.Name)
		fieldNameToLower := strings.ToLower(field.Name)
		if kindOfField := field.Type.Kind(); kindOfField == reflect.Struct {
			payload := fieldValue.Convert(field.Type).Interface()
			serialized, err := Serialize(payload)
			if err != nil {
				return nil, err
			}
			for _, item := range serialized {
				fields = append(fields, Field{
					Name:  fmt.Sprintf("%s.%s", fieldNameToLower, item.Name),
					Value: item.Value,
					Tags:  item.Tags,
				})
			}
		} else if kindOfField := field.Type.Kind(); kindOfField == reflect.Slice {
			j := fieldValue.Len()
			for i := 0; i < j; i++ {
				sliceFieldValue := fieldValue.Index(i)
				payload := sliceFieldValue.Convert(sliceFieldValue.Type()).Interface()
				serialized, err := Serialize(payload)
				if err != nil {
					return nil, err
				}
				for _, item := range serialized {
					fields = append(fields, Field{
						Name:  fmt.Sprintf("%s[%v].%s", fieldNameToLower, i, item.Name),
						Value: item.Value,
						Tags:  item.Tags,
					})
				}
			}
		} else {
			fieldValueString := fmt.Sprintf("%v", fieldValue)
			fields = append(fields, Field{
				Name:  fieldNameToLower,
				Value: fieldValueString,
				Tags:  tags,
			})
		}
	}

	return fields, nil
}
