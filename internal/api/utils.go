// SPDX-FileCopyrightText: 2023 SAP SE
// SPDX-License-Identifier: Apache-2.0

package api

import (
	"net/http"
	"strconv"
	"time"

	"github.com/sapcc/go-bits/httpext"

	"github.com/sapcc/keppel/internal/auth"
	"github.com/sapcc/keppel/internal/keppel"
	"github.com/sapcc/keppel/internal/models"
)

func CheckRateLimit(r *http.Request, rle *keppel.RateLimitEngine, account models.ReducedAccount, authz *auth.Authorization, action keppel.RateLimitedAction, amount uint64) error {
	// rate-limiting is optional
	if rle == nil {
		return nil
	}

	// cluster-internal traffic is exempt from rate-limits (if the request is
	// caused by a user API request, the rate-limit has been checked already
	// before the cluster-internal request was sent)
	userType := authz.UserIdentity.UserType()
	if userType == keppel.PeerUser || userType == keppel.TrivyUser {
		return nil
	}

	allowed, result, err := rle.RateLimitAllows(r.Context(), httpext.GetRequesterIPFor(r), account, action, amount)
	if err != nil {
		return err
	}
	if !allowed {
		retryAfterStr := strconv.FormatUint(keppel.AtLeastZero(int64(result.RetryAfter/time.Second)), 10)
		return keppel.ErrTooManyRequests.With("").WithHeader("Retry-After", retryAfterStr)
	}

	return nil
}
