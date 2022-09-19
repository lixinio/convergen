package generator

import (
	"bytes"
	"fmt"
	"go/format"
	"os"

	"github.com/reedom/convergen/pkg/model"
	"golang.org/x/tools/imports"
)

type Generator struct {
	code model.Code
}

func NewGenerator(code model.Code) *Generator {
	return &Generator{
		code: code,
	}
}

func (g *Generator) Generate(outPath string, output, dryRun bool) ([]byte, error) {
	content, err := g.generateContent()
	if err != nil {
		return nil, err
	}

	optimized, err := imports.Process(outPath, content, nil)
	if err != nil {
		if output {
			fmt.Println(string(content))
		}
		return nil, fmt.Errorf("error on optimizing imports of the generated code.\n%w", err)
	}

	formatted, err := format.Source(optimized)
	if err != nil {
		if output {
			fmt.Println(string(content))
		}
		return nil, fmt.Errorf("error on formatting the generated code.\n%w", err)
	}

	if dryRun {
		if output {
			fmt.Println(string(content))
		}
		return formatted, nil
	}

	err = os.WriteFile(outPath, formatted, 0644)
	if err != nil {
		return nil, fmt.Errorf("error on writing to the file.\n%w", err)
	}

	return formatted, nil
}

func (g *Generator) generateContent() (content []byte, err error) {
	buf := bytes.Buffer{}
	_, err = buf.WriteString("// Code generated by github.com/reedom/convergen\n// DO NOT EDIT.\n\n")
	if err != nil {
		return
	}
	_, err = buf.WriteString(g.code.Pre)
	if err != nil {
		return
	}

	for _, f := range g.code.Functions {
		_, err = buf.WriteString(g.FuncToString(f))
		if err != nil {
			return
		}
	}

	_, err = buf.WriteString(g.code.Post)
	if err != nil {
		return
	}

	return buf.Bytes(), nil
}
