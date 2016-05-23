package csvutil

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"os"
	"reflect"
	"strconv"
)

/*
csv format
column1		column2 	column3 	column4 	...
int			string		int			int			comment...
data11		data12		data13		data14		data...
data21		data22		data23		data24		data...
*/

func ReadFile(fname string, out interface{}) error {
	v := reflect.ValueOf(out).Elem()
	return readFile(fname, v)
}

func readFile(fname string, v reflect.Value) error {
	if v.Kind() != reflect.Slice {
		return errors.New("Only accept slice type for parse csv file.")
	}
	item := reflect.New(v.Type().Elem()).Elem()

	fmt.Println(item.Type().String())

	f, err := os.Open(fname)

	if err != nil {
		return err
	}

	defer f.Close()

	rd := csv.NewReader(f)
	columns, err := rd.Read()
	if err != nil {
		return err
	}

	colmap := make(map[string]int)

	for i := 0; i < item.NumField(); i++ {
		fieldInfo := item.Type().Field(i)
		tag := fieldInfo.Tag
		name := tag.Get("csv")
		if name == "" {
			name = fieldInfo.Name
		}

		found := false
		for j := 0; j < len(columns); j++ {
			if name == columns[j] {
				colmap[fieldInfo.Name] = j
				found = true
				break
			}
		}

		if !found {
			return errors.New(fmt.Sprintf("cannot found column %s\n", name))
		}
	}

	fmt.Println(colmap)
	//skip comment
	rd.Read()

	for {
		rec, err := rd.Read()

		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}

		data := reflect.New(v.Type().Elem()).Elem()

		for i := 0; i < data.NumField(); i++ {
			inf := data.Type().Field(i)
			idx := colmap[inf.Name]
			populate(data.Field(i), rec[idx])
		}
		fmt.Println(data)
		v.Set(reflect.Append(v, data))
	}
	return nil
}

func populate(v reflect.Value, value string) error {
	switch v.Kind() {
	case reflect.Bool:
		b, err := strconv.ParseBool(value)
		if err != nil {
			return err
		}
		v.SetBool(b)
	case reflect.String:
		v.SetString(value)
	case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int, reflect.Int64:
		i, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return err
		}
		v.SetInt(i)
	case reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint, reflect.Uint64:
		u, err := strconv.ParseUint(value, 10, 64)
		if err != nil {
			return err
		}
		v.SetUint(u)
	case reflect.Float32:
		f, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return err
		}
		v.SetFloat(f)
	default:
		return errors.New("Unknown type " + v.Type().String())
	}
	return nil
}
