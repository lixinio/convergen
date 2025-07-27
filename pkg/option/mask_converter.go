package option

import (
	"go/token"
	"go/types"
)

type MaskExtension struct {
	Name   string
	Var    *types.Var
	Struct *types.Struct
}

type Mask struct {
	*MaskExtension
	SkipFields map[string]int
}

type MaskConverter struct {
	m          *NameMatcher // A name matcher that matches the name of the source and destination fields.
	mask       string       // The name of the mask.
	maskType   *types.Const
	underlying *types.Basic
}

func NewMaskConverter(mask, src, dst string, pos token.Pos) *MaskConverter {
	return &MaskConverter{
		m:    NewNameMatcher(src, dst, pos),
		mask: mask,
	}
}

func (c *MaskConverter) Match(src, dst string) bool {
	return c.m.Match(src, dst, true)
}

// Src returns the FieldConverter's source identifier matcher.
func (c *MaskConverter) Src() *IdentMatcher {
	return c.m.src
}

// Dst returns the FieldConverter's destination identifier matcher.
func (c *MaskConverter) Dst() *IdentMatcher {
	return c.m.dst
}

// Pos returns the position of the FieldConverter.
func (c *MaskConverter) Pos() token.Pos {
	return c.m.pos
}

func (c *MaskConverter) Mask() string {
	return c.mask
}

func (c *MaskConverter) GetMaskConst() *types.Const {
	return c.maskType
}

func (c *MaskConverter) GetMaskBasic() *types.Basic {
	return c.underlying
}

func (c *MaskConverter) Set(
	maskType *types.Const, underlying *types.Basic,
) {
	c.maskType = maskType
	c.underlying = underlying
}
