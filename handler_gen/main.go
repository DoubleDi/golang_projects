package main

import (
	"encoding/json"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"log"
	"os"
	"reflect"
	"strconv"
	"strings"
)

type MethodParams struct {
	Url    string
	Auth   bool
	Method string
}

// go build ./handler_gen/* && ./main site/api.go site/api_handlers.go && gofmt -w site/api_handlers.go

func main() {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, os.Args[1], nil, parser.ParseComments)
	if err != nil {
		log.Fatal(err)
	}

	out, _ := os.Create("/tmp/123123123123.go")

	fmt.Fprintln(out, `package `+node.Name.Name)
	fmt.Fprintln(out)
	fmt.Fprintln(out, `import "encoding/json"`)
	fmt.Fprintln(out, `import "net/http"`)
	fmt.Fprintln(out, `import "strconv"`)
	fmt.Fprintln(out, `import "fmt"`)
	fmt.Fprintln(out, `import "context"`)
	fmt.Fprintln(out)
	fmt.Fprintln(out, "type Response struct {")
	fmt.Fprintln(out, "Error    string        `json:\"error\"`")
	fmt.Fprintln(out, "Response interface{}   `json:\"response,omitempty\"`")
	fmt.Fprintln(out, "}")
	fmt.Fprintln(out)
	fmt.Fprintln(out, `func MakeError(text string) string {
	errResp := &Response{Error: text}
	errStr, err := json.Marshal(errResp)

	if err != nil {
		panic("JSON ERROR")
	}

	return string(errStr)
}

func MakeResponse(error string, response interface{}) string {
	resp := &Response{Error: error, Response: response}
	respStr, err := json.Marshal(resp)

	if err != nil {
		panic("JSON ERROR")
	}

	return string(respStr)
}
`)

	serveHTTPHandlers := map[string](map[string]string){}
	starExprs := map[string]bool{}

	for _, f := range node.Decls {
		switch f.(type) {
		case *ast.GenDecl:
			g, _ := f.(*ast.GenDecl)
			for _, spec := range g.Specs {
				currType, ok := spec.(*ast.TypeSpec)
				if !ok {
					continue
				}

				currStruct, ok := currType.Type.(*ast.StructType)
				if !ok {
					continue
				}

				fmt.Fprintln(out, "func Validate"+currType.Name.Name+"(w http.ResponseWriter, r *http.Request) ("+currType.Name.Name+", error) {")
				fmt.Fprintln(out, "str := "+currType.Name.Name+"{}")

				for _, field := range currStruct.Fields.List {
					var validator string

					if field.Tag != nil {
						tag := reflect.StructTag(field.Tag.Value[1 : len(field.Tag.Value)-1])
						validator = tag.Get("apivalidator")
					}

					if validator == "" {
						continue
					}

					fieldName := field.Names[0].Name
					typeName := field.Type.(*ast.Ident).Name

					validations := strings.Split(validator, ",")
					validationsMap := make(map[string]interface{})
					for _, v := range validations {
						if strings.HasPrefix(v, "required") {
							validationsMap["required"] = true
						} else if strings.HasPrefix(v, "paramname") {
							validationsMap["paramname"] = strings.TrimPrefix(v, "paramname=")
						} else if strings.HasPrefix(v, "enum") {
							validationsMap["enum"] = strings.Split(strings.TrimPrefix(v, "enum="), "|")
						} else if strings.HasPrefix(v, "default") {
							validationsMap["default"] = strings.TrimPrefix(v, "default=")
						} else if strings.HasPrefix(v, "min") {
							minValue, err := strconv.Atoi(strings.TrimPrefix(v, "min="))
							if err != nil {
								panic("Bad number in min " + strings.TrimPrefix(v, "min="))
							}
							validationsMap["min"] = minValue
						} else if strings.HasPrefix(v, "max") {
							maxValue, err := strconv.Atoi(strings.TrimPrefix(v, "max="))
							if err != nil {
								panic("Bad number in max " + strings.TrimPrefix(v, "max="))
							}
							validationsMap["max"] = maxValue
						}
					}

					fmt.Fprintln(out, "// Field "+fieldName)
					variable := strings.ToLower(fieldName)

					if str, ok := validationsMap["paramname"].(string); ok {
						fmt.Fprintln(out, "// Custom parameter name")
						fmt.Fprintln(out, variable+` := r.FormValue("`+str+`")`)
						fmt.Fprintln(out)
					} else {
						fmt.Fprintln(out, "// Ordinary parameter name")
						fmt.Fprintln(out, variable+` := r.FormValue("`+strings.ToLower(fieldName)+`")`)
						fmt.Fprintln(out)
					}

					if req, ok := validationsMap["required"].(bool); ok && req {
						fmt.Fprintln(out, "// Required validation")
						fmt.Fprintln(out, "if "+variable+` == "" {`)
						fmt.Fprintln(out, `return str, &ApiError{HTTPStatus: http.StatusBadRequest, Err: fmt.Errorf("`+variable+` must me not empty")}`)
						fmt.Fprintln(out, "}")
						fmt.Fprintln(out)
					}

					if typeName == "int" {
						fmt.Fprintln(out, "// Type cast to int")
						fmt.Fprintln(out, variable+"Value, err := strconv.Atoi("+variable+")")
						fmt.Fprintln(out, "if err != nil {")
						fmt.Fprintln(out, `return str, &ApiError{HTTPStatus: http.StatusBadRequest, Err: fmt.Errorf("`+variable+` must be int")}`)
						fmt.Fprintln(out, "}")
						fmt.Fprintln(out)
					} else if typeName == "string" {
						fmt.Fprintln(out, "// Type cast to string")
						fmt.Fprintln(out, variable+"Value := "+variable)
						fmt.Fprintln(out)
					}

					if def, ok := validationsMap["default"].(string); ok {
						fmt.Fprintln(out, "// Check default")
						if typeName == "int" {
							fmt.Fprintln(out, "if "+variable+"Value == 0 {")
							fmt.Fprintln(out, variable+"Value = "+def)
						}
						if typeName == "string" {
							fmt.Fprintln(out, "if "+variable+`Value == "" {`)
							fmt.Fprintln(out, variable+`Value = "`+def+`"`)
						}
						fmt.Fprintln(out, "}")
						fmt.Fprintln(out)
					}

					if enum, ok := validationsMap["enum"]; ok {
						fmt.Fprintln(out, "// Check enum")
						enumSlice := enum.([]string)
						if typeName == "int" {
							fmt.Fprintln(out, variable+"Enum := []int{"+strings.Join(enumSlice, ", ")+"}")
						}
						if typeName == "string" {
							fmt.Fprintln(out, variable+`Enum := []string{"`+strings.Join(enumSlice, `", "`)+`"}`)
						}

						fmt.Fprintln(out, "enumFlag := false")
						fmt.Fprintln(out, `for _, v := range `+variable+`Enum {`)
						fmt.Fprintln(out, `if v == `+variable+`Value {`)
						fmt.Fprintln(out, `enumFlag = true`)
						fmt.Fprintln(out, `}`)
						fmt.Fprintln(out, `}`)
						fmt.Fprintln(out)
						fmt.Fprintln(out, `if !enumFlag {`)
						fmt.Fprintln(out, `return str, &ApiError{HTTPStatus: http.StatusBadRequest, Err: fmt.Errorf("`+variable+` must be one of [`+strings.Join(enumSlice, ", ")+`]")}`)
						fmt.Fprintln(out, `}`)
						fmt.Fprintln(out)
					}

					if min, ok := validationsMap["min"]; ok {
						fmt.Fprintln(out, "// Check min")
						minValue := min.(int)
						minValueStr := strconv.Itoa(minValue)
						fmt.Fprintln(out, variable+"Min := "+minValueStr)
						if typeName == "int" {
							fmt.Fprintln(out, `if `+variable+`Min > `+variable+`Value {`)
							fmt.Fprintln(out, `return str, &ApiError{HTTPStatus: http.StatusBadRequest, Err: fmt.Errorf("`+variable+` must be >= `+minValueStr+`")}`)
							fmt.Fprintln(out, `}`)
							fmt.Fprintln(out)
						}
						if typeName == "string" {
							fmt.Fprintln(out, `if `+variable+`Min > len(`+variable+`Value) {`)
							fmt.Fprintln(out, `return str, &ApiError{HTTPStatus: http.StatusBadRequest, Err: fmt.Errorf("`+variable+` len must be >= `+minValueStr+`")}`)
							fmt.Fprintln(out, `}`)
							fmt.Fprintln(out)
						}
					}

					if max, ok := validationsMap["max"]; ok {
						fmt.Fprintln(out, "// Check max")
						maxValue := max.(int)
						maxValueStr := strconv.Itoa(maxValue)
						fmt.Fprintln(out, variable+"Max := "+maxValueStr)
						if typeName == "int" {
							fmt.Fprintln(out, `if `+variable+`Max < `+variable+`Value {`)
							fmt.Fprintln(out, `return str, &ApiError{HTTPStatus: http.StatusBadRequest, Err: fmt.Errorf("`+variable+` must be <= `+maxValueStr+`")}`)
							fmt.Fprintln(out, `}`)
							fmt.Fprintln(out)
						}
						if typeName == "string" {
							fmt.Fprintln(out, `if `+variable+`Max < len(`+variable+`Value) {`)
							fmt.Fprintln(out, `return str, &ApiError{HTTPStatus: http.StatusBadRequest, Err: fmt.Errorf("`+variable+` len must be <= `+maxValueStr+`")}`)
							fmt.Fprintln(out, `}`)
							fmt.Fprintln(out)
						}
					}
					fmt.Fprintln(out)
					fmt.Fprintln(out, "str."+fieldName+" = "+variable+"Value")
					fmt.Fprintln(out)
					fmt.Fprintln(out)

				}

				fmt.Fprintln(out, "return str, nil")
				fmt.Fprintln(out, "}")
				fmt.Fprintln(out, "")
			}

		case *ast.FuncDecl:
			fun, _ := f.(*ast.FuncDecl)
			methodStruct := ""
			isStarExpr := false
			if starExpr, ok := fun.Recv.List[0].Type.(*ast.StarExpr); ok {
				methodStruct = starExpr.X.(*ast.Ident).Name
				isStarExpr = true
			}
			if ident, ok := fun.Recv.List[0].Type.(*ast.Ident); ok {
				methodStruct = ident.Name
			}

			doc := fun.Doc
			if doc == nil {
				continue
			} else {
				if !strings.HasPrefix(doc.List[0].Text, "// apigen:api ") {
					continue
				}
				starExprs[methodStruct] = isStarExpr
				if serveHTTPHandlers[methodStruct] == nil {
					serveHTTPHandlers[methodStruct] = make(map[string]string)
				}
				if isStarExpr {
					fmt.Fprintln(out, "func (str * "+methodStruct+") handler"+fun.Name.Name+"(w http.ResponseWriter, r *http.Request) {")
				} else {
					fmt.Fprintln(out, "func (str "+methodStruct+") handler"+fun.Name.Name+"(w http.ResponseWriter, r *http.Request) {")

				}
				for _, comment := range doc.List {
					if strings.HasPrefix(comment.Text, "// apigen:api ") {
						jsonStr := strings.TrimPrefix(comment.Text, "// apigen:api ")
						params := &MethodParams{}
						err := json.Unmarshal([]byte(jsonStr), params)

						if err != nil {
							panic("JSON ERROR" + jsonStr)
							continue
						}

						if params.Method != "" {
							fmt.Fprintln(out, "// check method "+params.Method)
							fmt.Fprintln(out, `if r.Method != http.MethodPost {`)
							fmt.Fprintln(out, "w.WriteHeader(http.StatusNotAcceptable)")
							fmt.Fprintln(out, `fmt.Fprintln(w, MakeError("bad method"))`)
							fmt.Fprintln(out, `return`)
							fmt.Fprintln(out, "}")
							fmt.Fprintln(out)
						}

						if params.Auth {
							fmt.Fprintln(out, "// check auth ")
							fmt.Fprintln(out, `auth := r.Header.Get("X-Auth")`)

							fmt.Fprintln(out, `if auth == "" {`)
							fmt.Fprintln(out, "w.WriteHeader(http.StatusForbidden)")
							fmt.Fprintln(out, `fmt.Fprintln(w, MakeError("unauthorized"))`)
							fmt.Fprintln(out, "return")
							fmt.Fprintln(out, "}")
							fmt.Fprintln(out)
						}

						fmt.Fprintln(out, "// params validation ")
						fmt.Fprintln(out, "params, err := Validate"+fun.Name.Name+"Params(w, r)")
						fmt.Fprintln(out, `if err != nil {`)
						fmt.Fprintln(out, `err := err.(*ApiError)`)
						fmt.Fprintln(out, "w.WriteHeader(err.HTTPStatus)")
						fmt.Fprintln(out, `fmt.Fprintln(w, MakeError(err.Error()))`)
						fmt.Fprintln(out, `return`)
						fmt.Fprintln(out, "}")
						fmt.Fprintln(out)

						fmt.Fprintln(out)
						fmt.Fprintln(out, "ctx, _ := context.WithCancel(context.Background())")
						fmt.Fprintln(out, "res, err := str."+fun.Name.Name+"(ctx, params)")
						fmt.Fprintln(out, "switch err.(type) {")
						fmt.Fprintln(out, `case ApiError:`)
						fmt.Fprintln(out, "err := err.(ApiError)")
						fmt.Fprintln(out, "w.WriteHeader(err.HTTPStatus)")
						fmt.Fprintln(out, "fmt.Fprintf(w, MakeError(err.Error()))")
						fmt.Fprintln(out, `case error:`)
						fmt.Fprintln(out, "w.WriteHeader(http.StatusInternalServerError)")
						fmt.Fprintln(out, "fmt.Fprintf(w, MakeError(err.Error()))")
						fmt.Fprintln(out, `default:`)
						fmt.Fprintln(out, `fmt.Fprintf(w, MakeResponse("", res))`)
						fmt.Fprintln(out, "}")

						fmt.Fprintln(out, "}")
						fmt.Fprintln(out)
						serveHTTPHandlers[methodStruct][params.Url] = fun.Name.Name
					}
				}
			}
		}
	}

	for k, v := range serveHTTPHandlers {
		if starExprs[k] {
			fmt.Fprintln(out, "func (str * "+k+") ServeHTTP(w http.ResponseWriter, r *http.Request) {")
		} else {
			fmt.Fprintln(out, "func (str "+k+") ServeHTTP(w http.ResponseWriter, r *http.Request) {")
		}
		fmt.Fprintln(out, "query := r.URL.Path")
		fmt.Fprintln(out, "switch query {")
		for url, handler := range v {
			fmt.Fprintln(out, "case \""+url+"\":")
			fmt.Fprintln(out, "str.handler"+handler+"(w, r)")
		}
		fmt.Fprintln(out, "default:")
		fmt.Fprintln(out, "w.WriteHeader(http.StatusNotFound)")
		fmt.Fprintln(out, "fmt.Fprintf(w, MakeError(`unknown method`))")
		fmt.Fprintln(out, "}")
		fmt.Fprintln(out, "}")
	}

	out.Close()

	f, err := os.Open("/tmp/123123123123.go")
	if err != nil {
		panic("Where is my file??!?!")
	}
	fi, err := f.Stat()
	if err != nil {
		panic("Where is my stat??!?!")
	}

	fileContent := make([]byte, fi.Size())
	f.Read(fileContent)
	f.Close()

	os.Remove("/tmp/123123123123.go")
	if err != nil {
		panic("Cant remove??!?!")
	}

	formatedFileContent, err := format.Source(fileContent)
	if err != nil {
		panic("Bad formating??!?!")
	}

	final, _ := os.Create(os.Args[2])
	defer final.Close()

	fmt.Fprintf(final, string(formatedFileContent))
}
