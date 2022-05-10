package walker

import "testing"

func TestConvertSize(t *testing.T) {
	type args struct {
		size string
	}
	tests := []struct {
		name    string
		args    args
		want    int64
		wantErr bool
	}{
		{name: "100K",
			args:    args{size: "100K"},
			want:    102400,
			wantErr: false},
		{name: "wrong suffix",
			args:    args{size: "10l"},
			wantErr: true,
			want:    -1},
		{name: "bytes",
			args:    args{size: "345"},
			want:    345,
			wantErr: false},
		{name: "single byte",
			args:    args{size: "5"},
			wantErr: false,
			want:    5},
		{name: "zero",
			args:    args{size: "0"},
			want:    0,
			wantErr: false},
		{name: "empty",
			args:    args{size: ""},
			want:    -1,
			wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ConvertSize(tt.args.size)
			if (err != nil) != tt.wantErr {
				t.Errorf("ConvertSize() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ConvertSize() got = %v, want %v", got, tt.want)
			}
		})
	}
}
