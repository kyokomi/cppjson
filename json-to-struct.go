// cppjson generates cpp struct defintions from JSON documents.
//
// Reads from stdin and prints to stdout
//
// Example:
// 	curl -s https://api.github.com/repos/kyokomi/cppjson | cppjson -name=Repository
//
// Output:
//
//  struct Hoge {
//		int a;
//		int b;
//
//		struct Fuga {
//			int b;
//		};
//		Fuga fuga;
//
//		struct Piyo {
//			int b;
//		};
//		std::vector<Piyo> piyo;
//	};
//
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"reflect"
	"sort"
	"strings"
	"unicode"
)

const templateStruct = "%s \n};\n"

var (
	name = flag.String("name", "Foo", "the name of the struct")
	pkg  = flag.String("pkg", "main", "the name of the package for the generated code")
)

// Given a JSON string representation of an object and a name structName,
// attemp to generate a struct definition
func generate(input io.Reader, structName string) ([]byte, error) {
	var iresult interface{}
	var result map[string]interface{}
	if err := json.NewDecoder(input).Decode(&iresult); err != nil {
		return nil, err
	}

	switch iresult := iresult.(type) {
	case map[string]interface{}:
		result = iresult
	case []map[string]interface{}:
		if len(iresult) > 0 {
			result = iresult[0]
		} else {
			return nil, fmt.Errorf("empty array")
		}
	default:
		return nil, fmt.Errorf("unexpected type: %T", iresult)
	}

	src := fmt.Sprintf(templateStruct,
		generateTypes(structName, result, 0))
//	formatted, err := format.Source([]byte(src))
//	if err != nil {
//		err = fmt.Errorf("error formatting: %s, was formatting\n%s", err, src)
//	}
	return []byte(src), nil
}

// Generate go struct entries for a map[string]interface{} structure
func generateTypes(structName string, obj map[string]interface{}, depth int) string {
	structure := fmt.Sprintf("\nstruct %s {", structName)

	keys := make([]string, 0, len(obj))
	for key := range obj {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	for _, key := range keys {
		value := obj[key]
		valueType := typeForValue(key, value)

		//If a nested value, recurse
		switch value := value.(type) {
		case []map[string]interface{}:
			valueType = "[]" + generateTypes(key, value[0], depth+1) + "\n};"
		case map[string]interface{}:
			valueType = generateTypes(key, value, depth+1) + "\n};"
		}

		fieldName := fmtFieldName(key)
		structure += fmt.Sprintf("\n%s %s;",
			valueType,
			fieldName)
	}
	return structure
}

var uppercaseFixups = map[string]bool{"id": true, "url": true}

func isSeparator(r rune) bool {
	// ASCII alphanumerics and underscore are not separators
	if r <= 0x7F {
		switch {
		case '0' <= r && r <= '9':
			return false
		case 'a' <= r && r <= 'z':
			return false
		case 'A' <= r && r <= 'Z':
			return false
		case r == '_':
			return false
		}
		return true
	}
	// Letters and digits are not separators
	if unicode.IsLetter(r) || unicode.IsDigit(r) {
		return false
	}
	// Otherwise, all we can do for now is treat spaces as separators.
	return unicode.IsSpace(r)
}

func Camel(s string) string {
	// Use a closure here to remember state.
	// Hackish but effective. Depends on Map scanning in order and calling
	// the closure once per rune.
	prev := ' '
	return strings.Map(
			func(r rune) rune {
				if isSeparator(prev) {
					prev = r
					return unicode.ToLower(r)
				}
				prev = r
				return r
			},
			s)
}

// fmtFieldName formats a string as a struct key
//
// Example:
// 	fmtFieldName("foo_id")
// Output: FooID
func fmtFieldName(s string) string {
	parts := strings.Split(s, "_")
	for i := range parts {
		parts[i] = strings.Title(parts[i])
	}
	if len(parts) > 0 {
		last := parts[len(parts)-1]
		if uppercaseFixups[strings.ToLower(last)] {
			parts[len(parts)-1] = strings.ToUpper(last)
		}
	}
	assembled := strings.Join(parts, "")
	runes := []rune(assembled)
	for i, c := range runes {
		ok := unicode.IsLetter(c) || unicode.IsDigit(c)
		if i == 0 {
			ok = unicode.IsLetter(c)
		}
		if !ok {
			runes[i] = '_'
		}
	}
	return Camel(string(runes))
}

var cppTypeMapping = map[string]string{
	"string": "std::string",
	"float64": "float",
	"int64": "int64_t",
}

// generate an appropriate struct type entry
func typeForValue(key string, value interface{}) string {
	//Check if this is an array
	if objects, ok := value.([]interface{}); ok {
		types := make(map[reflect.Type]bool, 0)
		for _, o := range objects {
			types[reflect.TypeOf(o)] = true
		}
		if len(types) == 1 {
			return typeForValue(key, objects[0]) + fmt.Sprintf("\nstd::vector<%s>", key)
		}
		return "[]interface{}"
	} else if object, ok := value.(map[string]interface{}); ok {
		return generateTypes(key, object, 0) + "\n};"
	} else if reflect.TypeOf(value) == nil {
		return "interface{}"
	}

	convert := cppTypeMapping[reflect.TypeOf(value).Name()]
	if convert != "" {
		return convert
	}
	return reflect.TypeOf(value).Name()
}

// Return true if os.Stdin appears to be interactive
func isInteractive() bool {
	fileInfo, err := os.Stdin.Stat()
	if err != nil {
		return false
	}
	return fileInfo.Mode()&(os.ModeCharDevice|os.ModeCharDevice) != 0
}

func main() {
	flag.Parse()

	if isInteractive() {
		flag.Usage()
		fmt.Fprintln(os.Stderr, "Expects input on stdin")
		os.Exit(1)
	}

	if output, err := generate(os.Stdin, *name); err != nil {
		fmt.Fprintln(os.Stderr, "error parsing", err)
		os.Exit(1)
	} else {
		fmt.Print(string(output))
	}
}
