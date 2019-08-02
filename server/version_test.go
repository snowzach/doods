package server

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/snowzach/doods/conf"
	"github.com/snowzach/doods/mocks"
)

func TestVersionGet(t *testing.T) {

	// Mock Store and server
	ts := new(mocks.ThingStore)
	s, err := New(ts)
	assert.Nil(t, err)

	response, err := s.Version(context.Background(), nil)
	assert.Nil(t, err)
	assert.Equal(t, conf.GitVersion, response.Version)

	// Check remaining expectations
	ts.AssertExpectations(t)

}
