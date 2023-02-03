package models

import (
	"database/sql"
	"encoding/json"
	"time"
)
type JsonNullInt16 struct {
    sql.NullInt16
}

func (v *JsonNullInt16) MarshalJSON() ([]byte, error) {
    if v.Valid {
        return json.Marshal(v.Int16)
    } else {
        return json.Marshal(nil)
    }
}

func (v *JsonNullInt16) UnmarshalJSON(data []byte) error {
    // Unmarshalling into a pointer will let us detect null
    var x *int16
    if err := json.Unmarshal(data, &x); err != nil {
        return err
    }
    if x != nil {
        v.Valid = true
        v.Int16 = *x
    } else {
        v.Valid = false
    }
    return nil
}

type JsonNullInt64 struct {
    sql.NullInt64
}

func (v *JsonNullInt64) MarshalJSON() ([]byte, error) {
    if v.Valid {
        return json.Marshal(v.Int64)
    } else {
        return json.Marshal(nil)
    }
}

func (v *JsonNullInt64) UnmarshalJSON(data []byte) error {
    // Unmarshalling into a pointer will let us detect null
    var x *int64
    if err := json.Unmarshal(data, &x); err != nil {
        return err
    }
    if x != nil {
        v.Valid = true
        v.Int64 = *x
    } else {
        v.Valid = false
    }
    return nil
}


type JsonNullString struct {
    sql.NullString
}

func (v *JsonNullString) MarshalJSON() ([]byte, error) {
    if v.Valid {
        return json.Marshal(v.String)
    } else {
        return json.Marshal(nil)
    }
}

func (v *JsonNullString) UnmarshalJSON(data []byte) error {
    // Unmarshalling into a pointer will let us detect null
    var x *string
    if err := json.Unmarshal(data, &x); err != nil {
        return err
    }
    if x != nil {
        v.Valid = true
        v.String = *x
    } else {
        v.Valid = false
    }
    return nil
}

type JsonNullTime struct {
    sql.NullTime
}

func (v *JsonNullTime) MarshalJSON() ([]byte, error) {
    if v.Valid {
        return json.Marshal(v.Time)
    } else {
        return json.Marshal(nil)
    }
}

func (v *JsonNullTime) UnmarshalJSON(data []byte) error {
    // Unmarshalling into a pointer will let us detect null
    var x *time.Time
    if err := json.Unmarshal(data, &x); err != nil {
        return err
    }
    if x != nil {
        v.Valid = true
        v.Time = *x
    } else {
        v.Valid = false
    }
    return nil
}