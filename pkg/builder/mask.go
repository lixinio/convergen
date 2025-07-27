package builder

import (
	"fmt"
	"go/types"
	"sort"
	"strings"
	"text/template"

	bmodel "github.com/reedom/convergen/pkg/builder/model"
	gmodel "github.com/reedom/convergen/pkg/generator/model"
	"github.com/reedom/convergen/pkg/option"
	"github.com/reedom/convergen/pkg/util"
)

var structMemberTmpl = `{{- if $.EnableMaskExtension -}}
func {{ $.FuncName }}WithMask(
	src {{ $.SrcStruct}},
	existFn func() ({{.Receiver}}, error),
	masker msku.Mask,
)({{.Receiver}}, msku.Mask, error) {
	{{ if $.RetError -}}
	dst, err := {{ $.FuncName }}(src)
	if err != nil {
		return nil, nil, err
	}
	{{- else -}}
	dst := {{ $.FuncName }}(src)
	{{- end }}

	// 转换并检查是否包含mask字段
	newMasker := msku.TransferMask(masker, dst.MaskMap())
	{{ $lastIndex := ( len .DstMaskList | add -1 ) -}}
	if {{ range $i, $k := .DstMaskList -}}
	!newMasker.IsExist("{{ $k.Mask }}") 
	{{- if ne $i $lastIndex -}}&&
	{{ else }}{{ end -}}
	{{ end -}} {
		return dst, newMasker, nil // 直接退出
	}

	// 获得已经存在的 (ctx等参数通过闭包自行解决)
	old, err := existFn()
	if err != nil {
		return nil, nil, err
	}

	// 根据fieldMask 将更新的propertyMask字段更新到老字段
	for key := range masker {
		switch key {
		{{- range $k := .SrcMaskList }}
		case {{ $.MaskExtension }}.{{ $k.DstMaskName }}:
			old.Set{{ $k.SrcFlagName }}(src.{{ $k.DstMaskName }}, {{ $k.MaskName }})
		{{- end }}
		}
	}

	// 回写 (combine了未更新的propertyMask bit)
	{{- range $k := .DstMaskList }}
	dst.{{ $k.Mask }} = old.{{ $k.Mask }}
	{{- end }}
	
	return dst, newMasker, nil
}
{{- end }}

{{ if or ($.EnableMask) ($.EnableMaskExtension) -}}
func ({{.Receiver}})MaskMap() map[string]string {
	return map[string]string{
		{{- range $k := .SrcMaskList }}
		{{ $.MaskExtension }}.{{ $k.DstMaskName }}:"{{ $k.SrcFlagName }}",
		{{- end }}
	}
}
{{- end }}

{{ if $.EnableMask -}}
func {{ $.FuncName }}WithMaskToMap(
	src {{ $.SrcStruct}},
	masker msku.Mask,
	transfer interface {
		Int64(string, int64, int64) (any, error)
		Uint64(string, uint64, uint64) (any, error)
		Int32(string, int32, int32) (any, error)
		Uint32(string, uint32, uint32) (any, error)
		Int16(string, int16, int16) (any, error)
		Uint16(string, uint16, uint16) (any, error)
		Int8(string, int8, int8) (any, error)
		Uint8(string, uint8, uint8) (any, error)
	},
)(map[string]any, error) {
	var (
		{{- range $k := .DstMaskList }}
		set{{ $k.Mask }} {{ $k.Type }}
		unset{{ $k.Mask }} {{ $k.Type }}
		{{- end }}
		result = map[string]any{}
		{{- range $k := .DstMaskList }}
		set{{ $k.Mask }}Fn = func(flag bool, mask {{ $k.FlagType }}) {
			if flag {
				set{{  $k.Mask }} |= {{ $k.Type }}(mask)
			} else {
				unset{{ $k.Mask }} |= {{ $k.Type }}(mask)
			}
		}
		{{- end }}
	)

	{{ if .DstOtherFields -}}
	{{ if $.RetError -}}
	dst, err := {{ $.FuncName }}(src)
	if err != nil {
		return nil, err
	}
	{{- else -}}
	dst := {{ $.FuncName }}(src)
	{{- end }}
	{{- end }}

	for key := range masker {
		switch key {
		{{- range $k := .SrcMaskList }}
		case {{ $.MaskExtension }}.{{ $k.DstMaskName }}:
			set{{ $k.SrcFlagName }}Fn(src.{{ $k.DstMaskName }}, {{ $k.MaskName }})
		{{- end }}
		}
	}
	{{- range $k := .DstMaskList }}

	if v, err := transfer.{{ title $k.Type }}("{{ $k.Mask }}", set{{ $k.Mask }}, unset{{ $k.Mask }}); err != nil {
		return nil, err
	} else if v != nil {
		result["{{ $k.Mask }}"] = v
	}
	{{- end }}

	newMasker := msku.TransferMask(masker, ({{$.Receiver}})(nil).MaskMap())
	{{ if .DstOtherFields -}}
	newMasker = msku.TransferMaskToCamel(newMasker)
	{{- end }}

	for key := range newMasker {
		switch key {
		{{- range $k := .DstOtherFields }}
		case "{{ . }}":
			result["{{ . }}"] = dst.{{ . }}
		{{- end }}
		{{- range $k := .DstMaskList }}
		case "{{ $k.Mask }}": // skip
		{{- end }}
		default:
			return nil, fmt.Errorf("invalid mask field '%s'", key)
		}
	}

	return result, nil
}
{{- end }}

{{ range $k := .DstMaskList }}
func (dst {{$.Receiver}})Set{{ $k.Mask }}(flag bool, mask {{ $k.FlagType }}) {
	if flag {
		dst.{{ $k.Mask }} |= {{ $k.Type }}(mask)
	} else {
		dst.{{ $k.Mask }} &= {{ $k.Type }}(^mask)
	}
}

func (dst {{$.Receiver}})Get{{ $k.Mask }}(mask {{ $k.FlagType }}) bool {
	return ({{ $k.FlagType }}(dst.{{ $k.Mask }}) & mask) != 0
}
{{ end }}
`

// 自定义fieldQuery 是否包含了这些字段
func checkExternsion(opts *option.Options) (string, error) {
	var maskExtension *option.MaskExtension
	if opts.MaskExtension != nil {
		maskExtension = opts.MaskExtension
	} else if opts.Mask != nil {
		maskExtension = opts.Mask.MaskExtension
	}

	if maskExtension == nil || maskExtension.Name == "" {
		return "", nil
	}

	fields := map[string]*types.Var{}
	for i := 0; i < maskExtension.Struct.NumFields(); i++ {
		field := maskExtension.Struct.Field(i)
		fields[field.Name()] = field
	}

	for _, c := range opts.BuildMaskConverters {
		if _, ok := fields[c.Src().NameAt(0)]; !ok {
			return "", fmt.Errorf(
				"field '%s' NOT in struct '%s' varible",
				c.Src().NameAt(0), maskExtension.Name,
			)
		}
	}

	return maskExtension.Name, nil
}

// 获得右值的struct对象，方便遍历所有的field
func (b *assignmentBuilder) getRhsStruct(rhsStruct bmodel.Node) (*types.Struct, error) {
	tp := util.DerefPtr(rhsStruct.ExprType())
	ntp, ok := tp.(*types.Named)
	if !ok {
		return nil, fmt.Errorf("dst obj NOT struct '%s'", tp.String())
	}

	sntp, ok := ntp.Underlying().(*types.Struct)
	if !ok {
		return nil, fmt.Errorf("dst obj NOT struct '%s'", ntp.String())
	}

	return sntp, nil
}

func (b *assignmentBuilder) buildPostAssignmentImp(
	lhsStruct, rhsStruct bmodel.Node, retError bool,
) (*strings.Builder, error) {
	sb := &strings.Builder{}

	sntp, err := b.getRhsStruct(rhsStruct)
	if err != nil {
		return nil, err
	}

	tmpl, err := template.New("postAssignment").Funcs(template.FuncMap{
		"add": func(a, b int) int {
			return a + b
		},
		"title": func(s string) string {
			// 注意：strings.Title 已弃用（Go 1.18+），不推荐使用！
			return strings.ToUpper(string(s[0])) + strings.ToLower(s[1:])
		},
	}).Parse(structMemberTmpl)
	if err != nil {
		return nil, err
	}

	srcMaskMap := map[string]*option.MaskConverter{}
	dstMaskMap := map[string]*option.MaskConverter{}

	for _, m := range b.opts.BuildMaskConverters {
		srcMaskMap[m.Src().NameAt(0)] = m
		dstMaskMap[m.Dst().NameAt(0)] = m
	}

	srcMaskList := []struct {
		SrcFlagName string // 原始对象中的一个个bit位对应的bool值名称
		DstMaskName string // 目标对象中的PropertyMask的名称
		MaskName    string // 具体的bit定义常量名称， 例如 PropertyMaskSomeDef
		Type        string
	}{}
	for _, v := range srcMaskMap {
		srcMaskList = append(srcMaskList, struct {
			SrcFlagName string
			DstMaskName string
			MaskName    string
			Type        string
		}{
			SrcFlagName: v.Dst().NameAt(0),
			DstMaskName: v.Src().NameAt(0),
			MaskName:    v.Mask(),
			Type:        v.GetMaskBasic().String(),
		})
	}
	// 所有的都要排序， 确保生成代码的顺序一致性， 下同
	sort.Slice(srcMaskList, func(i, j int) bool {
		return srcMaskList[i].DstMaskName < srcMaskList[j].DstMaskName
	})

	// 所有的 mask （不是flag）
	dstMaskList := []struct {
		Mask     string // mask字段名称， 例如 PropertyMask
		Type     string // mask字段类型，例如 int64
		FlagType string // mask字段bit位的typedef 名称， 例如 type `PropertyMask` int64
	}{}
	for k, v := range dstMaskMap {
		dstMaskList = append(dstMaskList, struct {
			Mask     string
			Type     string
			FlagType string
		}{
			Mask:     k,
			Type:     v.GetMaskBasic().String(),
			FlagType: b.imports.TypeName(v.GetMaskConst().Type()),
		})
	}
	sort.Slice(dstMaskList, func(i, j int) bool {
		return dstMaskList[i].Mask < dstMaskList[j].Mask
	})

	maskExtension, err := checkExternsion(&b.opts)
	if err != nil {
		return nil, err
	}

	// 其他非mask字段
	dstOtherFields := []string{}
	for i := sntp.NumFields() - 1; i >= 0; i-- {
		f := sntp.Field(i).Name()
		if _, ok := dstMaskMap[f]; ok {
			continue
		}

		if b.opts.Mask != nil && len(b.opts.Mask.SkipFields) > 0 {
			if _, ok := b.opts.Mask.SkipFields[f]; ok {
				continue
			}
		}

		dstOtherFields = append(dstOtherFields, f)
	}
	sort.Slice(dstOtherFields, func(i, j int) bool {
		return dstOtherFields[i] < dstOtherFields[j]
	})

	params := map[string]any{
		"FuncName":            b.funcName,
		"Receiver":            b.imports.TypeName(rhsStruct.ExprType()),
		"SrcStruct":           b.imports.TypeName(lhsStruct.ExprType()),
		"SrcMaskList":         srcMaskList,
		"DstMaskList":         dstMaskList,
		"DstOtherFields":      dstOtherFields,
		"EnableMask":          b.opts.Mask != nil,
		"EnableMaskExtension": b.opts.MaskExtension != nil,
		"MaskExtension":       maskExtension,
		"RetError":            retError,
	}
	if err = tmpl.Execute(sb, params); err != nil {
		return nil, err
	}

	return sb, nil
}

func (b *assignmentBuilder) buildPostAssignment(
	lhsStruct, rhsStruct bmodel.Node, retError bool,
) (gmodel.Assignment, error) {
	if len(b.opts.BuildMaskConverters) < 1 {
		return nil, nil
	} else if b.opts.Receiver != "" {
		return nil, nil
	}

	// 对象必须是本包
	if b.imports.IsExternal(lhsStruct.ExprType()) {
		return nil, nil
	}

	sb, err := b.buildPostAssignmentImp(rhsStruct, lhsStruct, retError)
	if err != nil {
		return nil, err
	}

	// fmt.Println(sb.String())
	return gmodel.RawAssignment{Err: false, Raw: sb.String()}, nil
}
