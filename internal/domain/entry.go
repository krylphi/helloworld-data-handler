package domain

import (
	"github.com/krylphi/helloworld-data-handler/internal/errs"
	"github.com/krylphi/helloworld-data-handler/internal/utils"
	"strconv"

	"github.com/valyala/fastjson"
)

// Entry log entry
type Entry struct {
	// ContentId content id
	ContentId int64 `json:"content_id"`
	// Timestamp unix timestamp (with nano)
	Timestamp int64 `json:"timestamp"`
	//ClientId id of the sender
	ClientId int `json:"client_id"`
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
	res.ClientId = v.GetInt("client_id")
	if res.ClientId == 0 {
		return nil, errs.ErrInvalidClientId
	}
	res.Timestamp = v.GetInt64("timestamp")
	if res.Timestamp <= 0 || res.Timestamp > maxMillis {
		return nil, errs.ErrInvalidTimestamp
	}
	res.ContentId = v.GetInt64("content_id")
	if res.ContentId == 0 {
		return nil, errs.ErrInvalidContentId
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
		`","content_id":`, strconv.FormatInt(e.ContentId, 10),
		`,"client_id":`, strconv.Itoa(e.ClientId),
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
		return errs.ErrInvalidClientId
	}
	i64 := v.GetInt64("timestamp")
	if i64 <= 0 || i64 > maxMillis {
		return errs.ErrInvalidTimestamp
	}
	i64 = v.GetInt64("content_id")
	if i64 == 0 {
		return errs.ErrInvalidContentId
	}
	s := string(v.GetStringBytes("text"))
	if s == "" {
		return errs.ErrEmptyText
	}
	return nil
}
