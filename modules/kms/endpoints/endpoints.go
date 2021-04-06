// Package endpoints defines all route handlers.
package endpoints

import (
	"net/http"

	"github.com/erda-project/erda/pkg/httpserver"
	"github.com/erda-project/erda/pkg/kms"
)

// Endpoints 定义 endpoint 方法
type Endpoints struct {
	KmsMgr *kms.Manager
}

type Option func(*Endpoints)

// New 创建 Endpoints 对象.
func New(options ...Option) *Endpoints {
	e := &Endpoints{}

	for _, op := range options {
		op(e)
	}

	return e
}

func WithKmsManager(mgr *kms.Manager) Option {
	return func(e *Endpoints) {
		e.KmsMgr = mgr
	}
}

// Routes 返回 endpoints 的所有 endpoint 方法，也就是 route.
func (e *Endpoints) Routes() []httpserver.Endpoint {
	return []httpserver.Endpoint{
		{Path: "/health", Method: http.MethodGet, Handler: e.Health},

		// kms
		{Path: "/api/kms", Method: http.MethodPost, Handler: e.KmsCreateKey},
		{Path: "/api/kms/encrypt", Method: http.MethodPost, Handler: e.KmsEncrypt},
		{Path: "/api/kms/decrypt", Method: http.MethodPost, Handler: e.KmsDecrypt},
		{Path: "/api/kms/generate-data-key", Method: http.MethodPost, Handler: e.KmsGenerateDataKey},
		{Path: "/api/kms/rotate-key-version", Method: http.MethodPost, Handler: e.KmsRotateKeyVersion},
		{Path: "/api/kms/describe-key", Method: http.MethodGet, Handler: e.KmsRotateKeyVersion},
	}
}