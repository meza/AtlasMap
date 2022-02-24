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

type FTribeEntity struct {
	EntityID                                 FUInt32Property
	ParentEntityID                           FUInt32Property
	EntityType                               FByteProperty
	ShipType                                 FByteProperty
	EntityName                               FStringProperty
	ServerId                                 FUInt32Property
	ServerRelativeLocationInCurrentServerMap FVector2DProperty
	NextAllowedUseTime                       FUInt32Property
	BInLandClaimedFlagRange                  FBoolProperty
	BReachedMaxTravelCount                   FBoolProperty
	BIsDead                                  FBoolProperty
}

type FTribeNotificationChat struct {
	SenderName       FString
	SenderSteamName  FString
	SenderTribeName  FString
	SenderId         uint32 `struc:"uint32,little"`
	Message          FString
	SenderTeamIndex  int32 `struc:"int32,little"`
	SendMode         FString
	UserId           FString
	BUseAdminIcon    bool  `struc:"bool"`
	BIsTribeOwner    bool  `struc:"bool"`
	PlayerBadgeGroup int32 `struc:"int32,little"`
}

type FTribeNotificationAddRemoveEntity struct {
	BIsNewEntity          bool `struc:"bool"`
	BIsJustLocationChange bool `struc:"bool"`
	TribeEntity           FTribeEntity
}

type BubbleWrap struct {
}
