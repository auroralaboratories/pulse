package pulse

import (
    "fmt"
    "reflect"
    "strings"
    "unsafe"

    // log "github.com/Sirupsen/logrus"
)

func UnmarshalMap(data map[string]interface{}, target interface{}) error {
    var targetStruct reflect.Value

    if kindOf := reflect.TypeOf(target).Kind(); kindOf == reflect.Ptr {
        if pointedTo := reflect.Indirect(reflect.ValueOf(target)); pointedTo.IsValid() {
            if kindOf := reflect.TypeOf(pointedTo).Kind(); kindOf == reflect.Struct {
                targetStruct = pointedTo
            }else{
                return fmt.Errorf("Cannot unmarshal map into type %T", target)
            }
        }else{
            return fmt.Errorf("Cannot unmarshal map into non-existent target")
        }
    }else{
        return fmt.Errorf("Unmarshal map only accepts a pointer to a struct")
    }

    for i := 0; i < targetStruct.NumField(); i++ {
        field := targetStruct.Type().Field(i)
        value := targetStruct.Field(i)

        if keyTagSpec := field.Tag.Get(`key`); keyTagSpec != `` {
            specParts := strings.Split(keyTagSpec, `,`)

            if specParts[0] == `-` {
                continue
            }

            if dataValue, ok := data[specParts[0]]; ok {
                skipField := false
                dv := reflect.ValueOf(dataValue)

                for _, tagFlag := range specParts[1:] {
                    switch tagFlag {
                    case `omitempty`:
                        if !reflect.ValueOf(dataValue).IsValid() {
                            skipField = true
                        }
                    }
                }

                if skipField {
                    continue
                }

                if value.CanSet() {
                    if !value.Type().AssignableTo(dv.Type()) {
                        if dv.Type().ConvertibleTo(value.Type()) {
                            dv = dv.Convert(value.Type())
                        }
                    }

                    if value.Type().AssignableTo(dv.Type()) {
                        switch dv.Kind() {
                        case reflect.Bool:
                            value.SetBool(dv.Bool())
                        case reflect.Complex64, reflect.Complex128:
                            value.SetComplex(dv.Complex())
                        case reflect.Float32, reflect.Float64:
                            value.SetFloat(dv.Float())
                        case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
                            value.SetInt(dv.Int())
                        // case reflect.Map:
                        case reflect.String:
                            value.SetString(dv.String())
                        case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
                            value.SetUint(dv.Uint())
                        case reflect.Uintptr, reflect.Ptr, reflect.UnsafePointer:
                            value.SetPointer(unsafe.Pointer(dv.UnsafeAddr()))
                        }
                    }else{
                        return fmt.Errorf("Cannot assign '%v' (type %T) to field %s (type %s)", dataValue, dataValue, field.Name, field.Type.String())
                    }
                // }else{
                //     log.Warnf("Field %s cannot be changed", field.Name)
                }
            }
        }
    }

    return nil
}