package main

import (
	"bufio"
	"encoding/csv"
	"errors"
	"flag"
	"fmt"
	"os"
	"path"
	"strings"
)

const (
	cspliter     string = ","
	fspliter     string = ":"
	trimset      string = " "
	leadingSpace string = "    "
	space        string = " "
)

type FieldInfo struct {
	FieldName string
	Tag       string
	Type      string
}

func main() {
	var src = flag.String("s", "", "use -s to specific csv source.")
	var file = flag.String("f", "", "use -f to specific file used for generating corresponding go code.")
	var cols = flag.String("c", "", "use -c to specific columns needed. eg, -c col1,col2,col3..")
	var tar = flag.String("t", "", "use -t to specific target file.")
	var name = flag.String("n", "", "use -n to give type name.")
	var packName = flag.String("p", "script", "use -p give package name")

	flag.Parse()

	srcFile, err := getSource(src)
	if err != nil {
		fmt.Println(err)
		return
	}

	parseFile, err := getParseFile(file)

	var fields []FieldInfo

	if err != nil {
		if len(*cols) == 0 {
			fmt.Println(err)
			return
		} else {
			fields, err = getColumns(cols)
			if err != nil {
				fmt.Println(err)
				return
			}
		}
	} else {
		fields, err = getFileColumns(parseFile)
		if err != nil {
			fmt.Println(err)
			return
		}
	}

	tarFile, err := getTargetFileName(src, tar)
	if err != nil {
		fmt.Println(err)
		return
	}

	className := getClassName(srcFile, name)

	packageName, err := getPackageName(packName)
	if err != nil {
		fmt.Println(err)
		return
	}

	err = generateCode(srcFile, tarFile, fields, packageName, className)

	if err == nil {
		fmt.Println("generate code sucess.")
	} else {
		fmt.Println(err)
	}
}

func getSource(src *string) (string, error) {
	if len(*src) == 0 {
		return "", errors.New("source is not given, please give a csv file with -s.")
	}

	if _, err := os.Stat(*src); err != nil {
		return "", err
	}

	return *src, nil
}

func getParseFile(file *string) (string, error) {
	if len(*file) == 0 {
		return "", errors.New("Columns are not specified, please use -f or -c to give selected columns.")
	}

	if _, err := os.Stat(*file); err != nil {
		return "", err
	}

	return *file, nil
}

func getFileColumns(file string) ([]FieldInfo, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	scaner := bufio.NewScanner(f)
	var results []FieldInfo
	for scaner.Scan() {
		row := scaner.Text()
		info := strings.Split(row, fspliter)
		if len(info) != 2 {
			return nil, errors.New("Incorrect field info." + row)
		}

		field := FieldInfo{
			FieldName: strings.Trim(info[0], trimset),
			Tag:       strings.Trim(info[1], trimset),
		}
		results = append(results, field)
	}

	if err := scaner.Err(); err != nil {
		return nil, err
	}
	return results, nil
}

func getColumns(cols *string) ([]FieldInfo, error) {
	var results []FieldInfo
	if len(*cols) == 0 {
		return nil, errors.New("Columns are not specified, please use -f or -c to give selected columns.")
	}

	columns := strings.Split(*cols, cspliter)
	for i := 0; i < len(columns); i++ {
		inf := FieldInfo{
			FieldName: strings.Trim(columns[i], trimset),
		}
		results = append(results, inf)
	}
	return results, nil
}

func getTargetFileName(src *string, tar *string) (string, error) {
	if len(*tar) != 0 {
		return strings.Trim(*tar, trimset), nil
	}

	result := strings.Replace(*src, ".csv", ".go", 1)
	return result, nil
}

func generateCode(csvFile string, tarFile string, fields []FieldInfo, packageName string, className string) error {
	err := setFiledType(csvFile, fields)
	if err != nil {
		return err
	}

	f, err := os.Create(tarFile)
	if err != nil {
		return err
	}
	defer f.Close()

	writer := bufio.NewWriter(f)
	writer.WriteString("package " + packageName + "\n\n")
	writer.WriteString("type " + className + " struct {\n")
	for i := 0; i < len(fields); i++ {
		writer.WriteString(leadingSpace + fields[i].FieldName + space + fields[i].Type)
		if len(fields[i].Tag) != 0 {
			writer.WriteString(space + "`csv:" + "\"" + fields[i].Tag + "\"`")
		}
		writer.WriteString("\n")
	}
	writer.WriteString("}\n")
	writer.Flush()
	return nil
}

func setFiledType(csvFile string, fields []FieldInfo) error {
	f, err := os.Open(csvFile)
	if err != nil {
		return err
	}
	defer f.Close()

	reader := csv.NewReader(f)

	columns, err := reader.Read()
	if err != nil {
		return err
	}

	typeInfo, err := reader.Read()
	if err != nil {
		return err
	}

	for i := 0; i < len(fields); i++ {
		matched := false
		for j := 0; j < len(columns); j++ {
			if strings.Trim(columns[j], trimset) == fields[i].Tag {
				fields[i].Type = strings.Trim(typeInfo[j], trimset)
				matched = true
			} else if strings.Trim(columns[j], trimset) == fields[i].FieldName {
				fields[i].Type = strings.Trim(typeInfo[j], trimset)
				matched = true
			}
		}
		if !matched {
			fmt.Println("Field mismatch. FieldName=", fields[i].FieldName, ",Tag:", fields[i].Tag)
		}
	}

	return nil
}

func getClassName(csvFile string, name *string) string {
	if len(*name) != 0 {
		return path.Base(*name)
	}
	return strings.Replace(path.Base(csvFile), ".csv", "", 1)
}

func getPackageName(pack *string) (string, error) {
	if len(*pack) == 0 {
		return "", errors.New("Error package name.")
	}

	return *pack, nil
}
