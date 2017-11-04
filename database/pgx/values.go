package pgx

import (
	"fmt"
	"github.com/jackc/pgx"
	"math"
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

// NullInt16 represents a smallint that may be null. NullInt16 implements the
// Scanner and Encoder interfaces so it may be used both as an argument to
// Query[Row] and a destination for Scan for prepared and unprepared queries.
//
// If Valid is false then the value is NULL.
type NullInt16 struct {
	Int16 int16
	Valid bool // Valid is true if Int16 is not NULL
}

func (n *NullInt16) Scan(vr *pgx.ValueReader) error {
	if vr.Type().DataType != pgx.Int2Oid {
		return pgx.SerializationError(fmt.Sprintf("NullInt16.Scan cannot decode OID %d", vr.Type().DataType))
	}

	if vr.Len() == -1 {
		n.Int16, n.Valid = 0, false
		return nil
	}
	n.Valid = true
	n.Int16 = decodeInt2(vr)
	return vr.Err()
}

func (n NullInt16) FormatCode() int16 { return pgx.BinaryFormatCode }

func (n NullInt16) Encode(w *pgx.WriteBuf, oid pgx.Oid) error {
	if oid != pgx.Int2Oid {
		return pgx.SerializationError(fmt.Sprintf("NullInt16.Encode cannot encode into OID %d", oid))
	}

	if !n.Valid {
		w.WriteInt32(-1)
		return nil
	}

	return encodeInt16(w, oid, n.Int16)
}

// NullInt32 represents an integer that may be null. NullInt32 implements the
// Scanner and Encoder interfaces so it may be used both as an argument to
// Query[Row] and a destination for Scan.
//
// If Valid is false then the value is NULL.
type NullInt32 struct {
	Int32 int32
	Valid bool // Valid is true if Int32 is not NULL
}

func (n *NullInt32) Scan(vr *pgx.ValueReader) error {
	if vr.Type().DataType != pgx.Int4Oid {
		return pgx.SerializationError(fmt.Sprintf("NullInt32.Scan cannot decode OID %d", vr.Type().DataType))
	}

	if vr.Len() == -1 {
		n.Int32, n.Valid = 0, false
		return nil
	}
	n.Valid = true
	n.Int32 = decodeInt4(vr)
	return vr.Err()
}

func (n NullInt32) FormatCode() int16 { return pgx.BinaryFormatCode }

func (n NullInt32) Encode(w *pgx.WriteBuf, oid pgx.Oid) error {
	if oid != pgx.Int4Oid {
		return pgx.SerializationError(fmt.Sprintf("NullInt32.Encode cannot encode into OID %d", oid))
	}

	if !n.Valid {
		w.WriteInt32(-1)
		return nil
	}

	return encodeInt32(w, oid, n.Int32)
}

// NullInt64 represents an bigint that may be null. NullInt64 implements the
// Scanner and Encoder interfaces so it may be used both as an argument to
// Query[Row] and a destination for Scan.
//
// If Valid is false then the value is NULL.
type NullInt64 struct {
	Int64 int64
	Valid bool // Valid is true if Int64 is not NULL
}

func (n *NullInt64) Scan(vr *pgx.ValueReader) error {
	if vr.Type().DataType != pgx.Int8Oid {
		return pgx.SerializationError(fmt.Sprintf("NullInt64.Scan cannot decode OID %d", vr.Type().DataType))
	}

	if vr.Len() == -1 {
		n.Int64, n.Valid = 0, false
		return nil
	}
	n.Valid = true
	n.Int64 = decodeInt8(vr)
	return vr.Err()
}

func (n NullInt64) FormatCode() int16 { return pgx.BinaryFormatCode }

func (n NullInt64) Encode(w *pgx.WriteBuf, oid pgx.Oid) error {
	if oid != pgx.Int8Oid {
		return pgx.SerializationError(fmt.Sprintf("NullInt64.Encode cannot encode into OID %d", oid))
	}

	if !n.Valid {
		w.WriteInt32(-1)
		return nil
	}

	return encodeInt64(w, oid, n.Int64)
}

func decodeInt(vr *pgx.ValueReader) int64 {
	switch vr.Type().DataType {
	case pgx.Int2Oid:
		return int64(decodeInt2(vr))
	case pgx.Int4Oid:
		return int64(decodeInt4(vr))
	case pgx.Int8Oid:
		return int64(decodeInt8(vr))
	}

	vr.Fatal(pgx.ProtocolError(fmt.Sprintf("Cannot decode oid %v into any integer type", vr.Type().DataType)))
	return 0
}

func decodeInt2(vr *pgx.ValueReader) int16 {
	if vr.Len() == -1 {
		vr.Fatal(pgx.ProtocolError("Cannot decode null into int16"))
		return 0
	}

	if vr.Type().DataType != pgx.Int2Oid {
		vr.Fatal(pgx.ProtocolError(fmt.Sprintf("Cannot decode oid %v into int16", vr.Type().DataType)))
		return 0
	}

	if vr.Type().FormatCode != pgx.BinaryFormatCode {
		vr.Fatal(pgx.ProtocolError(fmt.Sprintf("Unknown field description format code: %v", vr.Type().FormatCode)))
		return 0
	}

	if vr.Len() != 2 {
		vr.Fatal(pgx.ProtocolError(fmt.Sprintf("Received an invalid size for an int2: %d", vr.Len())))
		return 0
	}

	return vr.ReadInt16()
}

func decodeInt4(vr *pgx.ValueReader) int32 {
	if vr.Len() == -1 {
		vr.Fatal(pgx.ProtocolError("Cannot decode null into int32"))
		return 0
	}

	if vr.Type().DataType != pgx.Int4Oid {
		vr.Fatal(pgx.ProtocolError(fmt.Sprintf("Cannot decode oid %v into int32", vr.Type().DataType)))
		return 0
	}

	if vr.Type().FormatCode != pgx.BinaryFormatCode {
		vr.Fatal(pgx.ProtocolError(fmt.Sprintf("Unknown field description format code: %v", vr.Type().FormatCode)))
		return 0
	}

	if vr.Len() != 4 {
		vr.Fatal(pgx.ProtocolError(fmt.Sprintf("Received an invalid size for an int4: %d", vr.Len())))
		return 0
	}

	return vr.ReadInt32()
}

func decodeInt8(vr *pgx.ValueReader) int64 {
	if vr.Len() == -1 {
		vr.Fatal(pgx.ProtocolError("Cannot decode null into int64"))
		return 0
	}

	if vr.Type().DataType != pgx.Int8Oid {
		vr.Fatal(pgx.ProtocolError(fmt.Sprintf("Cannot decode oid %v into int8", vr.Type().DataType)))
		return 0
	}

	if vr.Type().FormatCode != pgx.BinaryFormatCode {
		vr.Fatal(pgx.ProtocolError(fmt.Sprintf("Unknown field description format code: %v", vr.Type().FormatCode)))
		return 0
	}

	if vr.Len() != 8 {
		vr.Fatal(pgx.ProtocolError(fmt.Sprintf("Received an invalid size for an int8: %d", vr.Len())))
		return 0
	}

	return vr.ReadInt64()
}

func encodeInt(w *pgx.WriteBuf, oid pgx.Oid, value int) error {
	switch oid {
	case pgx.Int2Oid:
		if value < math.MinInt16 {
			return fmt.Errorf("%d is less than min pg:int2", value)
		} else if value > math.MaxInt16 {
			return fmt.Errorf("%d is greater than max pg:int2", value)
		}
		w.WriteInt32(2)
		w.WriteInt16(int16(value))
	case pgx.Int4Oid:
		if value < math.MinInt32 {
			return fmt.Errorf("%d is less than min pg:int4", value)
		} else if value > math.MaxInt32 {
			return fmt.Errorf("%d is greater than max pg:int4", value)
		}
		w.WriteInt32(4)
		w.WriteInt32(int32(value))
	case pgx.Int8Oid:
		if int64(value) <= int64(math.MaxInt64) {
			w.WriteInt32(8)
			w.WriteInt64(int64(value))
		} else {
			return fmt.Errorf("%d is larger than max int64 %d", value, int64(math.MaxInt64))
		}
	default:
		return fmt.Errorf("cannot encode %s into oid %v", "int8", oid)
	}

	return nil
}

func encodeInt16(w *pgx.WriteBuf, oid pgx.Oid, value int16) error {
	switch oid {
	case pgx.Int2Oid:
		w.WriteInt32(2)
		w.WriteInt16(value)
	case pgx.Int4Oid:
		w.WriteInt32(4)
		w.WriteInt32(int32(value))
	case pgx.Int8Oid:
		w.WriteInt32(8)
		w.WriteInt64(int64(value))
	default:
		return fmt.Errorf("cannot encode %s into oid %v", "int16", oid)
	}

	return nil
}

func encodeInt32(w *pgx.WriteBuf, oid pgx.Oid, value int32) error {
	switch oid {
	case pgx.Int2Oid:
		if value <= math.MaxInt16 {
			w.WriteInt32(2)
			w.WriteInt16(int16(value))
		} else {
			return fmt.Errorf("%d is greater than max int16 %d", value, math.MaxInt16)
		}
	case pgx.Int4Oid:
		w.WriteInt32(4)
		w.WriteInt32(value)
	case pgx.Int8Oid:
		w.WriteInt32(8)
		w.WriteInt64(int64(value))
	default:
		return fmt.Errorf("cannot encode %s into oid %v", "int32", oid)
	}

	return nil
}

func encodeInt64(w *pgx.WriteBuf, oid pgx.Oid, value int64) error {
	switch oid {
	case pgx.Int2Oid:
		if value <= math.MaxInt16 {
			w.WriteInt32(2)
			w.WriteInt16(int16(value))
		} else {
			return fmt.Errorf("%d is greater than max int16 %d", value, math.MaxInt16)
		}
	case pgx.Int4Oid:
		if value <= math.MaxInt32 {
			w.WriteInt32(4)
			w.WriteInt32(int32(value))
		} else {
			return fmt.Errorf("%d is greater than max int32 %d", value, math.MaxInt32)
		}
	case pgx.Int8Oid:
		w.WriteInt32(8)
		w.WriteInt64(value)
	default:
		return fmt.Errorf("cannot encode %s into oid %v", "int64", oid)
	}

	return nil
}
