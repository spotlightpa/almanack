package aws_test

import (
	"crypto/md5"
	"github.com/earthboundkid/assert"
	"github.com/spotlightpa/almanack/internal/almlog"
	"github.com/spotlightpa/almanack/internal/services/aws"
	"os"
	"path/filepath"
	"testing"
)

func TestMD5(t *testing.T) {
	be := assert.FailNow(t)
	almlog.UseTestLogger(t)
	dir := t.ArtifactDir()
	const teststr = "Hello, World!"
	wantMD5 := md5.Sum([]byte(teststr))

	ctx := t.Context()
	bucket := aws.NewTestBlobStore(dir)
	err := bucket.WriteFile(ctx, "hello.txt", nil, []byte(teststr))
	be.Zero(err)

	hash, size, err := bucket.ReadMD5(ctx, "hello.txt")
	be.Zero(err)
	be.SlicesEqual(hash, wantMD5[:])
	be.Equal(size, int64(len(teststr)))

	be.Zero(os.Remove(filepath.Join(dir, "hello.txt.attrs")))

	hash, size, err = bucket.ReadMD5(ctx, "hello.txt")
	be.Zero(err)
	be.SlicesEqual(hash, wantMD5[:])
	be.EqualLength(teststr, int(size))
}
