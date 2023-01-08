package atlasdb

import (
	"encoding/binary"
	"strconv"
)

// ServerID unpacks the packed server ID. Each Server has an X and Y ID which
// corresponds to its 2D location in the game world. The ID is packed into
// 32-bits as follows:
//   +--------------+--------------+
//   | X (uint16_t) | Y (uint16_t) |
//   +--------------+--------------+
func ServerID(packed string) (split [2]uint16, err error) {
	var id uint64
	id, err = strconv.ParseUint(packed, 10, 32)
	if err != nil {
		return
	}

	buf := make([]byte, 4)
	binary.LittleEndian.PutUint32(buf, uint32(id))
	split[0] = binary.LittleEndian.Uint16(buf[:2])
	split[1] = binary.LittleEndian.Uint16(buf[2:])
	return
}
