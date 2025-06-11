package gateway

import (
	"reflect"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/whosefriendA/firEtcd/client"
	"github.com/whosefriendA/firEtcd/pkg/firconfig"
)

func TestNewGateway(t *testing.T) {
	type args struct {
		conf firconfig.Gateway
	}
	tests := []struct {
		name string
		args args
		want *Gateway
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewGateway(tt.args.conf); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewGateway() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGateway_Run(t *testing.T) {
	type fields struct {
		ck   *client.Clerk
		conf firconfig.Gateway
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
			g := &Gateway{
				ck:   tt.fields.ck,
				conf: tt.fields.conf,
			}
			if err := g.Run(); (err != nil) != tt.wantErr {
				t.Errorf("Gateway.Run() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGateway_LoadRouter(t *testing.T) {
	type fields struct {
		ck   *client.Clerk
		conf firconfig.Gateway
	}
	type args struct {
		r *gin.Engine
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
			g := &Gateway{
				ck:   tt.fields.ck,
				conf: tt.fields.conf,
			}
			g.LoadRouter(tt.args.r)
		})
	}
}

func TestGateway_watch(t *testing.T) {
	type fields struct {
		ck   *client.Clerk
		conf firconfig.Gateway
	}
	type args struct {
		c *gin.Context
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
			g := &Gateway{
				ck:   tt.fields.ck,
				conf: tt.fields.conf,
			}
			g.watch(tt.args.c)
		})
	}
}

func TestGateway_keys(t *testing.T) {
	type fields struct {
		ck   *client.Clerk
		conf firconfig.Gateway
	}
	type args struct {
		c *gin.Context
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
			g := &Gateway{
				ck:   tt.fields.ck,
				conf: tt.fields.conf,
			}
			g.keys(tt.args.c)
		})
	}
}

func TestGateway_kvs(t *testing.T) {
	type fields struct {
		ck   *client.Clerk
		conf firconfig.Gateway
	}
	type args struct {
		c *gin.Context
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
			g := &Gateway{
				ck:   tt.fields.ck,
				conf: tt.fields.conf,
			}
			g.kvs(tt.args.c)
		})
	}
}

func TestGateway_put(t *testing.T) {
	type fields struct {
		ck   *client.Clerk
		conf firconfig.Gateway
	}
	type args struct {
		c *gin.Context
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
			g := &Gateway{
				ck:   tt.fields.ck,
				conf: tt.fields.conf,
			}
			g.put(tt.args.c)
		})
	}
}

func TestGateway_putCAS(t *testing.T) {
	type fields struct {
		ck   *client.Clerk
		conf firconfig.Gateway
	}
	type args struct {
		c *gin.Context
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
			g := &Gateway{
				ck:   tt.fields.ck,
				conf: tt.fields.conf,
			}
			g.putCAS(tt.args.c)
		})
	}
}

func TestGateway_del(t *testing.T) {
	type fields struct {
		ck   *client.Clerk
		conf firconfig.Gateway
	}
	type args struct {
		c *gin.Context
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
			g := &Gateway{
				ck:   tt.fields.ck,
				conf: tt.fields.conf,
			}
			g.del(tt.args.c)
		})
	}
}

func TestGateway_delWithPrefix(t *testing.T) {
	type fields struct {
		ck   *client.Clerk
		conf firconfig.Gateway
	}
	type args struct {
		c *gin.Context
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
			g := &Gateway{
				ck:   tt.fields.ck,
				conf: tt.fields.conf,
			}
			g.delWithPrefix(tt.args.c)
		})
	}
}

func TestGateway_get(t *testing.T) {
	type fields struct {
		ck   *client.Clerk
		conf firconfig.Gateway
	}
	type args struct {
		c *gin.Context
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
			g := &Gateway{
				ck:   tt.fields.ck,
				conf: tt.fields.conf,
			}
			g.get(tt.args.c)
		})
	}
}

func TestGateway_getWithPrefix(t *testing.T) {
	type fields struct {
		ck   *client.Clerk
		conf firconfig.Gateway
	}
	type args struct {
		c *gin.Context
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
			g := &Gateway{
				ck:   tt.fields.ck,
				conf: tt.fields.conf,
			}
			g.getWithPrefix(tt.args.c)
		})
	}
}
