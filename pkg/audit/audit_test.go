package audit

import (
	"testing"

	"github.com/libtnb/assert/check"
	"github.com/libtnb/assert/must"
)

func TestNew(t *testing.T) {
	a, err := New(Config{Driver: "aliyun", Aliyun: AliyunConfig{AccessKeyID: "ak", AccessKeySecret: "sk"}})
	must.NoError(t, err)
	ali, ok := a.driver.(*Aliyun)
	must.True(t, ok)
	check.Equal(t, ali.accessKeyId, "ak")
	check.Equal(t, ali.accessKeySecret, "sk")

	a, err = New(Config{Driver: "cos", COS: COSConfig{SecretID: "id", SecretKey: "key", Bucket: "b"}})
	must.NoError(t, err)
	cos, ok := a.driver.(*COS)
	must.True(t, ok)
	check.Equal(t, cos.bucket, "b")

	_, err = New(Config{Driver: "nope"})
	must.ErrorContains(t, err, `unsupported driver "nope"`)

	_, err = New(Config{})
	must.Error(t, err)
}
