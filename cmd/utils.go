package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path"
	"reflect"
	"sync"
	"unicode"
)

type ConcSchema[T any] struct {
	sync.RWMutex
	Schema map[string]T
}
type ConcSlice[T any] struct {
	sync.RWMutex
	Slice []T
}

func unpackJSON(bytes *[]byte) ([]json.RawMessage, error) {
	var allRaw []json.RawMessage
	err := json.Unmarshal(*bytes, &allRaw)
	if err != nil {
		return nil, err
	}
	return allRaw, nil
}

func copyMap[I comparable, T any](src map[I]T) map[I]T {
	copy := make(map[I]T, len(src))
	for key, val := range src {
		copy[key] = val
	}
	return copy
}

func parseJSON(bytes *[]byte, schema map[string]any) ([]map[string]any, error) {
	res := make([]map[string]any, 0)
	raw, err := unpackJSON(bytes)
	if err != nil {
		err = json.Unmarshal(*bytes, &schema)
		res = append(res, copyMap(schema))
	} else {
		for _, obj := range raw {
			json.Unmarshal(obj, &schema)
			res = append(res, copyMap(schema))
		}
	}
	return res, nil
}

func collectFields(file *os.File, schema *ConcSchema[any], sidekick *ConcSchema[any]) {

	text, err := io.ReadAll(file)
	if err != nil {
		fmt.Printf("Couldn't read file %s\n", file.Name())
		return
	}

	tempSchema := make(map[string]any, 0)
	raw, err := unpackJSON(&text)
	if err != nil {
		err = json.Unmarshal(text, &tempSchema)
		if err != nil {
			fmt.Printf("Couldnt parse %s in %s", raw, file.Name())
			return
		}
	} else {
		for _, obj := range raw {

			err = json.Unmarshal(obj, &tempSchema)
			if err != nil {
				fmt.Printf("Couldnt parse %s in %s", obj, file.Name())
				continue
			}
			writeFields(schema.Schema, tempSchema, sidekick)
		}
	}
}

func writeFields(targ map[string]any, src map[string]any, sidekicks *ConcSchema[any]) {
	for key, value := range src {
		valueType := reflect.TypeOf(value)
		switch valueType.Kind() {

		case reflect.Map:
			pascalKey := snakeToPascal(key)
			mapValue, ok := value.(map[string]any)
			if !ok {
				targ[key] = "unknown"
			}
			sidekickValue, ok := sidekicks.Schema[pascalKey]
			if !ok {
				sidekickValue = make(map[string]any)
				sidekicks.Schema[pascalKey] = sidekickValue
			}
			sidekickValueMap, ok := sidekickValue.(map[string]any)
			if !ok {
				continue
			}
			tempValue := make(map[string]any)
			targ[key] = pascalKey
			writeFields(tempValue, mapValue, sidekicks)
			for key, value := range tempValue {
				sidekickValueMap[key] = value
			}

		case reflect.Slice:
			_key := snakeToPascal(key)
			_value, ok := value.([]any)
			if !ok {
				targ[key] = "unknown"
			}
			targ[key] = getSliceType(&_value, _key, sidekicks)

		case reflect.Float64:
			targ[key] = "float64"

		case reflect.Bool:
			targ[key] = "bool"

		case reflect.String:
			targ[key] = "string"

		default:
			println(valueType.Kind().String())
		}

	}
}

func getSliceType(slice *[]any, keyBase string, sidekicks *ConcSchema[any]) string {
	if len(*slice) == 0 {
		return "[]unknown"
	}
	types := make([]string, 0)
	for idx, val := range *slice {
		_type := reflect.TypeOf(val).Kind().String()
		if reflect.TypeOf(val).Kind() == reflect.Map {
			key := keyBase + "[" + fmt.Sprint(idx) + "]"
			schema, ok := sidekicks.Schema[key]
			if !ok || reflect.TypeOf(schema).Kind() != reflect.Map {
				schema = make(map[string]any)
				sidekicks.Schema[key] = schema
			}
			writeFields(schema.(map[string]any), val.(map[string]any), sidekicks)
			_type = fmt.Sprint(schema)
		}
		if reflect.TypeOf(val).Kind() == reflect.Slice {
			valSlice, ok := val.([]any)
			if !ok {
				println("Fuckery")
				continue
			}
			_type = getSliceType(&valSlice, keyBase, sidekicks)
		}
		types = append(types, _type)
	}
	same := true
	cur := types[0]
	for _, _type := range types {
		if _type != cur {
			same = false
			break
		}
	}
	if same {
		return "[]" + cur
	}
	return fmt.Sprint(types)
}

func collectFieldVariants(file *os.File, field string, fieldTypes *map[string]int) error {
	bytes, err := io.ReadAll(file)
	if err != nil {
		return err
	}

	schema := make(map[string]any, 0)
	schema[field] = nil

	objs, err := parseJSON(&bytes, schema)
	if err != nil {
		return err
	}
	for _, obj := range objs {
		if obj[field] == nil {
			continue
		}
		fieldType := reflect.TypeOf((obj)[field]).String()
		(*fieldTypes)[fieldType]++
	}
	return nil
}

func snakeToPascal(text string) string {
	res := ""
	for idx := 0; idx < len(text); idx++ {
		if idx == 0 {
			res += string(unicode.ToUpper(rune(text[idx])))
		} else if text[idx] == '_' {
			if idx+1 >= len(text) {
			}
			res += string(unicode.ToUpper(rune(text[idx+1])))
			idx++
		} else {
			res += string(text[idx])
		}
	}
	return res
}

func producePaths(_path string) ([]string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	targPath := ""
	if _path[0] != '/' {
		targPath = path.Join(cwd, _path)
	}
	stat, err := os.Stat(targPath)
	if err != nil {
		return nil, err
	}
	if !stat.IsDir() {
		return []string{targPath}, nil
	}
	dir, err := os.ReadDir(targPath)
	if err != nil {
		return nil, err
	}
	paths := make([]string, 0)
	for _, file := range dir {
		paths = append(paths, path.Join(targPath, file.Name()))
	}
	return paths, nil
}
