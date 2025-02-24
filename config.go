package app

import (
	"encoding/json"
	"io"
	"os"
	"reflect"
	"strconv"
)

var config []byte

func readConfigFile(fileName string) error {
	file, err := os.Open(fileName)
	if err != nil {
		return err
	}
	defer file.Close()
	config, err = io.ReadAll(file)
	if err != nil {
		return err
	}
	return nil
}

func GetConfig(v any) error {
	if err := json.Unmarshal(config, v); err != nil {
		return err
	}
	setDefaults(reflect.ValueOf(v).Elem())
	return nil
}

func setDefaults(v reflect.Value) {
	for i := 0; i < v.NumField(); i++ {
		field := v.Type().Field(i)

		if field.Type.Kind() == reflect.Struct {
			setDefaults(v.FieldByName(field.Name))
			continue
		}

		if field.Type.Kind() == reflect.Slice {
			for j := 0; j < v.FieldByName(field.Name).Len(); j++ {
				if v.FieldByName(field.Name).Index(j).Kind() == reflect.Struct {
					setDefaults(v.FieldByName(field.Name).Index(j))
				}
			}
			continue
		}

		defValue := field.Tag.Get("default")
		if defValue == "" {
			continue
		}

		elem := v.FieldByName(field.Name)
		if !elem.IsValid() || !elem.CanSet() {
			continue
		}

		switch elem.Kind() {
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			if elem.Uint() == 0 {
				val, _ := strconv.ParseUint(defValue, 10, 64)
				elem.SetUint(val)
			}

		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			if elem.Int() == 0 {
				val, _ := strconv.ParseInt(defValue, 10, 64)
				elem.SetInt(val)
			}

		case reflect.Float32, reflect.Float64:
			if elem.Float() == 0 {
				val, _ := strconv.ParseFloat(defValue, 64)
				elem.SetFloat(val)
			}

		case reflect.String:
			if elem.String() == "" {
				elem.SetString(defValue)
			}

		default:
			continue
		}
	}
}
