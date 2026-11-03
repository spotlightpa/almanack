package aws_test

import (
	"crypto/md5"
	"os"
	"path/filepath"
	"testing"

	"github.com/earthboundkid/assert"
	"github.com/spotlightpa/almanack/internal/almlog"
	"github.com/spotlightpa/almanack/internal/services/aws"
)

func TestMD5(t *testing.T) {
	be := assert.FailsNow(t)
	almlog.UseTestLogger(t)
	dir := t.ArtifactDir()
	const teststr = "Hello, World!"
	wantMD5 := md5.Sum([]byte(teststr))

	ctx := t.Context()
	bucket := aws.NewTestBlobStore(dir)
	be.NilError(bucket.WriteFile(ctx, "hello.txt", nil, []byte(teststr)))

	hash, size := be.OK2(bucket.ReadMD5(ctx, "hello.txt"))
	be.
		SlicesEqual(hash, wantMD5[:]).
		Equal(size, int64(len(teststr)))

	be.NilError(os.Remove(filepath.Join(dir, "hello.txt.attrs")))

	hash, size = be.OK2(bucket.ReadMD5(ctx, "hello.txt"))
	be.
		SlicesEqual(hash, wantMD5[:]).
		EqualLength(teststr, int(size))
}
