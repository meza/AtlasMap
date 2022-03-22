package atlasdb

import (
	"context"
	"encoding/json"
	"strconv"
	"strings"

	"github.com/antihax/AtlasMap/internal/atlasdata"
	"github.com/lunixbochs/struc"
	"github.com/rs/zerolog/log"
)

type TribeData struct {
	TribeID                                       int64  `redis:"TribeID"`
	TribeName                                     string `redis:"TribeName"`
	TribeAdmins                                   string `redis:"TribeAdmins"`
	BSetGovernment                                bool   `redis:"bSetGovernment"`
	TribeGovernmentPinCode                        int    `redis:"TribeGovernmentPinCode"`
	TribeGovernmentDinoOwnership                  int    `redis:"TribeGovernmentDinoOwnership"`
	TribeGovernmentStructureOwnership             int    `redis:"TribeGovernmentStructureOwnership"`
	TribeGovernmentDinoTaming                     int    `redis:"TribeGovernmentDinoTaming"`
	TribeGovernmentDinoUnclaimAdminOnly           int    `redis:"TribeGovernmentDinoUnclaimAdminOnly"`
	TribeOwnerPlayerDataID                        int64  `redis:"TribeOwnerPlayerDataID"`
	TribeRankGroups                               string `redis:"TribeRankGroups"`
	TribeGovernmentDefaultTerritoryBuildAllowance int    `redis:"TribeGovernmentDefaultTerritoryBuildAllowance"`
	TribeMOTD                                     string `redis:"TribeMOTD"`
	TribeMOTDNextChangeUTC                        int64  `redis:"TribeMOTDNextChangeUTC"`
	VerseBookState                                string `redis:"VerseBookState"`
	StolenVerseBookLastUsedAtUTC                  int64  `redis:"StolenVerseBookLastUsedAtUTC"`
	LogLineId                                     int64  `redis:"logline"`
}

// GetPlayerSteamID returns the SteamID of a playerID.
func (s *AtlasDB) GetTribeByID(ctx context.Context, tribeID int64) (*TribeData, error) {
	p := &TribeData{}
	if err := s.db.HGetAll(ctx, "tribedata:"+strconv.FormatInt(tribeID, 10)).Scan(p); err != nil {
		return nil, err
	}
	return p, nil
}

// GetTribeEntityIDList returns the tribe entitiy ID list.
func (s *AtlasDB) GetTribeEntityIDList(ctx context.Context, tribeID int64) ([]int64, error) {
	p := []int64{}
	if err := s.db.SMembers(ctx, "tribedata.entities:"+strconv.FormatInt(tribeID, 10)).ScanSlice(&p); err != nil {
		return nil, err
	}
	return p, nil
}

// GetTribeEntities returns the tribe entitiy ID list.
func (s *AtlasDB) GetTribeEntities(ctx context.Context, tribeID int64) ([]TribeEntityUpdate, error) {
	list := []TribeEntityUpdate{}
	ids, err := s.GetTribeEntityIDList(ctx, tribeID)
	if err != nil {
		return nil, err
	}

	for _, id := range ids {
		p := TribeEntityUpdate{}
		err := s.db.HGetAll(ctx, "entityinfo:"+strconv.FormatInt(id, 10)).Scan(&p)
		if err != nil {
			return nil, err
		}

		list = append(list, p)
	}

	return list, nil
}

type TribeEntityUpdate struct {
	EntityID       uint32  `redis:"EntityID"`
	ParentEntityID uint32  `redis:"ParentEntityID"`
	EntityType     string  `redis:"EntityType"`
	ShipType       string  `redis:"ShipType"`
	EntityName     string  `redis:"EntityName"`
	ServerId       uint32  `redis:"ServerId"`
	X              float32 `redis:"ServerXRelativeLocation"`
	Y              float32 `redis:"ServerYRelativeLocation"`
	IsDead         bool    `redis:"bIsDead"`
}

// SubTribe returns a channel pumped with json event data from the tribe.
func (s *AtlasDB) SubTribe(ctx context.Context, tribeID int64) <-chan string {
	sub := s.db.Subscribe(ctx, "tribemsg:"+strconv.FormatInt(tribeID, 10))
	channel := make(chan string, 10)
	go func() {
		for {
			select {
			case <-ctx.Done():
				close(channel)
				return
			default:
				// go-redis doesn't support cancellations 🙃🙃🙃
				// [TODO] Trace and add cancellation support
				msg, err := sub.ReceiveMessage(ctx)
				if err != nil {
					log.Err(err).Msg("SubTribe")
				}

				/*dmp := hex.Dumper(os.Stdout)
				dmp.Write([]byte(msg.Payload[12:]))
				dmp.Close()*/
				pos := 12
				switch []byte(msg.Payload)[8] {
				case 105:
					{
						b := &atlasdata.FTribeNotificationChat{}
						err = struc.Unpack(strings.NewReader(msg.Payload[pos:]), b)
						if err != nil {
							log.Err(err).Msg("Unpack")
						}
					}
				case 29:
					{
						// [TODO] find a better way to remove trailing null byte
						b := &atlasdata.FTribeNotificationAddRemoveEntity{}
						err = struc.Unpack(strings.NewReader(msg.Payload[pos:]), b)
						if err != nil {
							log.Err(err).Msg("Unpack")
						}
						v, err := json.Marshal(TribeEntityUpdate{
							EntityID:       b.TribeEntity.EntityID.Value,
							ParentEntityID: b.TribeEntity.ParentEntityID.Value,
							EntityType:     strings.TrimRight(b.TribeEntity.EntityType.Value.String, "\u0000"),
							ShipType:       strings.TrimRight(b.TribeEntity.ShipType.Value.String, "\u0000"),
							EntityName:     strings.TrimRight(b.TribeEntity.EntityName.Value.String, "\u0000"),
							ServerId:       b.TribeEntity.ServerId.Value,
							X:              b.TribeEntity.ServerRelativeLocationInCurrentServerMap.Value.X,
							Y:              b.TribeEntity.ServerRelativeLocationInCurrentServerMap.Value.Y,
							IsDead:         b.TribeEntity.BIsDead.Value,
						})
						if err != nil {
							log.Err(err).Msg("Marshal")
						}
						channel <- string(v)
					}
				}
			}
		}

	}()
	return channel
}
