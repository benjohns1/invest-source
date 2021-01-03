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

func TestApp_OutputDailyQuotes(t *testing.T) {
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
			app: app.App{
				Cache: func() app.Cache {
					c := mockCache{}
					c.On("ReadCurrent").Return(nil, fmt.Errorf("read cache error"))
					return &c
				}(),
				Provider: &mockProvider{},
				Output:   &mockOutput{},
			},
			wantErr: true,
		},
		{
			name: "should succeed with empty data",
			app: app.App{
				Cache: func() app.Cache {
					c := mockCache{}
					c.On("ReadCurrent").Return([]byte("{}"), nil)
					return &c
				}(),
				Provider: func() app.Provider {
					p := mockProvider{}
					p.On("ParseQuotes", []byte("{}")).Return([]app.Quote{}, nil)
					return &p
				}(),
				Output: func() app.Output {
					o := mockOutput{}
					o.On("Write", []app.Quote{}).Return(nil, nil)
					return &o
				}(),
			},
			wantErr: false,
		},
		{
			name: "should fail if provider ParseQuotes() returns an error",
			app: app.App{
				Cache: func() app.Cache {
					c := mockCache{}
					c.On("ReadCurrent").Return([]byte("{}"), nil)
					return &c
				}(),
				Provider: func() app.Provider {
					p := mockProvider{}
					p.On("ParseQuotes", []byte("{}")).Return(nil, fmt.Errorf("provider parsing error"))
					return &p
				}(),
				Output: &mockOutput{},
			},
			wantErr: true,
		},
		{
			name: "should fail if output Write() returns an error",
			app: app.App{
				Cache: func() app.Cache {
					c := mockCache{}
					c.On("ReadCurrent").Return([]byte("{}"), nil)
					return &c
				}(),
				Provider: func() app.Provider {
					p := mockProvider{}
					p.On("ParseQuotes", []byte("{}")).Return([]app.Quote{}, nil)
					return &p
				}(),
				Output: func() app.Output {
					o := mockOutput{}
					o.On("Write", []app.Quote{}).Return(nil, fmt.Errorf("output writer error"))
					return &o
				}(),
			},
			wantErr: true,
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
			err := tt.app.OutputDailyQuotes(tt.args.ctx)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			if c, ok := tt.app.Cache.(*mockCache); ok {
				c.AssertExpectations(t)
			}
			if p, ok := tt.app.Provider.(*mockProvider); ok {
				p.AssertExpectations(t)
			}
			if o, ok := tt.app.Output.(*mockOutput); ok {
				o.AssertExpectations(t)
			}
		})
	}
}