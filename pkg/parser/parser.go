package parser

import (
	"bytes"
	"errors"
	"fmt"
	"go/ast"
	"go/constant"
	"go/parser"
	"go/printer"
	"go/token"
	"go/types"
	"os"
	"regexp"

	"github.com/reedom/convergen/pkg/builder"
	"github.com/reedom/convergen/pkg/builder/model"
	"github.com/reedom/convergen/pkg/config"
	"github.com/reedom/convergen/pkg/logger"
	"github.com/reedom/convergen/pkg/option"
	"github.com/reedom/convergen/pkg/util"
	"golang.org/x/tools/go/packages"
)

const buildTag = "convergen"

// Parser represents a parser for a Go source file that contains convergen blocks.
type Parser struct {
	srcPath     string            // The path to the source file being parsed.
	file        *ast.File         // The parsed AST of the source file.
	fset        *token.FileSet    // The token file set used for parsing.
	pkg         *packages.Package // The package information for the parsed file.
	opts        option.Options    // The options for the parser.
	imports     util.ImportNames  // The import names used in the parsed file.
	intfEntries []*intfEntry      // The interface entries parsed from the file.
}

// parserLoadMode is a packages.Load mode that loads types and syntax trees.
const parserLoadMode = packages.NeedName | packages.NeedImports | packages.NeedDeps |
	packages.NeedTypes | packages.NeedSyntax | packages.NeedTypesInfo

// NewParser returns a new parser for convergen annotations.
func NewParser(conf *config.Config) (*Parser, error) {
	var (
		fileSet = token.NewFileSet()
		fileSrc *ast.File
		srcPath = conf.Input
		dstPath = conf.Output
		opts    = option.NewOptions()
	)

	opts.Getter = conf.Getter
	opts.ExactCase = conf.ExactCase
	opts.Stringer = conf.Stringer
	opts.Typecast = conf.Typecast

	srcStat, err := os.Stat(srcPath)
	if err != nil {
		return nil, err
	}

	dstStat, _ := os.Stat(dstPath)
	var parseErr error
	cfg := &packages.Config{
		Mode:       parserLoadMode,
		BuildFlags: []string{"-tags", buildTag},
		Fset:       fileSet,
		ParseFile: func(fset *token.FileSet, filename string, src []byte) (*ast.File, error) {
			stat, err := os.Stat(filename)
			if err != nil {
				return nil, err
			}

			// If previously generation target file exists, skip reading it.
			if os.SameFile(stat, dstStat) {
				return nil, nil
			}

			// 如果不是源文件， 简单parse就好了， 源文件需要特殊处理
			if !os.SameFile(stat, srcStat) {
				return parser.ParseFile(fset, filename, src, 0)
			}

			file, err := parser.ParseFile(fset, filename, src, parser.ParseComments)
			if err != nil {
				parseErr = err
				return nil, err
			}
			fileSrc = file
			return file, nil
		},
	}
	pkgs, err := packages.Load(cfg, "file="+srcPath)
	if err != nil {
		return nil, logger.Errorf("%v: failed to load type information: \n%w", srcPath, err)
	}
	if len(pkgs) == 0 {
		return nil, logger.Errorf("%v: failed to load package information", srcPath)
	}

	if fileSrc == nil && parseErr != nil {
		return nil, logger.Errorf("%v: %v", srcPath, parseErr)
	}
	return &Parser{
		// 获得完整的绝对路径
		srcPath: fileSet.Position(fileSrc.Pos()).Filename,
		fset:    fileSet,
		file:    fileSrc,
		pkg:     pkgs[0],
		opts:    opts,
		imports: util.NewImportNames(fileSrc.Imports),
	}, nil
}

// Parse parses convergen annotations in the source code.
func (p *Parser) Parse() ([]*model.MethodsInfo, error) {
	// 获得所有的interface
	entries, err := p.findConvergenEntries()
	if err != nil {
		return nil, err
	}

	var allMethods []*model.MethodEntry

	var list []*model.MethodsInfo
	for _, entry := range entries {
		// 解析单个interface
		methods, err := p.parseMethods(entry)
		if err != nil {
			return nil, err
		}
		info := &model.MethodsInfo{
			Marker:  entry.marker,
			Methods: methods,
		}
		list = append(list, info)
		allMethods = append(allMethods, methods...)
	}

	// Resolve converters.
	// Some converters may refer to-be-generated functions that go/types doesn't contain
	// so that they are needed to be resolved manually.
	for _, method := range allMethods {
		for _, conv := range method.Opts.Converters {
			err = p.resolveConverters(allMethods, conv)
			if err != nil {
				return nil, err
			}
		}

		if err := p.resolveMaskConverter(method); err != nil {
			return nil, err
		}
	}

	p.intfEntries = entries
	return list, nil
}

func (p *Parser) resolveMaskConverter(method *model.MethodEntry) error {
	if len(method.Opts.ParseMaskConverters) > 0 {
		for _, conv := range method.Opts.ParseMaskConverters {
			_, err := p.resolveMasks(conv)
			if err != nil {
				return err
			}
		}

		/*
			检查 flag 相同， 但是bit的（类型）不一致
			type FlagBitA int
			type FlagBitB int
			const FlagBit1 FlagBitA
			const FlagBit2 FlagBitB
			:buildmask PropertyMask FieldFlag1 FlagBit1
			:buildmask PropertyMask FieldFlag2 FlagBit2

			上面都是 PropertyMask， 但是 FlagBit1 FlagBit2 却是不同类型
		*/
		if err := checkGetMaskTheSame(method.Opts.ParseMaskConverters, true); err != nil {
			return err
		}
	}

	if len(method.Opts.BuildMaskConverters) > 0 {
		buildMaskIngores := []*option.MaskConverter{}
		newBuildMaskConverters := []*option.MaskConverter{}
		pkgs := map[*packages.Package][]*types.Const{}
		for _, conv := range method.Opts.BuildMaskConverters {
			pkg, err := p.resolveMasks(conv)
			if err != nil {
				return err
			}

			// 当从多个flag到一个propertyMask时， 如果增加一个flagBit 和 bool字段
			// 漏了加转换是发现不了的， 所以这里加一个额外判断
			v, ok := pkgs[pkg]
			if ok {
				v = append(v, conv.GetMaskConst())
			} else {
				v = []*types.Const{conv.GetMaskConst()}
			}

			pkgs[pkg] = v

			if !conv.Src().Match("-", true) {
				newBuildMaskConverters = append(newBuildMaskConverters, conv)
			} else {
				buildMaskIngores = append(buildMaskIngores, conv)
			}
		}

		if err := checkGetMaskTheSame(method.Opts.BuildMaskConverters, false); err != nil {
			return err
		}

		if err := p.checkBuildMaskMissingField(pkgs); err != nil {
			return err
		}

		// 排除 ‘-’
		// :buildmask DbMask BizMaskA1 biz_alias.MaskA
		// :buildmask DbMask BizMaskB2 biz_alias.MaskB
		// :buildmask DbMask - biz_alias.MaskC
		method.Opts.BuildMaskConverters = newBuildMaskConverters
		method.Opts.BuildMaskIgnores = buildMaskIngores
	}

	if method.Opts.MaskExtension != nil || method.Opts.Mask != nil {
		var maskExtension *option.MaskExtension
		if method.Opts.MaskExtension != nil {
			maskExtension = method.Opts.MaskExtension
		} else {
			maskExtension = method.Opts.Mask.MaskExtension
		}

		if v, s, err := p.lookupStructVarible(maskExtension.Name, method.Method.Pos()); err != nil {
			return err
		} else {
			maskExtension.Var = v
			maskExtension.Struct = s
		}

		// if method.Opts.MaskExtension != nil {
		if obj, _, _ := p.lookupType("msku.Mask", method.Method.Pos()); obj == nil {
			return errors.New(
				"MaskExtension need pkg 'go.lixinio.com/apis/pkg/msku' with 'type _ msku.Mask'",
			)
		}
	}

	return nil
}

func (p *Parser) checkBuildMaskMissingField(
	pkgs map[*packages.Package][]*types.Const,
) error {
	for pkg, cs := range pkgs {
		// 将所有的const 按照 type聚合
		csMap := map[string][]*types.Const{}
		for _, c := range cs {
			tp := c.Type().String()

			v, ok := csMap[tp]
			if ok {
				v = append(v, c)
			} else {
				v = []*types.Const{c}
			}

			csMap[tp] = v
		}

		// 遍历每个Type
		for tp, cs := range csMap {
			// 获得该type的所有值
			vs, err := p.EnumConstValues(pkg, cs[0])
			if err != nil {
				return err
			}

			if len(vs) != len(cs) {
				// 已经定义的转成map
				defined := map[string]int{}
				for _, c := range cs {
					defined[c.Name()] = 1
				}

				// 找出具体哪个，方便debug
				for _, v := range vs {
					if _, ok := defined[v]; !ok {
						return fmt.Errorf("maskBit '%s' value '%s' matched flag NOT exist", tp, v)
					}
				}

				return fmt.Errorf("maskBit '%s' missing", tp)
			}
		}
	}

	return nil
}

// EnumConstValues 枚举与 targetConst 同类型的所有常量及其值
func (p *Parser) EnumConstValues(
	pkg *packages.Package, targetConst *types.Const,
) ([]string, error) {
	// 1. 获取目标常量的类型
	targetType := targetConst.Type()
	if targetType == nil {
		return nil, fmt.Errorf("target const has no type")
	}

	// 2. 遍历包中所有定义的对象，筛选同类型常量
	result := []string{}

	for _, obj := range pkg.TypesInfo.Defs {
		if obj == nil {
			continue // 跳过未定义的标识符
		}

		// 检查是否为常量
		c, ok := obj.(*types.Const)
		if !ok {
			continue
		}

		// 检查类型是否与目标常量一致
		if !types.Identical(c.Type(), targetType) {
			continue
		}

		// 3. 提取常量值（假设为 int 类型，其他类型需适配）
		val := c.Val()
		if val.Kind() != constant.Int {
			return nil, fmt.Errorf("const %s is not an integer", c.Name())
		}

		result = append(result, c.Name())
	}

	return result, nil
}

// 对于同一个PropertyMask字段， Mask应该也是一样的 且不能重复
func checkGetMaskTheSame(
	converters []*option.MaskConverter, readFromMask bool,
) error {
	if len(converters) == 1 {
		return nil
	}

	maskKeyGetter := func(c *option.MaskConverter) string {
		if readFromMask {
			return c.Src().String()
		} else {
			return c.Dst().String()
		}
	}

	maskFlagGetter := func(c *option.MaskConverter) string {
		if !readFromMask {
			return c.Src().String()
		} else {
			return c.Dst().String()
		}
	}

	// 按照 PropertyMask 聚合
	css := map[string][]*option.MaskConverter{}
	for _, c := range converters {
		maskkey := maskKeyGetter(c)
		v, ok := css[maskkey]
		if ok {
			v = append(v, c)
		} else {
			v = []*option.MaskConverter{c}
		}

		css[maskkey] = v
	}

	for k, cs := range css {
		if len(cs) <= 1 { // 只有一个无需比较
			continue
		}

		// MaskBit type必须一致
		tpName := cs[0].GetMaskConst().Type().String()
		for i := 1; i < len(cs); i++ {
			tp := cs[i].GetMaskConst().Type().String()
			if tp != tpName {
				return fmt.Errorf(
					"field '%s' have diffrent mask type '%s', '%s'",
					k, tpName, tp,
				)
			}
		}

		// MaskBit value 不能重复
		valueNames := map[string]int{cs[0].Mask(): 0}
		for i := 1; i < len(cs); i++ {
			maskBit := cs[i].Mask()
			if _, ok := valueNames[maskBit]; ok {
				return fmt.Errorf(
					"maskBit '%s' duplicated, type '%s'", maskBit, tpName,
				)
			}

			valueNames[maskBit] = 0
		}

		// MaskFlag 不能重复
		valueNames = map[string]int{maskFlagGetter(cs[0]): 0}
		for i := 1; i < len(cs); i++ {
			maskFlag := maskFlagGetter(cs[i])
			if _, ok := valueNames[maskFlag]; ok {
				return fmt.Errorf("maskFlag '%s' duplicated", maskFlag)
			}

			valueNames[maskFlag] = 0
		}
	}

	return nil
}

func (p *Parser) resolveMasks(conv *option.MaskConverter) (*packages.Package, error) {
	name := conv.Mask()
	pos := conv.Pos()
	c, u, cp, err := p.lookupMaskConst(name, pos)
	if err != nil {
		return nil, err
	}

	conv.Set(c, u)
	return cp, nil
}

func (p *Parser) lookupMaskConst(
	constName string, pos token.Pos,
) (
	// if `type PropertyMask int64`
	c *types.Const, // = PropertyMask
	u *types.Basic, // = int64
	cp *packages.Package,
	err error,
) {
	var obj types.Object
	posStr := p.fset.Position(pos)
	_, obj, cp = p.lookupType(constName, pos)
	if obj == nil {
		err = fmt.Errorf("%v: const %v not found", posStr, constName)
		return
	}

	maskField, ok := obj.(*types.Const)
	if !ok {
		err = fmt.Errorf("%v: %v isn't a const", posStr, constName)
		return
	}

	tp := maskField.Type()
	maskOriginType := tp.String() // typedef 别名

	// mask的定义必须是type xxx integer
	underlying, ok := tp.Underlying().(*types.Basic)
	if !ok {
		err = fmt.Errorf(
			"%v: not 'intunderlying basic' type for '%v', auctual type '%s'",
			posStr, constName, maskOriginType,
		)

		return
	}

	// 例如 type PropertyMask uint64
	//  maskOriginType = PropertyMask
	// 	underlyingType = uint64
	underlyingType := underlying.String()
	if underlyingType == maskOriginType { // mask的定义必须是type xxx integer
		err = fmt.Errorf(
			"%v: mask type must be 'typedef' type for '%v', auctual type '%s'",
			posStr, constName, maskOriginType,
		)
		return
	}

	// 源字段必须是integer类型, 不支持int、uint, 用int32, uint32代替
	if underlyingType != "int64" && underlyingType != "uint64" &&
		underlyingType != "int32" && underlyingType != "uint32" &&
		underlyingType != "int16" && underlyingType != "uint16" &&
		underlyingType != "int8" && underlyingType != "uint8" {
		err = fmt.Errorf(
			"%v: not 'interger' type for %v, auctual type '%s'/'%s'",
			posStr, constName, maskOriginType, underlyingType,
		)

		return
	}

	return maskField, underlying, cp, nil
}

func (p *Parser) lookupStructVarible(
	constName string, pos token.Pos,
) (
	c *types.Var,
	s *types.Struct,
	err error,
) {
	posStr := p.fset.Position(pos)
	_, obj, _ := p.lookupType(constName, pos)
	if obj == nil {
		err = fmt.Errorf("%v: const %v not found", posStr, constName)
		return
	}

	v, ok := obj.(*types.Var)
	if !ok {
		err = fmt.Errorf("%v: %v isn't a varible", posStr, constName)
		return
	}

	vt := v.Type()
	vtn, ok := vt.(*types.Named)
	if !ok {
		err = fmt.Errorf("%v: %v isn't a Named", posStr, constName)
		return
	}

	vtns, ok := vtn.Underlying().(*types.Struct)
	if !ok {
		err = fmt.Errorf("%v: %v isn't a Struct", posStr, constName)
		return
	}

	return v, vtns, nil
}

// CreateBuilder creates a new function builder.
func (p *Parser) CreateBuilder() *builder.FunctionBuilder {
	return builder.NewFunctionBuilder(p.file, p.fset, p.pkg, p.imports)
}

// GenerateBaseCode generates the base code without convergen annotations.
// The code is stripped of convergen annotations and the doc comments of interfaces.
// The resulting code can be used as a starting point for the code generation process.
// GenerateBaseCode returns the resulting code as a string, or an error if the generation process fails.
func (p *Parser) GenerateBaseCode() (code string, err error) {
	util.RemoveMatchComments(p.file, reGoBuildGen)

	// Remove doc comment of the interface.
	// And also find the range pos of the interface in the code.
	for _, entry := range p.intfEntries {
		nodes, _ := util.ToAstNode(p.file, entry.intf)
		var minPos, maxPos token.Pos

		for _, node := range nodes {
			switch n := node.(type) {
			case *ast.GenDecl:
				ast.Inspect(n, func(node ast.Node) bool {
					if node == nil {
						return true
					}
					if f, ok := node.(*ast.FieldList); ok {
						if minPos == 0 {
							minPos = f.Pos()
							maxPos = f.Closing
						} else if f.Pos() < minPos {
							minPos = f.Pos()
						} else if maxPos < f.Closing {
							maxPos = f.Closing
						}
					}
					return true
				})
			}
		}

		// fix interface内容短于marker的时候
		// 下面的代码会报错
		/*
			type Convergen interface {
				A2B(*A) *B
			}
		*/
		// Insert markers.
		util.InsertComment(p.file, entry.marker, maxPos)
		util.InsertComment(p.file, entry.marker, minPos)
	}

	var buf bytes.Buffer
	err = printer.Fprint(&buf, p.fset, p.file)
	if err != nil {
		return
	}

	base := buf.String()
	// Now each interfaces is marked with two <<marker>>s like below:
	//
	//	    type Convergen <<marker>>interface {
	//	      DomainToModel(pet *mx.Pet) *mx.Pet
	//      }   <<marker>>
	//
	// And then we're going to convert it to:
	//
	//	    <<marker>>

	for _, entry := range p.intfEntries {
		reMarker := regexp.QuoteMeta(entry.marker)
		re := regexp.MustCompile(`.+` + reMarker + ".*(\n|.)*?" + reMarker)
		base = re.ReplaceAllString(base, entry.marker)
	}

	return base, nil
}
