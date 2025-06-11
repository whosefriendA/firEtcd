package buntdbx

import (
	"reflect"
	"testing"

	"github.com/tidwall/buntdb"
	"github.com/whosefriendA/firEtcd/common"
)

func TestNewDB(t *testing.T) {
	tests := []struct {
		name string
		want *DB
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewDB(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewDB() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDB_Del(t *testing.T) {
	type fields struct {
		db *buntdb.DB
	}
	type args struct {
		key string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &DB{
				db: tt.fields.db,
			}
			if err := d.Del(tt.args.key); (err != nil) != tt.wantErr {
				t.Errorf("DB.Del() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDB_DelWithPrefix(t *testing.T) {
	type fields struct {
		db *buntdb.DB
	}
	type args struct {
		prefix string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &DB{
				db: tt.fields.db,
			}
			if err := d.DelWithPrefix(tt.args.prefix); (err != nil) != tt.wantErr {
				t.Errorf("DB.DelWithPrefix() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDB_Get(t *testing.T) {
	type fields struct {
		db *buntdb.DB
	}
	type args struct {
		key string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantRet []byte
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &DB{
				db: tt.fields.db,
			}
			gotRet, err := d.Get(tt.args.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("DB.Get() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotRet, tt.wantRet) {
				t.Errorf("DB.Get() = %v, want %v", gotRet, tt.wantRet)
			}
		})
	}
}

func TestDB_GetWithPrefix(t *testing.T) {
	type fields struct {
		db *buntdb.DB
	}
	type args struct {
		prefix string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantRet [][]byte
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &DB{
				db: tt.fields.db,
			}
			gotRet, err := d.GetWithPrefix(tt.args.prefix)
			if (err != nil) != tt.wantErr {
				t.Errorf("DB.GetWithPrefix() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotRet, tt.wantRet) {
				t.Errorf("DB.GetWithPrefix() = %v, want %v", gotRet, tt.wantRet)
			}
		})
	}
}

func TestDB_GetEntry(t *testing.T) {
	type fields struct {
		db *buntdb.DB
	}
	type args struct {
		key string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantRet common.Entry
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &DB{
				db: tt.fields.db,
			}
			gotRet, err := d.GetEntry(tt.args.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("DB.GetEntry() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotRet, tt.wantRet) {
				t.Errorf("DB.GetEntry() = %v, want %v", gotRet, tt.wantRet)
			}
		})
	}
}

func TestDB_GetEntryWithPrefix(t *testing.T) {
	type fields struct {
		db *buntdb.DB
	}
	type args struct {
		prefix string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantRet []common.Entry
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &DB{
				db: tt.fields.db,
			}
			gotRet, err := d.GetEntryWithPrefix(tt.args.prefix)
			if (err != nil) != tt.wantErr {
				t.Errorf("DB.GetEntryWithPrefix() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotRet, tt.wantRet) {
				t.Errorf("DB.GetEntryWithPrefix() = %v, want %v", gotRet, tt.wantRet)
			}
		})
	}
}

func TestDB_GetPairsWithPrefix(t *testing.T) {
	type fields struct {
		db *buntdb.DB
	}
	type args struct {
		prefix string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantRet []common.Pair
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &DB{
				db: tt.fields.db,
			}
			gotRet, err := d.GetPairsWithPrefix(tt.args.prefix)
			if (err != nil) != tt.wantErr {
				t.Errorf("DB.GetPairsWithPrefix() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotRet, tt.wantRet) {
				t.Errorf("DB.GetPairsWithPrefix() = %v, want %v", gotRet, tt.wantRet)
			}
		})
	}
}

func TestDB_Put(t *testing.T) {
	type fields struct {
		db *buntdb.DB
	}
	type args struct {
		key      string
		value    []byte
		DeadTime int64
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &DB{
				db: tt.fields.db,
			}
			if err := d.Put(tt.args.key, tt.args.value, tt.args.DeadTime); (err != nil) != tt.wantErr {
				t.Errorf("DB.Put() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDB_PutEntry(t *testing.T) {
	type fields struct {
		db *buntdb.DB
	}
	type args struct {
		key   string
		entry common.Entry
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &DB{
				db: tt.fields.db,
			}
			if err := d.PutEntry(tt.args.key, tt.args.entry); (err != nil) != tt.wantErr {
				t.Errorf("DB.PutEntry() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDB_Keys(t *testing.T) {
	type fields struct {
		db *buntdb.DB
	}
	type args struct {
		pageSize  int
		pageIndex int
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantRet []common.Pair
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &DB{
				db: tt.fields.db,
			}
			gotRet, err := d.Keys(tt.args.pageSize, tt.args.pageIndex)
			if (err != nil) != tt.wantErr {
				t.Errorf("DB.Keys() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotRet, tt.wantRet) {
				t.Errorf("DB.Keys() = %v, want %v", gotRet, tt.wantRet)
			}
		})
	}
}

func TestDB_KVs(t *testing.T) {
	type fields struct {
		db *buntdb.DB
	}
	type args struct {
		pageSize  int
		pageIndex int
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantRet []common.Pair
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &DB{
				db: tt.fields.db,
			}
			gotRet, err := d.KVs(tt.args.pageSize, tt.args.pageIndex)
			if (err != nil) != tt.wantErr {
				t.Errorf("DB.KVs() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotRet, tt.wantRet) {
				t.Errorf("DB.KVs() = %v, want %v", gotRet, tt.wantRet)
			}
		})
	}
}

func TestDB_SnapshotData(t *testing.T) {
	type fields struct {
		db *buntdb.DB
	}
	tests := []struct {
		name    string
		fields  fields
		want    []byte
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &DB{
				db: tt.fields.db,
			}
			got, err := d.SnapshotData()
			if (err != nil) != tt.wantErr {
				t.Errorf("DB.SnapshotData() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("DB.SnapshotData() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDB_InstallSnapshotData(t *testing.T) {
	type fields struct {
		db *buntdb.DB
	}
	type args struct {
		data []byte
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &DB{
				db: tt.fields.db,
			}
			if err := d.InstallSnapshotData(tt.args.data); (err != nil) != tt.wantErr {
				t.Errorf("DB.InstallSnapshotData() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDB_Close(t *testing.T) {
	type fields struct {
		db *buntdb.DB
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &DB{
				db: tt.fields.db,
			}
			if err := d.Close(); (err != nil) != tt.wantErr {
				t.Errorf("DB.Close() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
