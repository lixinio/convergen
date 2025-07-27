package biz

import "time"

type ClentStatus int32

type StructField struct {
	Test int32
}

type Client struct {
	ID           int64
	Status       ClentStatus
	Status2      ClentStatus
	StatusPtr    *ClentStatus
	StructPtr    *StructField
	StringPtr    *string
	IntPtr       *int32
	ClientID     string
	ClientSecret string
	TokenExpire  int
	CreateTime   time.Time
	UpdateTime   time.Time
	Provider     *ClientProvider
	Provider3    *ClientProvider
	Uris         []*ClientRedirectUri
	StringSlice  []string
	IntSlice2    []int
}

type ClientRedirectUri struct {
	ID       int64
	ClientID int64
	Uri      string
}

type ClientProvider struct {
	ID       int64
	ClientID int64
	Uri      string
}

type Xint int32

type BizPropertyMask struct {
	BizID             Xint
	BizPropertyMaskA1 bool
	BizPropertyMaskB2 bool
	BizPropertyMaskC3 bool
	BizMaskA1         bool
	BizMaskB2         bool
	BizMaskC3         bool
}

type (
	PropertyMask int64
	Mask         uint64
	Str          string
)

const (
	PropertyMaskA PropertyMask = 0b1 << iota
	PropertyMaskB
	PropertyMaskC
)

const (
	MaskA Mask = 0b1 << iota
	MaskB
	MaskC
	MaskD
)
