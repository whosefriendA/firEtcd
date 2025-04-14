package main

import (
	"reflect"
	"testing"
)

func TestStore_Delete(t *testing.T) {
	type fields struct {
		Nodes map[string]string
	}
	type args struct {
		key string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    string
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			name:    "Delete existing key",
			fields:  fields{Nodes: map[string]string{"key1": "value1"}},
			args:    args{key: "key1"},
			want:    "value1", // 被删除的旧值
			wantErr: false,
		},
		{
			name:    "Delete non-existing key",
			fields:  fields{Nodes: map[string]string{"key1": "value1"}},
			args:    args{key: "nonexistent"},
			want:    "",
			wantErr: true, // 删除不存在的 key 会报错
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Store{
				Nodes: tt.fields.Nodes,
			}
			got, err := s.Delete(tt.args.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("Delete() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Delete() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStore_Get(t *testing.T) {
	type fields struct {
		Nodes map[string]string
	}
	type args struct {
		key string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    string
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			name:    "Get existing key",
			fields:  fields{Nodes: map[string]string{"key1": "value1"}},
			args:    args{key: "key1"},
			want:    "value1",
			wantErr: false,
		},
		{
			name:    "Get non-existing key",
			fields:  fields{Nodes: map[string]string{"key1": "value1"}},
			args:    args{key: "nonexistent"},
			want:    "",
			wantErr: true, // 获取不存在的 key 会报错
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Store{
				Nodes: tt.fields.Nodes,
			}
			got, err := s.Get(tt.args.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("Get() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Get() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStore_Recovery(t *testing.T) {
	type fields struct {
		Nodes map[string]string
	}
	type args struct {
		state []byte
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			name:    "Successful recovery",
			fields:  fields{Nodes: map[string]string{"key1": "value1"}},
			args:    args{state: []byte(`{"nodes":{"key1":"value1"}}`)},
			wantErr: false,
		},
		{
			name:    "Failed recovery with corrupted state",
			fields:  fields{Nodes: map[string]string{"key1": "value1"}},
			args:    args{state: []byte(`{bad json}`)},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Store{
				Nodes: tt.fields.Nodes,
			}
			if err := s.Recovery(tt.args.state); (err != nil) != tt.wantErr {
				t.Errorf("Recovery() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestStore_Save(t *testing.T) {
	type fields struct {
		Nodes map[string]string
	}
	tests := []struct {
		name    string
		fields  fields
		want    []byte
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			name:    "Save existing store",
			fields:  fields{Nodes: map[string]string{"key1": "value1"}},
			want:    []byte(`{"nodes":{"key1":"value1"}}`),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Store{
				Nodes: tt.fields.Nodes,
			}
			got, err := s.Save()
			if (err != nil) != tt.wantErr {
				t.Errorf("Save() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Save() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStore_Set(t *testing.T) {
	type fields struct {
		Nodes map[string]string
	}
	type args struct {
		key   string
		value string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   string
		want1  bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Store{
				Nodes: tt.fields.Nodes,
			}
			got, got1 := s.Set(tt.args.key, tt.args.value)
			if got != tt.want {
				t.Errorf("Set() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("Set() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func Test_createStore(t *testing.T) {
	tests := []struct {
		name string
		want *Store
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := createStore(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("createStore() = %v, want %v", got, tt.want)
			}
		})
	}
}
