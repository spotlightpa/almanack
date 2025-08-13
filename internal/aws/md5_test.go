package aws_test

import (
	"context"
	"crypto/md5"
	"os"
	"path/filepath"
	"testing"

	"github.com/carlmjohnson/be"
	"github.com/spotlightpa/almanack/internal/aws"
	"github.com/spotlightpa/almanack/pkg/almlog"
)

func TestMD5(t *testing.T) {
	almlog.UseTestLogger(t)
	dir := t.TempDir()
	const teststr = "Hello, World!"
	wantMD5 := md5.Sum([]byte(teststr))

	ctx := context.Background()
	bucket := aws.NewBlobStore("file://" + dir + "/")
	err := bucket.WriteFile(ctx, "hello.txt", nil, []byte(teststr))
	be.NilErr(t, err)

	hash, size, err := bucket.ReadMD5(ctx, "hello.txt")
	be.NilErr(t, err)
	be.AllEqual(t, wantMD5[:], hash)
	be.Equal(t, int64(len(teststr)), size)

	be.NilErr(t, os.Remove(filepath.Join(dir, "hello.txt.attrs")))

	hash, size, err = bucket.ReadMD5(ctx, "hello.txt")
	be.NilErr(t, err)
	be.AllEqual(t, wantMD5[:], hash)
	be.EqualLength(t, int(size), teststr)
}
