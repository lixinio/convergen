本文档描述多个**bool**标志位 到 **PropertyMask** 的相互转换

# 1. 核心转换
## 1.2. 术语：
1. **PropertyMask** : 用来存储多个bit的字段
2. **MaskBit**: 一个常量， 表示某个位， 通常等于 ( 1 << x )
3. **MastFlag**: 对应特定MaskBit的 bool 值

## 1.2. 前置要求：
1. 标志位必须是 `bool` 类型
2. PropertyMask 类型必须是 **int64、uint64、int32、uint32、int16、uint16、int8、uint8**
   + **int、uint** 不支持， 用 **int32、uint32** 代替

每个 **bool** 对应 **PropertyMask** 的某一位， 这个位的定义， 必须基于`同类型`的 **typedef**

``` go
// MaskBit 类型
type (
	PropertyMaskBit int8   // 必须和 Model.PropertyMask 类型一致
	OtherMaskBit    uint32 // 必须和 Model.OtherMask 类型一致
)

// MaskBit 值
const (
	PropertyMaskBit1 PropertyMaskBit = 1 << iota
	PropertyMaskBit2
	PropertyMaskBit3
)
const (
	OtherMaskBit1 OtherMaskBit = 1 << iota
	OtherMaskBit2
	OtherMaskBit3
)

type Model struct {
	PropertyMask int8
	OtherMask    uint32
}

type BizModel struct {
	PropertyMaskFlag1 bool
	PropertyMaskFlag2 bool
	PropertyMaskFlag3 bool
	OtherMaskFlag1    bool
	OtherMaskFlag2    bool
	OtherMaskFlag3    bool
}
```

## 1.3. 注释语法
``` go
type Convergen interface {
	// :parsemask PropertyMask PropertyMaskFlag1 PropertyMaskBit1
	// :parsemask PropertyMask PropertyMaskFlag2 PropertyMaskBit2
	// :parsemask PropertyMask PropertyMaskFlag3 PropertyMaskBit3
	// :parsemask OtherMask OtherMaskFlag1 OtherMaskBit1
	// :parsemask OtherMask OtherMaskFlag2 OtherMaskBit2
	// :parsemask OtherMask OtherMaskFlag3 OtherMaskBit3
	ModelToBiz(*Model) *BizModel
	// :buildmask PropertyMask PropertyMaskFlag1 PropertyMaskBit1
	// :buildmask PropertyMask PropertyMaskFlag2 PropertyMaskBit2
	// :buildmask PropertyMask PropertyMaskFlag3 PropertyMaskBit3
	// :buildmask OtherMask OtherMaskFlag1 OtherMaskBit1
	// :buildmask OtherMask OtherMaskFlag2 OtherMaskBit2
	// :buildmask OtherMask OtherMaskFlag3 OtherMaskBit3
	NewModelFromBiz(*BizModel) *Model
}
```

注意：
1. 为了使用方便， 上下两段注释， 除了 **parsemask** 和 **buildmask** 的差异， 其他都是一样的， 这样方便`直接拷贝`后再简单修改
2. 同一个 **PropertyMask**， **MaskBit**必须必须类型相同， 以 `// :buildmask PropertyMask PropertyMaskFlag1 PropertyMaskBit1` 为例， 所有`// :buildmask PropertyMask`开头的注释第三个参数必须都是**相同类型**的 `PropertyMaskBit`, 错误类似 `field 'IdentMatcher{pattern: "OtherMask"}' have diffrent mask type 'OtherMaskBit', 'sample.PropertyMaskBit'`
3. MaskBit 必须存在， 错误类似 `convergen.go:9:2: const OtherMaskBit3 not found`
4. MaskBit 不能重复， 错误类似  `maskBit 'OtherMaskBit2' duplicated, type 'github.com/reedom/convergen/tests/sample.OtherMaskBit'`
5. MaskFlag 必须存在， 错误类似 `maskBit 'sample.OtherMaskBit' value 'OtherMaskBit4' matched flag NOT exist`
6. MaskFlag 不能重复
   1. 对于 **:parsemask**, 因为 MaskBit 不能重复， 所以一旦 MastFlag 重复， 意味着有一些 MastFlag匹配不上， 所有会生成**no match**, 类似 `// no match: dst.PropertyMaskFlag2`
   2. 对于 **:buildmask**, 生成的代码会有编译错误
7. PropertyMask 必须存在, 错误类似 `maskFlag 'IdentMatcher{pattern: "BizPropertyMaskA1"}' duplicated`

## 1.3. 生成代码
会对**每个Model**的**每个Mask** 生成 Set 和 Get 方法， 例如
``` go
func (dst *Model) SetOtherMask(flag bool, mask OtherMaskBit) {
	if flag {
		dst.OtherMask |= uint32(mask)
	} else {
		dst.OtherMask &= uint32(^mask)
	}
}

func (dst *Model) GetOtherMask(mask OtherMaskBit) bool {
	return (OtherMaskBit(dst.OtherMask) & mask) != 0
}

func (dst *Model) SetPropertyMask(flag bool, mask PropertyMaskBit) {
	if flag {
		dst.PropertyMask |= int8(mask)
	} else {
		dst.PropertyMask &= int8(^mask)
	}
}

func (dst *Model) GetPropertyMask(mask PropertyMaskBit) bool {
	return (PropertyMaskBit(dst.PropertyMask) & mask) != 0
}
```

## 1.4. 转换函数
``` go
func ModelToBiz(src *Model) (dst *BizModel) {
	if src == nil {
		return
	}

	dst = &BizModel{}
	dst.PropertyMaskFlag1 = src.GetPropertyMask(PropertyMaskBit1)
	dst.PropertyMaskFlag2 = src.GetPropertyMask(PropertyMaskBit2)
	dst.PropertyMaskFlag3 = src.GetPropertyMask(PropertyMaskBit3)
	dst.OtherMaskFlag1 = src.GetOtherMask(OtherMaskBit1)
	dst.OtherMaskFlag2 = src.GetOtherMask(OtherMaskBit2)
	dst.OtherMaskFlag3 = src.GetOtherMask(OtherMaskBit3)

	return
}
```
``` go
func NewModelFromBiz(src *BizModel) (dst *Model) {
	if src == nil {
		return
	}

	dst = &Model{}
	dst.SetPropertyMask(src.PropertyMaskFlag1, PropertyMaskBit1)
	dst.SetPropertyMask(src.PropertyMaskFlag2, PropertyMaskBit2)
	dst.SetPropertyMask(src.PropertyMaskFlag3, PropertyMaskBit3)
	dst.SetOtherMask(src.OtherMaskFlag1, OtherMaskBit1)
	dst.SetOtherMask(src.OtherMaskFlag2, OtherMaskBit2)
	dst.SetOtherMask(src.OtherMaskFlag3, OtherMaskBit3)

	return
}
```

## 1.5. 新增
如果后续增加了 **MaskBit** 和 **MaskFlag**， 如果未做任何处理
1. 对于 `parsemask`， 新的 **MaskFlag** 会提示 **NotMatch**, 错误类似 `// no match: dst.OtherMaskFlag4`
2. 对于 `buildmask`， 因为本身并未新增字段， 无法通过类似的方式暴露问题
   1. 为了解决此问题， 对于**buildmask**， 系统自动拉取 **MaskBit** 所有已经定义的值， 然后检查**每个值都必须在 `:parsemask` 中定义**, 否则报错， 错误类似 `maskBit 'sample.OtherMaskBit' value 'OtherMaskBit4' matched flag NOT exist`

## 1.6. 废弃Bit
如果某个Bit**废弃**， 但是Bit定义还是**保留（占位）**， BizModel 的bool **flag字段 删除**， 此时会命中下面的规则

> MaskFlag 必须存在， 错误类似 `mask type 'sample.OtherMaskBit' value 'OtherMaskBit4' matched flag NOT exist`

为了解决这个报错， 需要定义下面的注释, **MaskFlag** 用 `‘-’` 代替
``` go
// :buildmask OtherMask - OtherMaskBit4
NewModelFromBiz(*BizModel) *Model
```

此时会生成如下**代码（注释）**, 这样一眼就能看出 **OtherMaskBit4** 没有对应的 **maskFlag**
``` go
// skip 'OtherMask.OtherMaskBit4'
dst.SetOtherMask(src.OtherMaskFlag1, OtherMaskBit1)
dst.SetOtherMask(src.OtherMaskFlag2, OtherMaskBit2)
dst.SetOtherMask(src.OtherMaskFlag3, OtherMaskBit3)
```

# 2. Update

Update 用于在实际的更新数据库记录时， 简化开发工作

如果更新PropertyMask, 需要对已有的bit进行合并， 有两种方案
1. 直接在db层面做bit操作， 例如 `update table set f = (f | 1) & ~2 where id = 1`
2. 先读取已有记录， 再合并bit， 最后整体更新 PropertyMask

这个功能**默认关闭**， 为了使用此功能， 需要增加 **msku package**

``` go
import (
	"go.lixinio.com/apis/pkg/msku"
)

type _ msku.Mask
```

同时需要引入一个struct对象， 通过**lxg query** 可以生成， 以上面的BizModel为例， 生成代码（节选）如下
``` go
type unexportedBizModelQuery struct {
	PropertyMaskFlag1 string
	PropertyMaskFlag2 string
	PropertyMaskFlag3 string
	OtherMaskFlag1    string
	OtherMaskFlag2    string
	OtherMaskFlag3    string
}

var BizModelQuery = unexportedBizModelQuery{
	PropertyMaskFlag1: "property_mask_flag1",
	PropertyMaskFlag2: "property_mask_flag2",
	PropertyMaskFlag3: "property_mask_flag3",
	OtherMaskFlag1:    "other_mask_flag1",
	OtherMaskFlag2:    "other_mask_flag2",
	OtherMaskFlag3:    "other_mask_flag3",
}
```

## 2.1. Bit操作
增加 注释 **:mask**  (注意： 和 **:buildmask** 一起 ， `不是` **:parsemask**)
``` go
	// :buildmask OtherMask - OtherMaskBit4
	// :mask BizModelQuery
	NewModelFromBiz(*BizModel) *Model
```

第一个参数是一个**struct指针变量**， 这个struct包含每个 **:buildmask** 定义的 **maskFlag** 的string值

如果不存在， 错误类似 `field 'OtherMaskFlag3' NOT in struct 'BizModelQuery' varible`

> 生成的代码如下
> 
> 为了演示实际的情况， 增加两个字段
``` go
type XX struct {
	Field1       int
	Field2       int
}
```
``` go
func NewModelFromBizWithMaskToMap(
	src *BizModel,
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
) (map[string]any, error) {
	var (
		setOtherMask      uint32
		unsetOtherMask    uint32
		setPropertyMask   int8
		unsetPropertyMask int8
		result            = map[string]any{}
		setOtherMaskFn    = func(flag bool, mask OtherMaskBit) {
			if flag {
				setOtherMask |= uint32(mask)
			} else {
				unsetOtherMask |= uint32(mask)
			}
		}
		setPropertyMaskFn = func(flag bool, mask PropertyMaskBit) {
			if flag {
				setPropertyMask |= int8(mask)
			} else {
				unsetPropertyMask |= int8(mask)
			}
		}
	)

	dst := NewModelFromBiz(src)

	for key := range masker {
		switch key {
		case BizModelQuery.OtherMaskFlag1:
			setOtherMaskFn(src.OtherMaskFlag1, OtherMaskBit1)
		case BizModelQuery.OtherMaskFlag2:
			setOtherMaskFn(src.OtherMaskFlag2, OtherMaskBit2)
		case BizModelQuery.OtherMaskFlag3:
			setOtherMaskFn(src.OtherMaskFlag3, OtherMaskBit3)
		case BizModelQuery.PropertyMaskFlag1:
			setPropertyMaskFn(src.PropertyMaskFlag1, PropertyMaskBit1)
		case BizModelQuery.PropertyMaskFlag2:
			setPropertyMaskFn(src.PropertyMaskFlag2, PropertyMaskBit2)
		case BizModelQuery.PropertyMaskFlag3:
			setPropertyMaskFn(src.PropertyMaskFlag3, PropertyMaskBit3)
		}
	}

	if v, err := transfer.Uint32("OtherMask", setOtherMask, unsetOtherMask); err != nil {
		return nil, err
	} else if v != nil {
		result["OtherMask"] = v
	}

	if v, err := transfer.Int8("PropertyMask", setPropertyMask, unsetPropertyMask); err != nil {
		return nil, err
	} else if v != nil {
		result["PropertyMask"] = v
	}

	newMasker := msku.TransferMask(masker, (*Model)(nil).MaskMap())
	newMasker = msku.TransferMaskToCamel(newMasker)

	for key := range newMasker {
		switch key {
		case "Field1":
			result["Field1"] = dst.Field1
		case "Field2":
			result["Field2"] = dst.Field2
		case "OtherMask": // skip
		case "PropertyMask": // skip
		default:
			return nil, fmt.Errorf("invalid mask field '%s'", key)
		}
	}

	return result, nil
}
```

只要启用了 Update 选项， 都会生成下面的映射关系

``` go
func (*Model) MaskMap() map[string]string {
	return map[string]string{
		BizModelQuery.OtherMaskFlag1:    "OtherMask",
		BizModelQuery.OtherMaskFlag2:    "OtherMask",
		BizModelQuery.OtherMaskFlag3:    "OtherMask",
		BizModelQuery.PropertyMaskFlag1: "PropertyMask",
		BizModelQuery.PropertyMaskFlag2: "PropertyMask",
		BizModelQuery.PropertyMaskFlag3: "PropertyMask",
	}
}
```

## 2.2. 排除字段
某些字段禁止更新， 或者它本身是一个关联字段， 以gorm为例， 见 <https://gorm.io/zh_CN/docs/belongs_to.html>, 有 `Belongs To` `Has One` `Has Many` `Many To Many` 等情况

若需要排除， 用注释, 从第二个参数开始， 支持任意个需要忽略的参数（匹配DbModel的fieldName）
``` go
	// :buildmask OtherMask - OtherMaskBit4
	// :mask BizModelQuery Field1 ...
	NewModelFromBiz(*BizModel) *Model
```

此时生成的代码会去掉 Field1 （仅剩 Field2） , 当传入一个Field1的mask进来， 会报错， 因为不支持更新
``` go
	for key := range newMasker {
		switch key {
		case "Field2":
			result["Field2"] = dst.Field2
		case "OtherMask": // skip
		case "PropertyMask": // skip
		default:
			return nil, fmt.Errorf("invalid mask field '%s'", key)
		}
	}

	return result, nil
}
```

## 2.3. 自定义映射
上面的代码有一个假设， BizModelQuery的值， 通过 `ToCamel` (见 `newMasker = msku.TransferMaskToCamel(newMasker)`) 可以匹配 DbModel 的Field 名称

假如 `BizModelQuery.CreateTime = "create_time"`,  DbModel的字段名是 `CreatedAt`, 那就需要再**调用前先根据映射转换**， 再调用 `NewModelFromBizWithMaskToMap`

``` go
masker = msku.TransferMask(masker, map[string]string{
    // "CreatedAt" 应该避免裸写字符串， 应该用代码生成或其他手段， 降低维护成本
    BizModelQuery.CreateTime: "CreatedAt",
})
```

# 2.4. 合并bit
增加 注释 **:mask:ext**  (注意： 和 **:buildmask** 一起 ， 不是 **:parsemask**)
```
	// :buildmask OtherMask - OtherMaskBit4
	// :mask:ext BizModelQuery
	NewModelFromBiz(*BizModel) *Model
}
```

第一个参数 和  `:mask` 一致

生成代码如下
``` go
func NewModelFromBizWithMask(
	src *BizModel,
	existFn func() (*Model, error),
	masker msku.Mask,
) (*Model, msku.Mask, error) {
	dst := NewModelFromBiz(src)

	// 转换并检查是否包含mask字段
	newMasker := msku.TransferMask(masker, dst.MaskMap())
	if !newMasker.IsExist("OtherMask") &&
		!newMasker.IsExist("PropertyMask") {
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
		case BizModelQuery.OtherMaskFlag1:
			old.SetOtherMask(src.OtherMaskFlag1, OtherMaskBit1)
		case BizModelQuery.OtherMaskFlag2:
			old.SetOtherMask(src.OtherMaskFlag2, OtherMaskBit2)
		case BizModelQuery.OtherMaskFlag3:
			old.SetOtherMask(src.OtherMaskFlag3, OtherMaskBit3)
		case BizModelQuery.PropertyMaskFlag1:
			old.SetPropertyMask(src.PropertyMaskFlag1, PropertyMaskBit1)
		case BizModelQuery.PropertyMaskFlag2:
			old.SetPropertyMask(src.PropertyMaskFlag2, PropertyMaskBit2)
		case BizModelQuery.PropertyMaskFlag3:
			old.SetPropertyMask(src.PropertyMaskFlag3, PropertyMaskBit3)
		}
	}

	// 回写 (combine了未更新的propertyMask bit)
	dst.OtherMask = old.OtherMask
	dst.PropertyMask = old.PropertyMask

	return dst, newMasker, nil
}
```

和  `:mask` 一样， 也会生成函数 `func (*Model) MaskMap() map[string]string`

# 3. TODO
+ [ ] 去掉 msku 依赖
