package app_test

import (
	"context"
	"fmt"
	"github.com/benjohns1/invest-source/app"
	"github.com/stretchr/testify/assert"
	"log"
	"os"
	"testing"
)

func TestApp_OutputDailySourceCSV(t *testing.T) {
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name    string
		app app.App
		args    args
		wantErr bool
	}{
		{
			name: "should fail if cache ReadCurrent() returns an error",
			app: app.App{
				Cache:    func() app.Cache {
					c := mockCache{}
					c.On("ReadCurrent").Return(nil, fmt.Errorf("read cache error"))
					return &c
				}(),
				Provider: mockProvider{},
			},
			wantErr: true,
		},
		{
			name: "should fail if provider QueryLatest() returns an error",
			app: app.App{
				Cache:    func() app.Cache {
					c := mockCache{}
					c.On("ReadCurrent").Return(nil, nil)
					return &c
				}(),
				Provider: func() app.Provider {
					p := mockProvider{}
					p.On("QueryLatest").Return(nil, fmt.Errorf("provider query error"))
					return &p
				}(),
			},
			wantErr: true,
		},
		{
			name: "should fail if cache WriteCurrent() returns an error",
			app: app.App{
				Cache:    func() app.Cache {
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
			},
			wantErr: true,
		},
		{
			name: "should succeed if cache ReadCurrent() returns data",
			app: app.App{
				Cache:    func() app.Cache {
					c := mockCache{}
					c.On("ReadCurrent").Return([]byte("cached data"), nil)
					return &c
				}(),
				Provider: mockProvider{},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.args.ctx == nil {
				tt.args.ctx = context.Background()
			}
			if tt.app.Log == nil {
				tt.app.Log = log.New(os.Stdout, "test: ", log.LstdFlags)
			}
			err := tt.app.OutputDailySourceCSV(tt.args.ctx)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			if c, ok := tt.app.Cache.(mockCache); ok {
				c.AssertExpectations(t)
			}
			if p, ok := tt.app.Provider.(mockProvider); ok {
				p.AssertExpectations(t)
			}
		})
	}
}
