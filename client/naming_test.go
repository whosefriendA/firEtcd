package client

import (
	"reflect"
	"testing"
	"time"
)

func TestNode_Marshal(t *testing.T) {
	type fields struct {
		Name     string
		AppId    string
		Port     string
		IPs      []string
		Location string
		Connect  int32
		Weight   int32
		Env      string
		MetaDate map[string]string
	}
	tests := []struct {
		name   string
		fields fields
		want   []byte
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := &Node{
				Name:     tt.fields.Name,
				AppId:    tt.fields.AppId,
				Port:     tt.fields.Port,
				IPs:      tt.fields.IPs,
				Location: tt.fields.Location,
				Connect:  tt.fields.Connect,
				Weight:   tt.fields.Weight,
				Env:      tt.fields.Env,
				MetaDate: tt.fields.MetaDate,
			}
			if got := n.Marshal(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Node.Marshal() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNode_Unmarshal(t *testing.T) {
	type fields struct {
		Name     string
		AppId    string
		Port     string
		IPs      []string
		Location string
		Connect  int32
		Weight   int32
		Env      string
		MetaDate map[string]string
	}
	type args struct {
		data []byte
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := &Node{
				Name:     tt.fields.Name,
				AppId:    tt.fields.AppId,
				Port:     tt.fields.Port,
				IPs:      tt.fields.IPs,
				Location: tt.fields.Location,
				Connect:  tt.fields.Connect,
				Weight:   tt.fields.Weight,
				Env:      tt.fields.Env,
				MetaDate: tt.fields.MetaDate,
			}
			n.Unmarshal(tt.args.data)
		})
	}
}

func TestNode_Key(t *testing.T) {
	type fields struct {
		Name     string
		AppId    string
		Port     string
		IPs      []string
		Location string
		Connect  int32
		Weight   int32
		Env      string
		MetaDate map[string]string
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := &Node{
				Name:     tt.fields.Name,
				AppId:    tt.fields.AppId,
				Port:     tt.fields.Port,
				IPs:      tt.fields.IPs,
				Location: tt.fields.Location,
				Connect:  tt.fields.Connect,
				Weight:   tt.fields.Weight,
				Env:      tt.fields.Env,
				MetaDate: tt.fields.MetaDate,
			}
			if got := n.Key(); got != tt.want {
				t.Errorf("Node.Key() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNode_SetNode(t *testing.T) {
	type fields struct {
		Name     string
		AppId    string
		Port     string
		IPs      []string
		Location string
		Connect  int32
		Weight   int32
		Env      string
		MetaDate map[string]string
	}
	type args struct {
		ck  *Clerk
		TTL time.Duration
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
			n := &Node{
				Name:     tt.fields.Name,
				AppId:    tt.fields.AppId,
				Port:     tt.fields.Port,
				IPs:      tt.fields.IPs,
				Location: tt.fields.Location,
				Connect:  tt.fields.Connect,
				Weight:   tt.fields.Weight,
				Env:      tt.fields.Env,
				MetaDate: tt.fields.MetaDate,
			}
			if err := n.SetNode(tt.args.ck, tt.args.TTL); (err != nil) != tt.wantErr {
				t.Errorf("Node.SetNode() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNode_SetNode_Watch(t *testing.T) {
	type fields struct {
		Name     string
		AppId    string
		Port     string
		IPs      []string
		Location string
		Connect  int32
		Weight   int32
		Env      string
		MetaDate map[string]string
	}
	type args struct {
		ck *Clerk
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			name: "basic setNode_watch test",
			fields: fields{
				Name:     "node1",
				AppId:    "app1",
				Port:     "8080",
				IPs:      []string{"127.0.0.1"},
				Location: "local",
				Connect:  10,
				Weight:   100,
				Env:      "dev",
				MetaDate: map[string]string{"version": "v1"},
			},
			args: args{
				ck: &Clerk{}, // 注意初始化成真实的 Clerk，如果 SetNode_Watch 有依赖
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := &Node{
				Name:     tt.fields.Name,
				AppId:    tt.fields.AppId,
				Port:     tt.fields.Port,
				IPs:      tt.fields.IPs,
				Location: tt.fields.Location,
				Connect:  tt.fields.Connect,
				Weight:   tt.fields.Weight,
				Env:      tt.fields.Env,
				MetaDate: tt.fields.MetaDate,
			}

			cancel := n.SetNode_Watch(tt.args.ck)
			if cancel == nil {
				t.Errorf("SetNode_Watch() returned nil cancel function")
				return
			}

			defer func() {
				if r := recover(); r != nil {
					t.Errorf("SetNode_Watch() cancel function panicked: %v", r)
				}
			}()
			cancel() // 调用一下看是否能正常取消
		})
	}
}

func TestGetNode(t *testing.T) {
	type args struct {
		ck   *Clerk
		name string
	}
	tests := []struct {
		name    string
		args    args
		want    []*Node
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetNode(tt.args.ck, tt.args.name)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetNode() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetNode() = %v, want %v", got, tt.want)
			}
		})
	}
}
