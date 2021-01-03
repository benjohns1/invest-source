package app_test

import (
	"github.com/benjohns1/invest-source/app"
	"github.com/stretchr/testify/mock"
)

type mockCache struct {
	mock.Mock
}

func (mc *mockCache) ReadCurrent() ([]byte, error) {
	args := mc.Called()
	retB, _ := args.Get(0).([]byte)
	return retB, args.Error(1)
}

func (mc *mockCache) WriteCurrent(data []byte) error {
	args := mc.Called(data)
	return args.Error(0)
}

type mockProvider struct {
	mock.Mock
}

func (mp *mockProvider) QueryLatest() ([]byte, error) {
	args := mp.Called()
	retB, _ := args.Get(0).([]byte)
	return retB, args.Error(1)
}

func (mp *mockProvider) ParseQuotes(data []byte) ([]app.Quote, error) {
	args := mp.Called(data)
	retQ, _ := args.Get(0).([]app.Quote)
	return retQ, args.Error(1)
}
