// +build ignore

package linodego

import (
	"encoding/json"
	"fmt"

	"github.com/go-resty/resty"
)

// Template represents a Template object
type Template struct {
	ID int
	// UpdatedStr string `json:"updated"`
	// Updated *time.Time `json:"-"`
}

// TemplatesPagedResponse represents a paginated Template API response
type TemplatesPagedResponse struct {
	*PageOptions
	Data []*Template
}

// endpoint gets the endpoint URL for Template
func (TemplatesPagedResponse) endpoint(c *Client) string {
	endpoint, err := c.Templates.Endpoint()
	if err != nil {
		panic(err)
	}
	return endpoint
}

// appendData appends Templates when processing paginated Template responses
func (resp *TemplatesPagedResponse) appendData(r *TemplatesPagedResponse) {
	(*resp).Data = append(resp.Data, r.Data...)
}

// setResult sets the Resty response type of Template
func (TemplatesPagedResponse) setResult(r *resty.Request) {
	r.SetResult(TemplatesPagedResponse{})
}

// ListTemplates lists Templates
func (c *Client) ListTemplates(opts *ListOptions) ([]*Template, error) {
	response := TemplatesPagedResponse{}
	err := c.listHelper(&response, opts)
	for _, el := range response.Data {
		el.fixDates()
	}
	if err != nil {
		return nil, err
	}
	return response.Data, nil
}

// fixDates converts JSON timestamps to Go time.Time values
func (v *Template) fixDates() *Template {
	// v.Created, _ = parseDates(v.CreatedStr)
	// v.Updated, _ = parseDates(v.UpdatedStr)
	return v
}

// GetTemplate gets the template with the provided ID
func (c *Client) GetTemplate(id string) (*Template, error) {
	e, err := c.Templates.Endpoint()
	if err != nil {
		return nil, err
	}
	e = fmt.Sprintf("%s/%s", e, id)
	r, err := coupleAPIErrors(c.R().SetResult(&Template{}).Get(e))
	if err != nil {
		return nil, err
	}
	return r.Result().(*Template).fixDates(), nil
}

// CreateTemplate creates a Template
func (c *Client) CreateTemplate(Template *TemplateCreateOptions) (*Template, error) {
	var body string
	e, err := c.Templates.Endpoint()
	if err != nil {
		return nil, err
	}

	req := c.R().SetResult(&Template{})

	if bodyData, err := json.Marshal(template); err == nil {
		body = string(bodyData)
	} else {
		return nil, NewError(err)
	}

	r, err := coupleAPIErrors(req.
		SetHeader("Content-Type", "application/json").
		SetBody(body).
		Post(e))

	if err != nil {
		return nil, err
	}
	return r.Result().(*Template).fixDates(), nil
}

// UpdateTemplate updates the Template with the specified id
func (c *Client) UpdateTemplate(id int, updateOpts TemplateUpdateOptions) (*Template, error) {
	var body string
	e, err := c.Templates.Endpoint()
	if err != nil {
		return nil, err
	}
	e = fmt.Sprintf("%s/%d", e, id)

	req := c.R().SetResult(&Template{})

	if bodyData, err := json.Marshal(template); err == nil {
		body = string(bodyData)
	} else {
		return nil, NewError(err)
	}

	r, err := coupleAPIErrors(req.
		SetBody(body).
		Put(e))

	if err != nil {
		return nil, err
	}
	return r.Result().(*Template).fixDates(), nil
}

// DeleteTemplate deletes the Template with the specified id
func (c *Client) DeleteTemplate(id int) error {
	e, err := c.Templates.Endpoint()
	if err != nil {
		return err
	}
	e = fmt.Sprintf("%s/%d", e, id)

	if _, err := coupleAPIErrors(c.R().Delete(e)); err != nil {
		return err
	}

	return nil
}
