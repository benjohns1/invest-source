package app_test

import (
	"context"
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/benjohns1/invest-source/app"
	"github.com/stretchr/testify/assert"
)

func TestApp_CacheDailySourceData(t *testing.T) {
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name    string
		app     app.App
		args    args
		wantErr bool
	}{
		{
			name: "should fail if cache ReadCurrent() returns an error",
			app: app.App{Config: app.Config{
				Cache: func() app.Cache {
					c := mockCache{}
					c.On("ReadCurrent").Return(nil, fmt.Errorf("read cache error"))
					return &c
				}(),
				Provider: &mockProvider{},
				Output:   &mockOutput{},
			}},
			wantErr: true,
		},
		{
			name: "should fail if provider QueryLatest() returns an error",
			app: app.App{Config: app.Config{
				Cache: func() app.Cache {
					c := mockCache{}
					c.On("ReadCurrent").Return(nil, nil)
					return &c
				}(),
				Provider: func() app.Provider {
					p := mockProvider{}
					p.On("QueryLatest").Return(nil, fmt.Errorf("provider query error"))
					return &p
				}(),
				Output: &mockOutput{},
			}},
			wantErr: true,
		},
		{
			name: "should fail if cache WriteCurrent() returns an error",
			app: app.App{Config: app.Config{
				Cache: func() app.Cache {
					c := mockCache{}
					c.On("ReadCurrent").Return(nil, nil)
					c.On("WriteCurrent", []byte("query data response")).Return(fmt.Errorf("write cache error"))
					return &c
				}(),
				Provider: func() app.Provider {
					p := mockProvider{}
					p.On("QueryLatest").Return([]byte("query data response"), nil)
					return &p
				}(),
				Output: &mockOutput{},
			}},
			wantErr: true,
		},
		{
			name: "should succeed if cache ReadCurrent() returns data",
			app: app.App{Config: app.Config{
				Cache: func() app.Cache {
					c := mockCache{}
					c.On("ReadCurrent").Return([]byte("{}"), nil)
					return &c
				}(),
				Provider: &mockProvider{},
				Output:   &mockOutput{},
			}},
			wantErr: false,
		},
		{
			name: "should succeed if provider QueryLatest() returns data and cache WriteCurrent() succeeds",
			app: app.App{Config: app.Config{
				Cache: func() app.Cache {
					c := mockCache{}
					c.On("ReadCurrent").Return(nil, nil)
					c.On("WriteCurrent", []byte("query data response")).Return(nil)
					return &c
				}(),
				Provider: func() app.Provider {
					p := mockProvider{}
					p.On("QueryLatest").Return([]byte("query data response"), nil)
					return &p
				}(),
				Output: &mockOutput{},
			}},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.args.ctx == nil {
				tt.args.ctx = context.Background()
			}
			if tt.app.Config.Log == nil {
				tt.app.Config.Log = log.New(os.Stdout, "test: ", log.LstdFlags)
			}
			err := app.CacheDailySourceData(tt.args.ctx, tt.app)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			if c, ok := tt.app.Config.Cache.(*mockCache); ok {
				c.AssertExpectations(t)
			}
			if p, ok := tt.app.Config.Provider.(*mockProvider); ok {
				p.AssertExpectations(t)
			}
			if o, ok := tt.app.Config.Output.(*mockOutput); ok {
				o.AssertExpectations(t)
			}
		})
	}
}
