package main

import (
    "reflect"
    "fmt"
)

func i2s(data interface{}, out interface{}) error {
    if reflect.ValueOf(out).Type().Kind() != reflect.Ptr {
        return fmt.Errorf("bad out format type %#v, should be a pointer",  out)
    }

    dataRef := reflect.ValueOf(data)
    outRef := reflect.ValueOf(out).Elem()
    
    switch outRef.Type().Kind() {
    case reflect.Int:
        d, ok  := data.(float64)
        if !ok {
            return fmt.Errorf("bad type conversion to Int %#v", data)
        }
        outRef.Set(reflect.ValueOf(int(d)))
    case reflect.String:
        d, ok  := data.(string)
        if !ok {
            return fmt.Errorf("bad type conversion to String %#v", data)
        }
        outRef.SetString(string(d))
    case reflect.Bool:
        d, ok  := data.(bool)
        if !ok {
            return fmt.Errorf("bad type conversion to Bool %#v", data)
        }
        outRef.SetBool(bool(d))
    case reflect.Struct:
        for i := 0; i < outRef.NumField(); i++ {
    		valueField := outRef.Field(i)
    		typeField := outRef.Type().Field(i)
            
            var value reflect.Value
            switch dataRef.Type().Kind() {
            case reflect.Map:
                value = dataRef.MapIndex(reflect.ValueOf(typeField.Name))
            default:
                return fmt.Errorf("bad type: %v for field %v", typeField.Type.Kind(), typeField.Name)                
            }
            
            err := i2s(value.Interface(), valueField.Addr().Interface())
            if err != nil {
                // fmt.Println("Error: ", err)
                return err
            }
    	}
    case reflect.Slice:
        _, ok := dataRef.Interface().([]interface{})
        if !ok {
            return fmt.Errorf("Bad type conversion to Slice %#v",  dataRef.Interface())
        }
        
        length := len(dataRef.Interface().([]interface{}))
        outRef.Set(reflect.MakeSlice(outRef.Type(), length, length))
        
        for key, val := range dataRef.Interface().([]interface{}) {
            switch reflect.TypeOf(val).Kind() {
            case reflect.Map:
                err := i2s(val, outRef.Index(key).Addr().Interface())
                if err != nil {
                    // fmt.Println("Error: ", err)
                    return err
                }
                
            default:
                return fmt.Errorf("bad type: %v for field %v", outRef.Type().Kind(), outRef.Type().Name)
            }
            
        }
        
    default:
        return fmt.Errorf("bad type: %v for field %v", outRef.Type().Kind(), outRef.Type().Name)
    }
    
    return nil
}
