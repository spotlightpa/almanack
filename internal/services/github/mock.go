package github

import (
	"context"
	"os"
	"path/filepath"

	"github.com/spotlightpa/almanack/pkg/almlog"
)

type MockClient struct {
	basepath string
}

func NewMockClient(dirs ...string) *MockClient {
	almlog.Logger.Warn("mocking Github", "dir", dirs)
	return &MockClient{filepath.Join(dirs...)}
}

func (mc *MockClient) abspath(path string) string {
	return filepath.Join(mc.basepath, path)
}

func (mc *MockClient) ensureParent(fn string) {
	tmppath := filepath.Dir(fn)
	_ = os.MkdirAll(tmppath, os.ModePerm)
}

func (mc *MockClient) CreateFile(ctx context.Context, msg, path string, content []byte) error {
	tmpfn := mc.abspath(path)
	l := almlog.FromContext(ctx)
	l.InfoContext(ctx, "github.mock.CreateFile",
		"path", tmpfn,
	)
	mc.ensureParent(tmpfn)
	return os.WriteFile(tmpfn, content, 0644)
}

func (mc *MockClient) GetFile(ctx context.Context, path string) (contents string, err error) {
	tmpfn := mc.abspath(path)
	l := almlog.FromContext(ctx)
	l.InfoContext(ctx, "github.mock.GetFile",
		"path", tmpfn,
	)
	var b []byte
	b, err = os.ReadFile(tmpfn)
	return string(b), err
}

func (mc *MockClient) UpdateFile(ctx context.Context, msg, path string, content []byte) error {
	tmpfn := mc.abspath(path)
	l := almlog.FromContext(ctx)
	l.InfoContext(ctx, "github.mock.UpdateFile",
		"path", tmpfn,
	)
	mc.ensureParent(tmpfn)
	return os.WriteFile(tmpfn, content, 0644)
}

// ErrorClient is a test client that just always returns an error.
type ErrorClient struct {
	Error error
}

func (ec ErrorClient) CreateFile(ctx context.Context, msg, path string, content []byte) error {
	return ec.Error
}

func (ec ErrorClient) GetFile(ctx context.Context, path string) (contents string, err error) {
	return "", ec.Error
}

func (ec ErrorClient) UpdateFile(ctx context.Context, msg, path string, content []byte) error {
	return ec.Error
}
