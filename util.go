package pulse

import (
    "fmt"
    "reflect"
    "strings"
    // "unsafe"

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
        var specParts []string

        field := targetStruct.Type().Field(i)
        value := targetStruct.Field(i)

    //  get the tag from this field and parse the spec...
        if keyTagSpec := field.Tag.Get(`key`); keyTagSpec != `` {
            specParts = strings.Split(keyTagSpec, `,`)
        }else{
    //  ...or just assume a 1-to-1 map key to struct field name mapping
            specParts = []string{ field.Name }
        }

    //  a tag value '-' skips all processing of this struct field
        if specParts[0] == `-` {
            continue
        }

    //  attempt to retrieve the key specified in the field tag from the incoming map
        if dataValue, ok := data[specParts[0]]; ok {
            skipField := false
            dv := reflect.ValueOf(dataValue)

        //  handle tag flags
            for _, tagFlag := range specParts[1:] {
                switch tagFlag {
            //  omitempty: test if the incoming data value is a zero value, and if so, skip it
                case `omitempty`:
                    if !reflect.ValueOf(dataValue).IsValid() {
                        skipField = true
                    }
                }
            }

        //  if we're skipping this field for any reason, continue now
            if skipField {
                continue
            }

        //  verify that the destination struct field can be modified
            if value.CanSet() {
            //  if the source type cannot be directly assigned to the destination field,
            //  see if we can convert it (and do so)
                if !value.Type().AssignableTo(dv.Type()) {
                    if dv.Type().ConvertibleTo(value.Type()) {
                        dv = dv.Convert(value.Type())
                    }else{
                        return fmt.Errorf("Cannot convert '%v' (type %T) to field %s type %s", dataValue, dataValue, field.Name, field.Type.String())
                    }
                }

            //  double check that we can directly assign the data now.  if not, error out
                if value.Type().AssignableTo(dv.Type()) {
                    value.Set(dv)
                }else{
                    return fmt.Errorf("Cannot assign '%v' (type %T) to field %s (type %s)", dataValue, dataValue, field.Name, field.Type.String())
                }
            }
        }
    }

    return nil
}