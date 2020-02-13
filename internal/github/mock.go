package github

import (
	"context"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/spotlightpa/almanack/pkg/almanack"
)

type MockClient struct {
	basepath string
	l        almanack.Logger
}

func NewMockClient(l almanack.Logger) (*MockClient, error) {
	dir, err := ioutil.TempDir("", "example")
	if err != nil {
		return nil, err
	}
	// we don't clean up temp dir:
	// good for testing but don't use this in prod!
	l.Printf("mock Github base dir is %s", dir)
	return &MockClient{dir, l}, nil
}

func (mc *MockClient) CreateFile(ctx context.Context, msg, path string, content []byte) error {
	tmpfn := filepath.Join(mc.basepath, path)

	mc.l.Printf("creating file %s on mock Github", tmpfn)

	tmppath := filepath.Dir(tmpfn)
	os.MkdirAll(tmppath, os.ModePerm)

	return ioutil.WriteFile(tmpfn, content, 0644)
}
