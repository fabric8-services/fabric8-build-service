package controller

import (
	"context"
	"testing"
	"time"

	"github.com/fabric8-services/fabric8-build/app/test"
	"github.com/goadesign/goa"
	"github.com/stretchr/testify/assert"
)

func TestShowStatus(t *testing.T) {
	var (
		service = goa.New("status-test")
		ctrl    = NewStatusController(service)
	)
	_, res := test.ShowStatusOK(t, context.Background(), service, ctrl)

	assert.Equal(t, "0", res.Commit, "Commit not found")
	assert.Equal(t, StartTime, res.StartTime, "StartTime is not correct")
	_, err := time.Parse("2006-01-02T15:04:05Z", res.StartTime)
	assert.Nil(t, err, "Incorrect layout of StartTime")
}
