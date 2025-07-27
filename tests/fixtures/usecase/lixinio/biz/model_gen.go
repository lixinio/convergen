package biz

type unexportedBizPropertyMaskQuery struct {
	BizID             string
	BizPropertyMaskA1 string
	BizPropertyMaskB2 string
	BizPropertyMaskC3 string
	BizMaskA1         string
	BizMaskB2         string
	BizMaskC3         string
}

var BizPropertyMaskQuery = unexportedBizPropertyMaskQuery{
	BizID:             "biz_id",
	BizPropertyMaskA1: "biz_property_mask_a1",
	BizPropertyMaskB2: "biz_property_mask_b2",
	BizPropertyMaskC3: "biz_property_mask_c3",
	BizMaskA1:         "biz_mask_a1",
	BizMaskB2:         "biz_mask_b2",
	BizMaskC3:         "biz_mask_c3",
}
