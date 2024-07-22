package magicpen

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
)

type StabilityAIResult struct {
	Image        string `json:"image"`
	FinishReason string `json:"finish_reason"`
	Seed         int64  `json:"seed"`
}

func (mp *MagicPen) draw(cmd *DrawCommand) (res *StabilityAIResult, err error) {
	var b bytes.Buffer
	w := multipart.NewWriter(&b)
	w.WriteField("prompt", cmd.Prompt)
	w.WriteField("aspect_ratio", cmd.Ratio)
	w.WriteField("model", cmd.Model)
	w.WriteField("output_format", cmd.Output)
	w.WriteField("negative_prompt", cmd.Negative)
	w.Close()
	req, _ := http.NewRequest(http.MethodPost, "https://api.stability.ai/v2beta/stable-image/generate/sd3", &b)
	req.Header.Set("content-type", w.FormDataContentType())
	req.Header.Set("accept", "application/json")
	req.Header.Set("authorization", fmt.Sprintf("Bearer %v", mp.conf.Sk))
	resp, err := mp.httpCli.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	res = &StabilityAIResult{}
	err = json.Unmarshal(data, res)
	if err != nil {
		return nil, err
	}
	return
}
