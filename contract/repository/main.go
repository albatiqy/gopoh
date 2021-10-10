package repository

import (
	"fmt"
	"reflect"
)

type AttrFilter struct {
	Attr string
	Type string
	Val  string
}

type AttrOrder struct {
	Attr string
	Type string
}

type CursorData struct {
	NextToken string `json:"next_token"`
	PrevToken string `json:"prev_token"`
}

type FinderOptionCursor struct {
	Search             string
	PageSize           uint8
	CursorToken        string
	Orders             []AttrOrder
	Filters            []AttrFilter
	allowedOrderAttrs  []string
	allowedFilterAttrs []string
}

func (qOpt *FinderOptionCursor) AppendFilter(attr string, filterType string, val string) {
	qOpt.Filters = append(qOpt.Filters, AttrFilter{
		Attr: attr,
		Type: filterType,
		Val:  val,
	})
}

func (qOpt *FinderOptionCursor) AppendOrder(attr string, orderType string) {
	qOpt.Orders = append(qOpt.Orders, AttrOrder{
		Attr: attr,
		Type: orderType,
	})
}

func (qOpt *FinderOptionCursor) AllowedOrderAttrs(allowedOrderAttrs ...string) {
	qOpt.allowedOrderAttrs = allowedOrderAttrs
}

func (qOpt *FinderOptionCursor) AllowedFilterAttrs(allowedFilterAttrs ...string) {
	qOpt.allowedFilterAttrs = allowedFilterAttrs
}

type Input struct {
	values       map[string]reflect.Value
	structFields []reflect.StructField
}

func (sip Input) Struct() interface{} { // untuk validation
	structValue := reflect.New(reflect.StructOf(sip.structFields))
	structType := reflect.TypeOf(structValue)
	for i := 0; i < structType.NumField(); i++ {
		jsonTag := structType.Field(i).Tag.Get("json")
		structValue.FieldByName(structType.Field(i).Name).Set(sip.values[jsonTag])
	}
	return structValue.Interface()
}

func (sip Input) Map() map[string]interface{} {
	mapData := map[string]interface{}{}
	for attr, value := range sip.values {
		mapData[attr] = value.Interface()
	}
	return mapData
}

func NewInput(structInterface interface{}, validationRules map[string]string) (*Input, error) {
	structType := reflect.TypeOf(structInterface)
	if structType.Kind() != reflect.Struct {
		return nil, fmt.Errorf("bukan struct")
	}
	structValue := reflect.ValueOf(structInterface)

	var (
		values       = map[string]reflect.Value{}
		structFields []reflect.StructField
	)
	if validationRules != nil {
		var validationRule string
		for i := 0; i < structType.NumField(); i++ {
			jsonTag := structType.Field(i).Tag.Get("json")
			if jsonTag == "" {
				return nil, fmt.Errorf("missing json tag")
			}
			if !structValue.Field(i).IsZero() {
				if rule, ok := validationRules[jsonTag]; ok {
					validationRule = rule
				} else {
					validationRule = ""
				}
				sTag := `json:"` + jsonTag + `"` // `json:"` + val + `" form:"` + val + `" xml:"` + val + `"`
				if validationRule != "" {
					sTag += ` validate:"` + validationRule + `"`
				}
				structField := structType.Field(i)
				structField.Tag = reflect.StructTag(sTag)
				structFields = append(structFields, structField)
				values[jsonTag] = structValue.FieldByName(structField.Name)
			}
		}
	} else {
		for i := 0; i < structType.NumField(); i++ {
			jsonTag := structType.Field(i).Tag.Get("json")
			if jsonTag == "" {
				return nil, fmt.Errorf("missing json tag")
			}
			if !structValue.Field(i).IsZero() {
				sTag := `json:"` + jsonTag + `"` // `json:"` + val + `" form:"` + val + `" xml:"` + val + `"`
				structField := structType.Field(i)
				structField.Tag = reflect.StructTag(sTag)
				structFields = append(structFields, structField)
				values[jsonTag] = structValue.FieldByName(structField.Name)
			}
		}
	}

	return &Input{
		values:       values,
		structFields: structFields,
	}, nil
}
