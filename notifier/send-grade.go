package notifier

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func (c *Client) SendGrade(ctx context.Context, grade *ApiGrade) error {
	body, err := json.Marshal(grade)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(
		"POST", fmt.Sprintf("%s/grades", c.BaseUrl),
		bytes.NewBuffer(body),
	)
	if err != nil {
		return err
	}

	req = req.WithContext(ctx)
	return c.sendRequest(req, nil)
}
