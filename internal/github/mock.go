package github

import (
	"context"
	"os"
	"path/filepath"

	"github.com/spotlightpa/almanack/pkg/common"
)

type MockClient struct {
	basepath string
	l        common.Logger
}

func NewMockClient(l common.Logger) (*MockClient, error) {
	dir, err := os.MkdirTemp("", "example")
	if err != nil {
		return nil, err
	}
	// we don't clean up temp dir:
	// good for testing but don't use this in prod!
	l.Printf("mock Github base dir is %s", dir)
	return &MockClient{dir, l}, nil
}

func (mc *MockClient) abspath(path string) string {
	return filepath.Join(mc.basepath, path)
}

func (mc *MockClient) ensureParent(fn string) {
	tmppath := filepath.Dir(fn)
	os.MkdirAll(tmppath, os.ModePerm)
}

func (mc *MockClient) CreateFile(ctx context.Context, msg, path string, content []byte) error {
	tmpfn := mc.abspath(path)
	mc.l.Printf("creating file %s on mock Github", tmpfn)
	mc.ensureParent(tmpfn)
	return os.WriteFile(tmpfn, content, 0644)
}

func (mc *MockClient) GetFile(ctx context.Context, path string) (contents string, err error) {
	tmpfn := mc.abspath(path)
	mc.l.Printf("getting file %s from mock Github", tmpfn)
	var b []byte
	b, err = os.ReadFile(tmpfn)
	return string(b), err
}

func (mc *MockClient) UpdateFile(ctx context.Context, msg, path string, content []byte) error {
	tmpfn := mc.abspath(path)
	mc.l.Printf("updating file %s on mock Github", tmpfn)
	mc.ensureParent(tmpfn)
	return os.WriteFile(tmpfn, content, 0644)
}
