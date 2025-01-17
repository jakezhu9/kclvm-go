package gen

import (
	"bytes"
	_ "embed"
	"fmt"
	"io"
	"strings"
	"text/template"
)

var (
	//go:embed templates/kcl/data.gotmpl
	dataTmpl string
	//go:embed templates/kcl/document.gotmpl
	documentTmpl string
	//go:embed templates/kcl/header.gotmpl
	headerTmpl string
	//go:embed templates/kcl/validator.gotmpl
	validatorTmpl string
	//go:embed templates/kcl/schema.gotmpl
	schemaTmpl string
	//go:embed templates/kcl/index.gotmpl
	indexTmpl string
)

var funcs = template.FuncMap{
	"formatType":  formatType,
	"formatValue": formatValue,
	"formatName":  formatName,
	"indentLines": indentLines,
	"isKclData": func(v interface{}) bool {
		_, ok := v.([]data)
		return ok
	},
	"isArray": func(v interface{}) bool {
		_, ok := v.([]interface{})
		return ok
	},
}

func (k *kclGenerator) genKcl(w io.Writer, s kclFile) error {
	tmpl := &template.Template{}

	// add "include" function. It works like "template" but can be used in pipeline.
	funcs["include"] = func(name string, data interface{}) (string, error) {
		buf := bytes.NewBuffer(nil)
		if err := tmpl.ExecuteTemplate(buf, name, data); err != nil {
			return "", err
		}
		return buf.String(), nil
	}

	tmpl = addTemplate(tmpl, "data", dataTmpl)
	tmpl = addTemplate(tmpl, "document", documentTmpl)
	tmpl = addTemplate(tmpl, "header", headerTmpl)
	tmpl = addTemplate(tmpl, "validator", validatorTmpl)
	tmpl = addTemplate(tmpl, "schema", schemaTmpl)
	tmpl = addTemplate(tmpl, "index", indexTmpl)
	return tmpl.Funcs(funcs).Execute(w, s)
}

func addTemplate(tmpl *template.Template, name, data string) *template.Template {
	newTmpl := template.Must(template.New(name).Funcs(funcs).Parse(data))
	return template.Must(tmpl.AddParseTree(name, newTmpl.Tree))
}

func formatType(t typeInterface) string {
	return t.Format()
}

func formatValue(v interface{}) string {
	if v == nil {
		return "None"
	}
	switch value := v.(type) {
	case string:
		return fmt.Sprintf("\"%s\"", value)
	case bool:
		if value {
			return "True"
		}
		return "False"
	case map[string]interface{}:
		var s strings.Builder
		for _, key := range getSortedKeys(value) {
			if s.Len() != 0 {
				s.WriteString(", ")
			}
			s.WriteString(fmt.Sprintf("%s: %s", formatValue(key), formatValue(value[key])))
		}
		return "{" + s.String() + "}"
	default:
		return fmt.Sprintf("%v", value)
	}
}

var kclKeywords = map[string]struct{}{
	"True":      {},
	"False":     {},
	"None":      {},
	"Undefined": {},
	"import":    {},
	"and":       {},
	"or":        {},
	"in":        {},
	"is":        {},
	"not":       {},
	"as":        {},
	"if":        {},
	"else":      {},
	"elif":      {},
	"for":       {},
	"schema":    {},
	"mixin":     {},
	"protocol":  {},
	"check":     {},
	"assert":    {},
	"all":       {},
	"any":       {},
	"map":       {},
	"filter":    {},
	"lambda":    {},
	"rule":      {},
}

func formatName(name string) string {
	if _, ok := kclKeywords[name]; ok {
		return fmt.Sprintf("$%s", name)
	}
	return name
}

func indentLines(s, indent string) string {
	s = strings.Replace(s, "\r\n", "\n", -1)
	n := strings.Count(s, "\n")
	if s[len(s)-1] == '\n' {
		// ignore last empty line
		n--
	}
	return indent + strings.Replace(s, "\n", "\n"+indent, n)
}
