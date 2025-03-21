package audit

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/hex"
	"encoding/xml"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/imroc/req/v3"
)

type COS struct {
	secretId  string
	secretKey string
	bucket    string
}

func NewCOS(secretId, secretKey, bucket string) *COS {
	return &COS{
		secretId:  secretId,
		secretKey: secretKey,
		bucket:    bucket,
	}
}

// Check 检查图片是否违规 true: 违规 false: 未违规
func (c *COS) Check(url string) (bool, string, error) {
	authorization, err := c.getAuthorization("GET", "/", 0)
	if err != nil {
		return false, "", err
	}

	client := req.C()
	resp, err := client.R().SetQueryParams(map[string]string{
		"ci-process": "sensitive-content-recognition",
		"detect-url": url,
	}).SetHeader("Authorization", authorization).Get("https://" + c.bucket + "/")
	if err != nil {
		return false, "", err
	}
	if !resp.IsSuccessState() {
		return false, "", fmt.Errorf("cos audit failed: %s", resp.String())
	}

	type checkResponse struct {
		XMLName  xml.Name `xml:"RecognitionResult"`
		JobId    string   `xml:"JobId"`
		Result   int      `xml:"Result"`
		Label    string   `xml:"Label"`
		SubLabel string   `xml:"SubLabel"`
		Score    int      `xml:"Score"`
	}

	var response checkResponse
	err = xml.Unmarshal(resp.Bytes(), &response)
	if err != nil {
		return false, "", fmt.Errorf("cos audit response unmarshal failed: %s", err)
	}

	if response.Result == 1 {
		return true, fmt.Sprintf("%d-%s(%s)", response.Score, response.Label, response.SubLabel), nil
	}

	return false, "", nil
}

func (c *COS) getAuthorization(method, path string, expires time.Duration) (string, error) {
	if expires <= 0 {
		expires = 30 * time.Minute
	}
	signTimeStart := time.Now().Add(-time.Minute).Unix()
	signTimeEnd := time.Now().Add(expires).Unix()
	signTime := fmt.Sprintf("%d;%d", signTimeStart, signTimeEnd)

	pathUnescaped, err := url.PathUnescape(path)
	if err != nil {
		return "", err
	}

	httpString := strings.ToLower(method) + "\n" + pathUnescaped + "\n\n\n"
	hasher := sha1.New()
	hasher.Write([]byte(httpString))
	sha1edHttpString := hex.EncodeToString(hasher.Sum(nil))
	stringToSign := "sha1\n" + signTime + "\n" + sha1edHttpString + "\n"

	h := hmac.New(sha1.New, []byte(c.secretKey))
	h.Write([]byte(signTime))
	signKey := hex.EncodeToString(h.Sum(nil))

	h2 := hmac.New(sha1.New, []byte(signKey))
	h2.Write([]byte(stringToSign))
	signature := hex.EncodeToString(h2.Sum(nil))

	return fmt.Sprintf("q-sign-algorithm=sha1&q-ak=%s&q-sign-time=%s&q-key-time=%s&q-header-list=&q-url-param-list=&q-signature=%s",
		c.secretId, signTime, signTime, signature), nil
}
