package domain

import (
	"strconv"

	"github.com/krylphi/helloworld-data-handler/internal/errs"
	"github.com/krylphi/helloworld-data-handler/internal/utils"

	"github.com/valyala/fastjson"
)

// Entry log entry
type Entry struct {
	// ContentID content id
	ContentID int64 `json:"content_id"`
	// Timestamp unix timestamp (with nano)
	Timestamp int64 `json:"timestamp"`
	//ClientID id of the sender
	ClientID int `json:"client_id"`
	// Text message text
	Text string `json:"text"`
}

const maxMillis = 9999999999999

// ParseEntry parse entry from json
func ParseEntry(json []byte) (*Entry, error) {
	var p fastjson.Parser
	v, err := p.Parse(string(json))
	if err != nil {
		return nil, err
	}
	res := &Entry{}
	res.ClientID = v.GetInt("client_id")
	if res.ClientID == 0 {
		return nil, errs.ErrInvalidClientID
	}
	res.Timestamp = v.GetInt64("timestamp")
	if res.Timestamp <= 0 || res.Timestamp > maxMillis {
		return nil, errs.ErrInvalidTimestamp
	}
	res.ContentID = v.GetInt64("content_id")
	if res.ContentID == 0 {
		return nil, errs.ErrInvalidContentID
	}
	res.Text = string(v.GetStringBytes("text"))
	if res.Text == "" {
		return nil, errs.ErrEmptyText
	}
	return res, nil
}

// Marshal marshals entry to json. We do not use json.Marshal because it's slower
func (e *Entry) Marshal() []byte {
	return []byte(utils.Concat(
		`{"text":"`, e.Text,
		`","content_id":`, strconv.FormatInt(e.ContentID, 10),
		`,"client_id":`, strconv.Itoa(e.ClientID),
		`,"timestamp":`, strconv.FormatInt(e.Timestamp, 10),
		`}`, "\r\n"))
}

// ValidateEntry validates request entry
func ValidateEntry(json []byte) error {
	var p fastjson.Parser
	v, err := p.Parse(string(json))
	if err != nil {
		return err
	}
	i := v.GetInt("client_id")
	if i == 0 {
		return errs.ErrInvalidClientID
	}
	i64 := v.GetInt64("timestamp")
	if i64 <= 0 || i64 > maxMillis {
		return errs.ErrInvalidTimestamp
	}
	i64 = v.GetInt64("content_id")
	if i64 == 0 {
		return errs.ErrInvalidContentID
	}
	s := string(v.GetStringBytes("text"))
	if s == "" {
		return errs.ErrEmptyText
	}
	return nil
}
