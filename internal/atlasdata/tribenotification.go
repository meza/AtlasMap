package atlasdata

type fTribeEntity struct {
	EntityID                                 FUInt32Property
	ParentEntityID                           FUInt32Property
	EntityType                               FByteProperty
	ShipType                                 FByteProperty
	EntityName                               FStringProperty
	ServerID                                 FUInt32Property
	ServerRelativeLocationInCurrentServerMap FVector2DProperty
	NextAllowedUseTime                       FUInt32Property
	BInLandClaimedFlagRange                  FBoolProperty
	BReachedMaxTravelCount                   FBoolProperty
	BIsDead                                  FBoolProperty
}

type Chat struct {
	SenderName       FString
	SenderSteamName  FString
	SenderTribeName  FString
	SenderID         uint32 `struc:"uint32,little"`
	Message          FString
	SenderTeamIndex  int32 `struc:"int32,little"`
	SendMode         FString
	UserID           FString
	BUseAdminIcon    bool  `struc:"bool"`
	BIsTribeOwner    bool  `struc:"bool"`
	PlayerBadgeGroup int32 `struc:"int32,little"`
}

type AddRemoveEntity struct {
	BIsNewEntity          bool `struc:"bool"`
	BIsJustLocationChange bool `struc:"bool"`
	TribeEntity           fTribeEntity
}

type MemberPresenceUpdated struct {
	PlayerID     uint32 `struc:"uint32,little"`
	LastOnlineAt int32  `struc:"int32,little"`
}
