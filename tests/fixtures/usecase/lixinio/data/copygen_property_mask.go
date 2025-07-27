//go:build convergen

/* Specify the name of the generated file's package. */
package data

//go:generate /tmp/convergen/convergen -suffix transfer

import (
	_ "strings"

	biz_alias "github.com/reedom/convergen/tests/fixtures/usecase/lixinio/biz"
	"go.lixinio.com/apis/pkg/msku"
)

type _ msku.Mask

func convID(a int) (int, error) {
	return a, nil
}

// :convergen
// :typecast
type ConvergenPropertyMask interface {
	// :recv client DbPropertyMask
	// :map DbID BizID
	// :parsemask DbPropertyMask BizPropertyMaskA1 biz_alias.PropertyMaskA
	// :parsemask DbPropertyMask BizPropertyMaskB2 biz_alias.PropertyMaskB
	// :parsemask DbPropertyMask BizPropertyMaskC3 biz_alias.PropertyMaskC
	// :parsemask DbMask BizMaskA1 biz_alias.MaskA
	// :parsemask DbMask BizMaskB2 biz_alias.MaskB
	// :parsemask DbMask BizMaskC3 biz_alias.MaskC
	DbPropertyMaskToBiz(*DbPropertyMask) *biz_alias.BizPropertyMask

	// :conv convID BizID DbID
	// :mask:ext biz_alias.BizPropertyMaskQuery
	// :mask biz_alias.BizPropertyMaskQuery DbID1
	// :buildmask DbPropertyMask BizPropertyMaskA1 biz_alias.PropertyMaskA
	// :buildmask DbPropertyMask BizPropertyMaskB2 biz_alias.PropertyMaskB
	// :buildmask DbPropertyMask BizPropertyMaskC3 biz_alias.PropertyMaskC
	// :buildmask DbMask BizMaskA1 biz_alias.MaskA
	// :buildmask DbMask BizMaskB2 biz_alias.MaskB
	// :buildmask DbMask BizMaskC3 biz_alias.MaskC
	// :buildmask DbMask - biz_alias.MaskD
	DbPropertyMaskFromBiz(*biz_alias.BizPropertyMask) (*DbPropertyMask, error)
}
