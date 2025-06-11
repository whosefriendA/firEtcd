package firconfig

import "testing"

func TestWriteRemote(t *testing.T) {
	type args struct {
		conf Firconfig
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := WriteRemote(tt.args.conf); (err != nil) != tt.wantErr {
				t.Errorf("WriteRemote() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestReadRemote(t *testing.T) {
	type args struct {
		conf Firconfig
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ReadRemote(tt.args.conf)
		})
	}
}

func TestInit(t *testing.T) {
	type args struct {
		Path string
		conf Firconfig
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			Init(tt.args.Path, tt.args.conf)
		})
	}
}

func TestWriteLocal(t *testing.T) {
	type args struct {
		Path string
		conf Firconfig
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := WriteLocal(tt.args.Path, tt.args.conf); (err != nil) != tt.wantErr {
				t.Errorf("WriteLocal() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestReadLocal(t *testing.T) {
	type args struct {
		Path string
		conf Firconfig
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := ReadLocal(tt.args.Path, tt.args.conf); (err != nil) != tt.wantErr {
				t.Errorf("ReadLocal() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
