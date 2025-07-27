package model

import (
	"strings"
)

// Assignment represents an assignment between fields in a struct.
type Assignment interface {
	// String returns the string representation of the assignment.
	String() string
	// RetError returns whether the assignment returns an error value.
	RetError() bool
}

// 不做任何加工处理
type RawAssignment struct {
	Raw string
	Err bool
}

func (s RawAssignment) String() string {
	return s.Raw
}

func (s RawAssignment) RetError() bool {
	return s.Err
}

// 组合多个Assignment
type RepeatAssignment struct {
	Assignments []Assignment
}

func (s RepeatAssignment) String() string {
	sb := strings.Builder{}
	for _, as := range s.Assignments {
		_, _ = sb.WriteString(as.String())
	}

	return sb.String()
}

func (s RepeatAssignment) RetError() bool {
	// 任意一个Assignment返回错误， 总的Assignment就返回错误
	for _, as := range s.Assignments {
		if as.RetError() {
			return true
		}
	}

	return false
}

// SkipField indicates that the field is skipped due to a :skip notation.
type SkipField struct {
	LHS string // LHS is the left-hand side of the skipped field.
}

// String returns the string representation of the skip field assignment.
func (s SkipField) String() string {
	var sb strings.Builder
	sb.WriteString("// skip: ")
	sb.WriteString(s.LHS)
	sb.WriteString("\n")
	return sb.String()
}

// RetError always returns false for skip field assignments.
func (s SkipField) RetError() bool {
	return false
}

// NoMatchField indicates that the field is skipped while there was no matching fields or getters.
type NoMatchField struct {
	LHS string // LHS is the name of the field that doesn't match any fields or getters.
	RHS string // 可选
}

// String returns the string representation of the no match field assignment.
func (s NoMatchField) String() string {
	var sb strings.Builder
	sb.WriteString("// no match: ")
	sb.WriteString(s.LHS)
	if s.RHS != "" {
		sb.WriteString("( from left value '")
		sb.WriteString(s.RHS)
		sb.WriteString("')")
	}
	sb.WriteString("\n")
	return sb.String()
}

// RetError always returns false for no match field assignments.
func (s NoMatchField) RetError() bool {
	return false
}

// SimpleField represents an RHS expression.
type SimpleField struct {
	LHS   string
	RHS   string
	Error bool
}

// String returns the string representation of the simple field assignment.
func (s SimpleField) String() string {
	var sb strings.Builder
	sb.WriteString(s.LHS)
	if s.Error {
		sb.WriteString(", err")
	}
	sb.WriteString(" = ")
	sb.WriteString(s.RHS)
	sb.WriteString("\n")
	return sb.String()
}

// RetError returns whether the assignment returns an error value.
func (s SimpleField) RetError() bool {
	return s.Error
}

// NestStruct represents a struct in a struct.
type NestStruct struct {
	InitExpr      string
	NullCheckExpr string
	Contents      []Assignment
}

// String returns the string representation of the nested struct assignment.
func (s NestStruct) String() string {
	var sb strings.Builder
	if s.NullCheckExpr != "" {
		sb.WriteString("if ")
		sb.WriteString(s.NullCheckExpr)
		sb.WriteString(" != nil {\n")
	}
	if s.InitExpr != "" {
		sb.WriteString(s.InitExpr)
		sb.WriteString("\n")
	}
	for _, content := range s.Contents {
		sb.WriteString(content.String())
	}
	if s.NullCheckExpr != "" {
		sb.WriteString("}\n")
	}
	return sb.String()
}

// RetError returns whether the assignment returns an error value.
func (s NestStruct) RetError() bool {
	return false
}

// SliceAssignment represents a slice assignment.
type SliceAssignment struct {
	LHS string
	RHS string
	Typ string
}

// String returns the string representation of the slice assignment.
func (c SliceAssignment) String() string {
	var sb strings.Builder
	sb.WriteString("if ")
	sb.WriteString(c.RHS)
	sb.WriteString(" != nil {\n")
	sb.WriteString(c.LHS)
	sb.WriteString(" = make(")
	sb.WriteString(c.Typ)
	sb.WriteString(", len(")
	sb.WriteString(c.RHS)
	sb.WriteString("))\ncopy(")
	sb.WriteString(c.LHS)
	sb.WriteString(", ")
	sb.WriteString(c.RHS)
	sb.WriteString(")\n}\n")
	return sb.String()
}

// RetError returns whether the assignment returns an error value.
func (c SliceAssignment) RetError() bool {
	return false
}

// SliceLoopAssignment represents a slice assignment with a loop.
type SliceLoopAssignment struct {
	LHS string
	RHS string
	Typ string
}

// String returns the string representation of the slice assignment with a loop.
func (c SliceLoopAssignment) String() string {
	var sb strings.Builder
	sb.WriteString("if ")
	sb.WriteString(c.RHS)
	sb.WriteString(" != nil {\n")
	sb.WriteString(c.LHS)
	sb.WriteString(" = make(")
	sb.WriteString(c.Typ)
	sb.WriteString(", len(")
	sb.WriteString(c.RHS)
	sb.WriteString("))\nfor i, e := range ")
	sb.WriteString(c.RHS)
	sb.WriteString("{\n")
	sb.WriteString(c.LHS)
	sb.WriteString("[i] = e\n}\n}\n")
	return sb.String()
}

// RetError returns whether the assignment returns an error value.
func (c SliceLoopAssignment) RetError() bool {
	return false
}

// SliceTypecastAssignment represents a slice assignment with a typecast.
type SliceTypecastAssignment struct {
	LHS   string
	RHS   string
	Typ   string
	Cast  string
	Error bool
}

// String returns the string representation of the slice assignment with a typecast.
func (c SliceTypecastAssignment) String() string {
	var sb strings.Builder
	sb.WriteString("if ")
	sb.WriteString(c.RHS)
	sb.WriteString(" != nil {\n")
	sb.WriteString(c.LHS)
	sb.WriteString(" = make(")
	sb.WriteString(c.Typ)
	sb.WriteString(", len(")
	sb.WriteString(c.RHS)
	sb.WriteString("))\nfor i, e := range ")
	sb.WriteString(c.RHS)
	sb.WriteString("{\n")
	sb.WriteString(c.LHS)
	sb.WriteString("[i]")
	if c.Error {
		sb.WriteString(", err")
	}
	sb.WriteString(" = ")
	sb.WriteString(c.Cast)
	sb.WriteString("(e)\n")
	if c.Error {
		sb.WriteString("if err != nil {\nreturn\n}")
	}
	sb.WriteString("}\n}\n")
	return sb.String()
}

// RetError returns whether the assignment returns an error value.
func (c SliceTypecastAssignment) RetError() bool {
	return false
}

// SliceMethodCallAssignment represents a slice assignment with a typecast.
type SliceMethodCallAssignment struct {
	LHS      string
	RHS      string
	Typ      string
	Method   string
	Nullable bool
	Error    bool
}

// String returns the string representation of the slice assignment with a typecast.
func (c SliceMethodCallAssignment) String() string {
	var sb strings.Builder
	sb.WriteString("if ")
	sb.WriteString(c.RHS)
	sb.WriteString(" != nil {\n")
	sb.WriteString(c.LHS)
	sb.WriteString(" = make(")
	sb.WriteString(c.Typ)
	sb.WriteString(", len(")
	sb.WriteString(c.RHS)
	sb.WriteString("))\nfor i, e := range ")
	sb.WriteString(c.RHS)
	sb.WriteString("{\n")
	if c.Nullable {
		sb.WriteString("if e != nil {")
	}
	sb.WriteString(c.LHS)
	sb.WriteString("[i]")
	if c.Error {
		sb.WriteString(", err")
	}
	sb.WriteString(" = e.")
	sb.WriteString(c.Method)
	sb.WriteString("()\n")
	if c.Error {
		sb.WriteString("if err != nil {\nreturn\n}")
	}
	if c.Nullable {
		sb.WriteString("}")
	}
	sb.WriteString("}\n}\n")
	return sb.String()
}

// RetError returns whether the assignment returns an error value.
func (c SliceMethodCallAssignment) RetError() bool {
	return false
}

// IfAssignment represents if check assignment
type IfAssignment struct {
	Inner    Assignment
	Nullable bool
	Expr     string
}

// String returns the string representation of the nested struct assignment.
func (s IfAssignment) String() string {
	if !s.Nullable {
		return s.Inner.String()
	}

	var sb strings.Builder
	sb.WriteString("if ")
	sb.WriteString(s.Expr)
	sb.WriteString(" != nil {\n")
	sb.WriteString(s.Inner.String())
	sb.WriteString("}\n")

	return sb.String()
}

// RetError returns whether the assignment returns an error value.
func (s IfAssignment) RetError() bool {
	return s.Inner.RetError()
}
