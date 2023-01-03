package atlasdb

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"strconv"
	"strings"

	"github.com/antihax/AtlasMap/internal/atlasdata"
	"github.com/go-redis/redis/v8"
	"github.com/lunixbochs/struc"
	"github.com/rs/zerolog/log"
)

// TribeData is the redis structure for ATLAS tribes
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
	LogLineID                                     int64  `redis:"logline"`
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
	ServerID       uint32  `redis:"ServerId"`
	X              float32 `redis:"ServerXRelativeLocation"`
	Y              float32 `redis:"ServerYRelativeLocation"`
	IsDead         bool    `redis:"bIsDead"`
}

// SubTribe returns a channel pumped with UE event data from the tribe.
func (s *AtlasDB) SubTribe(ctx context.Context, tribeID int64) <-chan string {
	sub := s.db.Subscribe(ctx, "tribemsg:"+strconv.FormatInt(tribeID, 10))
	channel := make(chan string, 40)
	go s.processTribeChannel(ctx, channel, sub)
	return channel
}

func (s *AtlasDB) processTribeMessage(msg string, channel chan string, CRC int32) error {

	// 1652749511 Tribe Log
	// 1466483860 remove entity

	switch CRC {
	case -1646244981: // MemberPresenceUpdated
		{
			b := &atlasdata.MemberPresenceUpdated{}
			err := struc.Unpack(strings.NewReader(msg), b)
			if err != nil {
				log.Err(err).Msg("Unpack")
				return err
			}
		}
	case 156265321:
		{
			b := &atlasdata.Chat{}
			err := struc.Unpack(strings.NewReader(msg), b)
			if err != nil {
				log.Err(err).Msg("Unpack")
				return err
			}
		}
	case 834710557:
		{
			// [TODO] find a better way to remove trailing null byte
			b := &atlasdata.AddRemoveEntity{}
			err := struc.Unpack(strings.NewReader(msg), b)
			if err != nil {
				log.Err(err).Msg("Unpack")
				return err
			}
			v, err := json.Marshal(TribeEntityUpdate{
				EntityID:       b.TribeEntity.EntityID.Value,
				ParentEntityID: b.TribeEntity.ParentEntityID.Value,
				EntityType:     strings.TrimRight(b.TribeEntity.EntityType.Value.String, "\u0000"),
				ShipType:       strings.TrimRight(b.TribeEntity.ShipType.Value.String, "\u0000"),
				EntityName:     strings.TrimRight(b.TribeEntity.EntityName.Value.String, "\u0000"),
				ServerID:       b.TribeEntity.ServerID.Value,
				X:              b.TribeEntity.ServerRelativeLocationInCurrentServerMap.Value.X,
				Y:              b.TribeEntity.ServerRelativeLocationInCurrentServerMap.Value.Y,
				IsDead:         b.TribeEntity.BIsDead.Value,
			})
			if err != nil {
				log.Err(err).Msg("Marshal")
				return err
			}
			channel <- string(v)
		}
	default:
		log.Info().Msgf("unknown crc %d", CRC)
		log.Debug().Msg(hex.Dump([]byte(msg)))
	}
	return nil
}

func (s *AtlasDB) processTribeChannel(ctx context.Context, channel chan string, sub *redis.PubSub) {
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

			// Unpack header from the message
			bubbleWrap := &atlasdata.BubbleWrap{}
			err = struc.Unpack(strings.NewReader(msg.Payload[:12]), bubbleWrap)
			if err != nil {
				log.Err(err).Msg("Unpack")
			}
			if err := s.processTribeMessage(msg.Payload[12:], channel, bubbleWrap.CRC); err != nil {

			}
		}
	}
}
