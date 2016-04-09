package pg

import (
	"fmt"
	"github.com/jackc/pgx"
)

type NullString struct {
	String string
	Valid  bool // Valid is true if String is not NULL
}

func (s *NullString) Scan(vr *pgx.ValueReader) error {
	// Not checking oid as so we can scan anything into into a NullString - may revisit this decision later

	if vr.Len() == -1 {
		s.String, s.Valid = "", false
		return nil
	}

	s.Valid = true
	s.String = decodeText(vr)
	return vr.Err()
}

func (n NullString) FormatCode() int16 { return pgx.TextFormatCode }

func (s NullString) Encode(w *pgx.WriteBuf, oid pgx.Oid) error {
	if !s.Valid {
		w.WriteInt32(-1)
		return nil
	}

	return encodeText(w, s.String)
}

func decodeText(vr *pgx.ValueReader) string {
	if vr.Len() == -1 {
		vr.Fatal(pgx.ProtocolError("Cannot decode null into string"))
		return ""
	}

	return vr.ReadString(vr.Len())
}

func encodeText(w *pgx.WriteBuf, value interface{}) error {
	switch t := value.(type) {
	case string:
		w.WriteInt32(int32(len(t)))
		w.WriteBytes([]byte(t))
	case []byte:
		w.WriteInt32(int32(len(t)))
		w.WriteBytes(t)
	default:
		return fmt.Errorf("Expected string, received %T", value)
	}

	return nil
}
