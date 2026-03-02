package service

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestMapDownstreamError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		wantCode codes.Code
	}{
		{
			name:     "standard non-grpc error",
			err:      errors.New("some standard error"),
			wantCode: codes.Unavailable,
		},
		{
			name:     "grpc not found",
			err:      status.Error(codes.NotFound, "not found"),
			wantCode: codes.NotFound,
		},
		{
			name:     "grpc unavailable",
			err:      status.Error(codes.Unavailable, "unavailable"),
			wantCode: codes.Unavailable,
		},
		{
			name:     "grpc deadline exceeded",
			err:      status.Error(codes.DeadlineExceeded, "deadline exceeded"),
			wantCode: codes.Unavailable,
		},
		{
			name:     "grpc regular error",
			err:      status.Error(codes.Internal, "internal error"),
			wantCode: codes.Internal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := mapDownstreamError(tt.err)
			assert.Error(t, err)
			s, ok := status.FromError(err)
			assert.True(t, ok)
			assert.Equal(t, tt.wantCode, s.Code())
		})
	}
}

func TestIsTimeout(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "grpc deadline exceeded",
			err:  status.Error(codes.DeadlineExceeded, "deadline exceeded"),
			want: true,
		},
		{
			name: "context deadline exceeded",
			err:  context.DeadlineExceeded,
			want: true,
		},
		{
			name: "grpc other error",
			err:  status.Error(codes.Internal, "internal error"),
			want: false,
		},
		{
			name: "standard error",
			err:  errors.New("standard error"),
			want: false,
		},
		{
			name: "nil error",
			err:  nil,
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isTimeout(tt.err)
			assert.Equal(t, tt.want, got)
		})
	}
}
