// SPDX-FileCopyrightText: 2018-2020 SAP SE
// SPDX-License-Identifier: Apache-2.0

package keppel

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/sapcc/go-bits/errext"
	"github.com/sapcc/go-bits/respondwith"
)

// RegistryV2ErrorCode is the closed set of error codes that can appear in type
// RegistryV2Error.
type RegistryV2ErrorCode string

// Possible values for RegistryV2ErrorCode.
const (
	ErrBlobUnknown         RegistryV2ErrorCode = "BLOB_UNKNOWN"
	ErrBlobUploadInvalid   RegistryV2ErrorCode = "BLOB_UPLOAD_INVALID"
	ErrBlobUploadUnknown   RegistryV2ErrorCode = "BLOB_UPLOAD_UNKNOWN"
	ErrDigestInvalid       RegistryV2ErrorCode = "DIGEST_INVALID"
	ErrManifestBlobUnknown RegistryV2ErrorCode = "MANIFEST_BLOB_UNKNOWN"
	ErrManifestInvalid     RegistryV2ErrorCode = "MANIFEST_INVALID"
	ErrManifestUnknown     RegistryV2ErrorCode = "MANIFEST_UNKNOWN"
	ErrManifestUnverified  RegistryV2ErrorCode = "MANIFEST_UNVERIFIED"
	ErrNameInvalid         RegistryV2ErrorCode = "NAME_INVALID"
	ErrNameUnknown         RegistryV2ErrorCode = "NAME_UNKNOWN"
	ErrSizeInvalid         RegistryV2ErrorCode = "SIZE_INVALID"
	ErrTagInvalid          RegistryV2ErrorCode = "TAG_INVALID"
	ErrUnauthorized        RegistryV2ErrorCode = "UNAUTHORIZED"
	ErrDenied              RegistryV2ErrorCode = "DENIED"
	ErrUnsupported         RegistryV2ErrorCode = "UNSUPPORTED"

	// not in opencontainers/distribution-spec, but appears in github.com/docker/distribution
	ErrUnknown         RegistryV2ErrorCode = "UNKNOWN"
	ErrUnavailable     RegistryV2ErrorCode = "UNAVAILABLE"
	ErrTooManyRequests RegistryV2ErrorCode = "TOOMANYREQUESTS"
)

// With is a convenience function for constructing type RegistryV2Error.
func (c RegistryV2ErrorCode) With(msg string, args ...any) *RegistryV2Error {
	if msg == "" {
		msg = apiErrorMessages[c]
	} else if len(args) > 0 {
		msg = fmt.Sprintf(msg, args...)
	}
	return &RegistryV2Error{
		Code:    c,
		Message: msg,
	}
}

var apiErrorMessages = map[RegistryV2ErrorCode]string{
	ErrBlobUnknown:         "blob unknown to registry",
	ErrBlobUploadInvalid:   "blob upload invalid",
	ErrBlobUploadUnknown:   "blob upload unknown to registry",
	ErrDigestInvalid:       "provided digest did not match uploaded content",
	ErrManifestBlobUnknown: "manifest blob unknown to registry",
	ErrManifestInvalid:     "manifest invalid",
	ErrManifestUnknown:     "manifest unknown",
	ErrManifestUnverified:  "manifest failed signature verification",
	ErrNameInvalid:         "invalid repository name",
	ErrNameUnknown:         "repository name not known to registry",
	ErrSizeInvalid:         "provided length did not match content length",
	ErrTagInvalid:          "manifest tag did not match URI",
	ErrUnauthorized:        "authentication required",
	ErrDenied:              "requested access to the resource is denied",
	ErrUnsupported:         "operation is unsupported",
	ErrUnknown:             "unknown error",
	ErrUnavailable:         "registry is currently unavailable",
	ErrTooManyRequests:     "too many requests; please slow down",
}

var apiErrorStatusCodes = map[RegistryV2ErrorCode]int{
	ErrBlobUnknown:         http.StatusNotFound,
	ErrBlobUploadInvalid:   http.StatusBadRequest,
	ErrBlobUploadUnknown:   http.StatusNotFound,
	ErrDigestInvalid:       http.StatusBadRequest,
	ErrManifestBlobUnknown: http.StatusNotFound,
	ErrManifestInvalid:     http.StatusBadRequest,
	ErrManifestUnknown:     http.StatusNotFound,
	ErrManifestUnverified:  http.StatusBadRequest,
	ErrNameInvalid:         http.StatusBadRequest,
	ErrNameUnknown:         http.StatusNotFound,
	ErrSizeInvalid:         http.StatusBadRequest,
	ErrTagInvalid:          http.StatusBadRequest,
	ErrUnauthorized:        http.StatusUnauthorized,
	ErrDenied:              http.StatusUnauthorized, // 403 would make more sense, but we need to show 401 for bug-for-bug compatibility with docker-registry, see e.g. <https://github.com/google/go-containerregistry/issues/724>
	ErrUnsupported:         http.StatusMethodNotAllowed,
	ErrUnknown:             http.StatusInternalServerError,
	ErrUnavailable:         http.StatusServiceUnavailable,
	ErrTooManyRequests:     http.StatusTooManyRequests,
}

// RegistryV2Error is the error type expected by clients of the docker-registry
// v2 API.
type RegistryV2Error struct {
	Code    RegistryV2ErrorCode `json:"code"`
	Message string              `json:"message"`
	// Detail is always a string for errors generated by Keppel, but may be a JSON
	// object (i.e. map[string]any or similar) for errors coming from
	// keppel-registry.
	Detail  any         `json:"detail"`
	Status  int         `json:"-"`
	Headers http.Header `json:"-"`
}

// AsRegistryV2Error tries to cast `err` into RegistryV2Error. If `err` is not a
// RegistryV2Error, it gets wrapped in ErrUnknown instead.
func AsRegistryV2Error(err error) *RegistryV2Error {
	if rerr, ok := errext.As[*RegistryV2Error](err); ok {
		return rerr
	}
	return ErrUnknown.With(err.Error())
}

// WithDetail adds detail information to this error.
func (e *RegistryV2Error) WithDetail(detail any) *RegistryV2Error {
	e.Detail = detail
	return e
}

// WithStatus changes the HTTP status code for this error.
func (e *RegistryV2Error) WithStatus(status int) *RegistryV2Error {
	e.Status = status
	return e
}

// WithHeader adds a HTTP response header to this error.
func (e *RegistryV2Error) WithHeader(key string, values ...string) *RegistryV2Error {
	if e.Headers == nil {
		e.Headers = make(http.Header)
	}
	e.Headers[http.CanonicalHeaderKey(key)] = values
	return e
}

// WriteAsRegistryV2ResponseTo reports this error in the format used by the Registry V2 API.
func (e *RegistryV2Error) WriteAsRegistryV2ResponseTo(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	for k, v := range e.Headers {
		w.Header()[k] = v
	}
	if e.Status == 0 {
		w.WriteHeader(apiErrorStatusCodes[e.Code])
	} else {
		w.WriteHeader(e.Status)
	}
	if r.Method != http.MethodHead {
		buf, _ := json.Marshal(struct {
			Errors []*RegistryV2Error `json:"errors"`
		}{
			Errors: []*RegistryV2Error{e},
		})
		w.Write(append(buf, '\n'))
	}
}

// WriteAsAuthResponseTo reports this error in the format used by the Auth API endpoint.
func (e *RegistryV2Error) WriteAsAuthResponseTo(w http.ResponseWriter) {
	for k, v := range e.Headers {
		w.Header()[k] = v
	}
	status := e.Status
	if status == 0 {
		status = apiErrorStatusCodes[e.Code]
	}
	respondwith.JSON(w, status, map[string]string{"details": e.Error()})
}

// WriteAsTextTo reports this error in a plain text format.
func (e *RegistryV2Error) WriteAsTextTo(w http.ResponseWriter) {
	for k, v := range e.Headers {
		w.Header()[k] = v
	}
	if e.Status == 0 {
		w.WriteHeader(apiErrorStatusCodes[e.Code])
	} else {
		w.WriteHeader(e.Status)
	}
	w.Write([]byte(e.Error() + "\n"))
}

// Error implements the builtin/error interface.
func (e *RegistryV2Error) Error() string {
	text := e.Message
	if e.Detail != nil {
		detailStr, ok := e.Detail.(string)
		if !ok {
			detailBytes, err := json.Marshal(e.Detail)
			if err == nil {
				detailStr = string(detailBytes)
			} else {
				detailStr = err.Error()
			}
		}
		text += ": " + detailStr
	}
	return text
}
