package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image"
	"image/png"
	"io"
	"mime/multipart"
	"net/http"

	t "github.com/ekzyis/snappy/types"
)

func (c *Client) UploadImage(img *image.RGBA) (string, error) {
	var (
		b      = img.Bounds()
		width  = b.Dx()
		height = b.Dy()
		type_  = "image/png"
		size   int
	)

	var imgBuf bytes.Buffer
	if err := png.Encode(&imgBuf, img); err != nil {
		return "", err
	}
	size = imgBuf.Len()

	// get signed URL for S3 upload
	body := t.GqlBody{
		Query: `
		mutation getSignedPOST($type: String!, $size: Int!, $width: Int!, $height: Int!, $avatar: Boolean) {
			getSignedPOST(type: $type, size: $size, width: $width, height: $height, avatar: $avatar) {
				url
				fields
			}
		}`,
		Variables: map[string]interface{}{
			"type":   type_,
			"size":   size,
			"width":  width,
			"height": height,
			"avatar": false,
		},
	}

	resp, err := c.callApi(body)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var respBody t.GetSignedPOSTResponse
	err = json.NewDecoder(resp.Body).Decode(&respBody)
	if err != nil {
		err = fmt.Errorf("error decoding getSignedPOST: %w", err)
		return "", err
	}

	err = c.checkForErrors(respBody.Errors)
	if err != nil {
		return "", err
	}

	s3Url := respBody.Data.GetSignedPOST.Url
	fields := respBody.Data.GetSignedPOST.Fields

	// create multipart form
	var (
		buf bytes.Buffer
		w   = multipart.NewWriter(&buf)
		fw  io.Writer
	)

	for k, v := range fields {
		if fw, err = w.CreateFormField(k); err != nil {
			return "", err
		}
		fw.Write([]byte(v))
	}

	for k, v := range map[string]string{
		"Content-Type":  type_,
		"Cache-Control": "max-age=31536000",
		"acl":           "public-read",
	} {
		if fw, err = w.CreateFormField(k); err != nil {
			return "", err
		}
		fw.Write([]byte(v))
	}

	if fw, err = w.CreateFormFile("file", "image.png"); err != nil {
		return "", err
	}
	fw.Write(imgBuf.Bytes())

	if err = w.Close(); err != nil {
		return "", err
	}

	// upload to S3
	var req *http.Request
	if req, err = http.NewRequest("POST", s3Url, &buf); err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", w.FormDataContentType())

	client := http.DefaultClient
	if resp, err = client.Do(req); err != nil {
		return "", err
	}
	defer resp.Body.Close()

	imgId := respBody.Data.GetSignedPOST.Fields["key"]
	imgUrl := fmt.Sprintf("%s/%s", c.MediaUrl, imgId)

	return imgUrl, nil
}
