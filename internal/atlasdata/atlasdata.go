// Package atlasdata contains data mapped to the redis raw formats
// Redis data is packed little endian
package atlasdata

type FString struct {
	Size   int    `struc:"int32,little,sizeof=String"`
	String string `struc:"[]byte"`
}

type FVector2D struct {
	X float32 `struc:"float32,little"`
	Y float32 `struc:"float32,little"`
}

type FProperty struct {
	Name FString
	Type FString
}

type FStringProperty struct {
	FProperty
	PropertyFlags uint64
	Value         FString
}

type FUInt32Property struct {
	FProperty
	Extra FString
	Value uint32 `struc:"uint32,little"`
}

type FVector2DProperty struct {
	FProperty
	PropertyFlags uint64
	Extra         FString
	Value         FVector2D
}

type FByteProperty struct {
	FProperty
	PropertyFlags uint64
	ValueType     FString
	Value         FString
}

type FBoolProperty struct {
	FProperty
	PropertyFlags uint64
	Value         bool
}

type BubbleWrap struct {
	ServerVersion int32  `struc:"int32,little"`
	ServerID      uint32 `struc:"uint32,little"`
	CRC           int32  `struc:"int32,little"`
}
