// Code generated by github.com/reedom/convergen
// DO NOT EDIT.

package data

import (
	"errors"
	"strings"
	_ "strings"

	"github.com/reedom/convergen/tests/fixtures/usecase/lixinio/biz"
)

func Int8(status biz.ClentStatus) int8 {
	return int8(status)
}

func ClientStatus(status int8) biz.ClentStatus {
	return biz.ClentStatus(status)
}

func prepareClientProvider(dst *ClientProvider, src *biz.ClientProvider) error {
	if src == nil {
		return errors.New("empty model")
	}

	return nil
}

func cleanUpClientProvider(dst *ClientProvider, src *biz.ClientProvider) error {
	if dst.ID == 0 {
		return errors.New("empty model id")
	}

	return nil
}

// 忽略特定字段
// 转换
// 使用成员函数, 函数名去掉前缀(ClientProvider, 就保留 ToBiz)
func (client *ClientProvider) ToBiz() (dst *biz.ClientProvider) {
	if client == nil {
		return
	}

	dst = &biz.ClientProvider{}
	dst.ID = client.ID
	// skip: dst.ClientID
	dst.Uri = strings.ToLower(client.Uri)

	return
}

// 映射并转换
// 使用成员函数, 函数名去掉前缀(ClientRedirectUri, 就保留 ToBiz)
func (client *ClientRedirectUri) ToBiz() (dst *biz.ClientRedirectUri) {
	if client == nil {
		return
	}

	dst = &biz.ClientRedirectUri{}
	dst.ID = client.ID
	dst.ClientID = client.ClientID
	dst.Uri = strings.ToUpper(client.Url)

	return
}

// 使用成员函数, 函数名去掉前缀(Client, 就保留 ToBiz)
// 忽略字段
// 也可以这样 :conv ClientStatus Status
// 字段映射
// 调用成员函数
func (client *Client) ToBiz() (dst *biz.Client) {
	if client == nil {
		return
	}

	dst = &biz.Client{}
	dst.ID = client.ID
	dst.Status = biz.ClentStatus(client.Status)
	dst.ClientID = client.ClientID
	// skip: dst.ClientSecret
	dst.TokenExpire = client.TokenExpire
	dst.CreateTime = client.CreateAt
	dst.UpdateTime = client.UpdateTime
	if client.Provider != nil {
		dst.Provider = client.Provider.ToBiz()
	}
	if client.Provider2 != nil {
		dst.Provider3 = client.Provider2.ToBiz()
	}
	if client.Uris != nil {
		dst.Uris = make([]*biz.ClientRedirectUri, len(client.Uris))
		for i, e := range client.Uris {
			if e != nil {
				dst.Uris[i] = e.ToBiz()
			}
		}
	}
	if client.StringSlice != nil {
		dst.StringSlice = make([]string, len(client.StringSlice))
		copy(dst.StringSlice, client.StringSlice)
	}
	dst.IntSlice2 = client.IntSlice

	return
}

// 忽略字段
// 用自定义函数转换
// 字段映射
// 转换
func NewClientFromBiz(src *biz.Client) (dst *Client, err error) {
	if src == nil {
		return
	}

	dst = &Client{}
	dst.ID = src.ID
	dst.Status = Int8(src.Status)
	dst.ClientID = src.ClientID
	// skip: dst.ClientSecret
	dst.TokenExpire = src.TokenExpire
	dst.CreateAt = src.CreateTime
	dst.UpdateTime = src.UpdateTime.Local()
	dst.Provider, err = NewClientProviderFromBiz(src.Provider)
	if err != nil {
		return nil, err
	}
	dst.Provider2, err = NewClientProviderFromBiz(src.Provider3)
	if err != nil {
		return nil, err
	}
	if src.Uris != nil {
		dst.Uris = make([]*ClientRedirectUri, len(src.Uris))
		for i, e := range src.Uris {
			dst.Uris[i] = NewClientRedirectUriFromBiz(e)
		}
	}
	if src.StringSlice != nil {
		dst.StringSlice = make([]string, len(src.StringSlice))
		copy(dst.StringSlice, src.StringSlice)
	}
	dst.IntSlice = src.IntSlice2

	return
}

// 转换
// 前置/后置检查
func NewClientProviderFromBiz(src *biz.ClientProvider) (dst *ClientProvider, err error) {
	if src == nil {
		return
	}

	dst = &ClientProvider{}
	err = prepareClientProvider(dst, src)
	if err != nil {
		return
	}
	dst.ID = src.ID
	dst.ClientID = src.ClientID
	dst.Uri = strings.ToLower(src.Uri)
	dst.InternalFlag = 123
	err = cleanUpClientProvider(dst, src)
	if err != nil {
		return
	}

	return
}

// 映射并转换
func NewClientRedirectUriFromBiz(src *biz.ClientRedirectUri) (dst *ClientRedirectUri) {
	if src == nil {
		return
	}

	dst = &ClientRedirectUri{}
	dst.ID = src.ID
	dst.ClientID = src.ClientID
	dst.Url = strings.ToUpper(src.Uri)

	return
}