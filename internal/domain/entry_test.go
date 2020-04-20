package domain

import (
	"fmt"
	"github.com/krylphi/helloworld-data-handler/internal/errs"
	"reflect"
	"testing"
)

func TestEntry_Marshal(t *testing.T) {
	type fields struct {
		ContentID int64
		Timestamp int64
		ClientID  int
		Text      string
	}
	tests := []struct {
		name   string
		fields fields
		want   []byte
	}{
		{
			name: "OK",
			fields: fields{
				ContentID: 1,
				Timestamp: 1586846680064,
				ClientID:  1,
				Text:      "hello world",
			},
			want: []byte("{\"text\":\"hello world\",\"content_id\":1,\"client_id\":1,\"timestamp\":1586846680064}\r\n"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &Entry{
				ContentID: tt.fields.ContentID,
				Timestamp: tt.fields.Timestamp,
				ClientID:  tt.fields.ClientID,
				Text:      tt.fields.Text,
			}
			if got := e.Marshal(); !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("Marshal() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseEntry(t *testing.T) {
	type args struct {
		json []byte
	}
	tests := []struct {
		name    string
		args    args
		want    *Entry
		wantErr bool
		err     error
	}{
		{
			name: "OK",
			args: args{
				json: []byte("{\"text\": \"hello world\", \"content_id\": 1, \"client_id\":1, \"timestamp\": 1586846680064}"),
			},
			want: &Entry{
				ContentID: 1,
				Timestamp: 1586846680064,
				ClientID:  1,
				Text:      "hello world",
			},
		},
		{
			name: "ErrInvalidContentID",
			args: args{
				json: []byte("{\"text\": \"hello world\", \"client_id\":1, \"timestamp\": 1586846680064}"),
			},
			want:    nil,
			err:     errs.ErrInvalidContentID,
			wantErr: true,
		},
		{
			name: "ErrInvalidClientID",
			args: args{
				json: []byte("{\"text\": \"hello world\", \"content_id\": 1, \"timestamp\": 1586846680064}"),
			},
			want:    nil,
			err:     errs.ErrInvalidClientID,
			wantErr: true,
		},
		{
			name: "ErrInvalidTimestamp",
			args: args{
				json: []byte("{\"text\": \"hello world\", \"content_id\": 1, \"client_id\":1, \"timestamp\": 15868466800640}"),
			},
			want:    nil,
			err:     errs.ErrInvalidTimestamp,
			wantErr: true,
		},
		{
			name: "ErrEmptyText",
			args: args{
				json: []byte("{\"content_id\": 1, \"client_id\":1, \"timestamp\": 1586846680064}"),
			},
			want:    nil,
			err:     errs.ErrEmptyText,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseEntry(tt.args.json)
			if (err != nil) != tt.wantErr {
				t.Fatalf("ParseEntry() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err != nil {
				if err.Error() != tt.err.Error() {
					t.Fatalf("ParseEntry() error = %v, wantErr %v", err.Error(), tt.err.Error())
					return
				}
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("ParseEntry() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestValidateEntry(t *testing.T) {
	type args struct {
		json []byte
	}
	tests := []struct {
		name    string
		args    args
		err     error
		wantErr bool
	}{
		{
			name: "OK",
			args: args{
				json: []byte("{\"text\": \"hello world\", \"content_id\": 1, \"client_id\":1, \"timestamp\": 1586846680064}"),
			},
		},
		{
			name: "ErrInvalidContentID",
			args: args{
				json: []byte("{\"text\": \"hello world\", \"client_id\":1, \"timestamp\": 1586846680064}"),
			},
			err:     errs.ErrInvalidContentID,
			wantErr: true,
		},
		{
			name: "ErrInvalidClientID",
			args: args{
				json: []byte("{\"text\": \"hello world\", \"content_id\": 1, \"timestamp\": 1586846680064}"),
			},
			err:     errs.ErrInvalidClientID,
			wantErr: true,
		},
		{
			name: "ErrInvalidTimestamp",
			args: args{
				json: []byte("{\"text\": \"hello world\", \"content_id\": 1, \"client_id\":1, \"timestamp\": 15868466800640}"),
			},
			err:     errs.ErrInvalidTimestamp,
			wantErr: true,
		},
		{
			name: "ErrEmptyText",
			args: args{
				json: []byte("{\"content_id\": 1, \"client_id\":1, \"timestamp\": 1586846680064}"),
			},
			err:     errs.ErrEmptyText,
			wantErr: true,
		},
		{
			name: "InvalidJson",
			args: args{
				json: []byte("{\"content_id\": 1, \"client_id\":1, \"timestamp\": 1586846680064"),
			},
			err:     fmt.Errorf(`cannot parse JSON: cannot parse object: unexpected end of object; unparsed tail: ""`),
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateEntry(tt.args.json)
			if (err != nil) != tt.wantErr {
				t.Fatalf("ValidateEntry() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr && err != nil {
				if err.Error() != tt.err.Error() {
					t.Fatalf("ParseEntry() error = %v, wantErr %v", err.Error(), tt.err.Error())
					return
				}
			}
		})
	}
}
