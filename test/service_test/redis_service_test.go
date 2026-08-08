// Copyright (c) 2026 Egor Dudnikov
// SPDX-License-Identifier: MIT

package service_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRateLimitCheckAllowed(t *testing.T) {
	env := newTestService(t)

	ok, err := env.svc.RateLimitCheck(context.Background(), "rate:1.1.1.1", 3)
	require.NoError(t, err)
	require.True(t, ok)
}

func TestRateLimitCheckExceeded(t *testing.T) {
	env := newTestService(t)

	key := "rate:2.2.2.2"
	for i := 0; i < 3; i++ {
		ok, err := env.svc.RateLimitCheck(context.Background(), key, 3)
		require.NoError(t, err)
		require.True(t, ok)
	}

	ok, err := env.svc.RateLimitCheck(context.Background(), key, 3)
	require.NoError(t, err)
	require.False(t, ok)
}
