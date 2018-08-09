package linodego

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"time"

	"github.com/go-resty/resty"
)

/*
 * https://developers.linode.com/v4/reference/endpoints/linode/instances
 */

type InstanceStatus string

// InstanceStatus enum represents potential Instance.Status values
const (
	InstanceBooting      InstanceStatus = "booting"
	InstanceRunning      InstanceStatus = "running"
	InstanceOffline      InstanceStatus = "offline"
	InstanceShuttingDown InstanceStatus = "shutting_down"
	InstanceRebooting    InstanceStatus = "rebooting"
	InstanceProvisioning InstanceStatus = "provisioning"
	InstanceDeleting     InstanceStatus = "deleting"
	InstanceMigrating    InstanceStatus = "migrating"
	InstanceRebuilding   InstanceStatus = "rebuilding"
	InstanceCloning      InstanceStatus = "cloning"
	InstanceRestoring    InstanceStatus = "restoring"
)

// Instance represents a linode object
type Instance struct {
	CreatedStr string `json:"created"`
	UpdatedStr string `json:"updated"`

	ID         int
	Created    *time.Time `json:"-"`
	Updated    *time.Time `json:"-"`
	Region     string
	Alerts     *InstanceAlert
	Backups    *InstanceBackup
	Image      string
	Group      string
	IPv4       []*net.IP
	IPv6       string
	Label      string
	Type       string
	Status     InstanceStatus
	Hypervisor string
	Specs      *InstanceSpec
}

// InstanceSpec represents a linode spec
type InstanceSpec struct {
	Disk     int
	Memory   int
	VCPUs    int
	Transfer int
}

// InstanceAlert represents a metric alert
type InstanceAlert struct {
	CPU           int `json:"cpu"`
	IO            int `json:"io"`
	NetworkIn     int `json:"network_in"`
	NetworkOut    int `json:"network_out"`
	TransferQuote int `json:"transfer_queue"`
}

// InstanceBackup represents backup settings for an instance
type InstanceBackup struct {
	Enabled  bool `json:"enabled"`
	Schedule struct {
		Day    string `json:"day,omitempty"`
		Window string `json:"window,omitempty"`
	}
}

// InstanceCreateOptions require only Region and Type
type InstanceCreateOptions struct {
	Region          string            `json:"region"`
	Type            string            `json:"type"`
	Label           string            `json:"label,omitempty"`
	Group           string            `json:"group,omitempty"`
	RootPass        string            `json:"root_pass,omitempty"`
	AuthorizedKeys  []string          `json:"authorized_keys,omitempty"`
	StackScriptID   int               `json:"stackscript_id,omitempty"`
	StackScriptData map[string]string `json:"stackscript_data,omitempty"`
	BackupID        int               `json:"backup_id,omitempty"`
	Image           string            `json:"image,omitempty"`
	BackupsEnabled  bool              `json:"backups_enabled,omitempty"`
	SwapSize        *int              `json:"swap_size,omitempty"`
	Booted          bool              `json:"booted,omitempty"`
}

// InstanceUpdateOptions is an options struct used when Updating an Instance
type InstanceUpdateOptions struct {
	Label   string         `json:"label,omitempty"`
	Group   string         `json:"group,omitempty"`
	Backups InstanceBackup `json:"backups,omitempty"`
	Alerts  InstanceAlert  `json:"alerts,omitempty"`
}

// InstanceCloneOptions is an options struct when sending a clone request to the API
type InstanceCloneOptions struct {
	Region         string   `json:"region"`
	Type           string   `json:"type"`
	LinodeID       int      `json:"linode_id"`
	Label          string   `json:"label"`
	Group          string   `json:"group"`
	BackupsEnabled bool     `json:"backups_enabled"`
	Disks          []string `json:"disks"`
	Configs        []string `json:"configs"`
}

func (l *Instance) fixDates() *Instance {
	l.Created, _ = parseDates(l.CreatedStr)
	l.Updated, _ = parseDates(l.UpdatedStr)
	return l
}

// InstancesPagedResponse represents a linode API response for listing
type InstancesPagedResponse struct {
	*PageOptions
	Data []*Instance
}

// endpoint gets the endpoint URL for Instance
func (InstancesPagedResponse) endpoint(c *Client) string {
	endpoint, err := c.Instances.Endpoint()
	if err != nil {
		panic(err)
	}
	return endpoint
}

// appendData appends Instances when processing paginated Instance responses
func (resp *InstancesPagedResponse) appendData(r *InstancesPagedResponse) {
	(*resp).Data = append(resp.Data, r.Data...)
}

// setResult sets the Resty response type of Instance
func (InstancesPagedResponse) setResult(r *resty.Request) {
	r.SetResult(InstancesPagedResponse{})
}

// ListInstances lists linode instances
func (c *Client) ListInstances(ctx context.Context, opts *ListOptions) ([]*Instance, error) {
	response := InstancesPagedResponse{}
	err := c.listHelper(ctx, &response, opts)
	for _, el := range response.Data {
		el.fixDates()
	}
	if err != nil {
		return nil, err
	}
	return response.Data, nil
}

// GetInstance gets the instance with the provided ID
func (c *Client) GetInstance(ctx context.Context, linodeID int) (*Instance, error) {
	e, err := c.Instances.Endpoint()
	if err != nil {
		return nil, err
	}
	e = fmt.Sprintf("%s/%d", e, linodeID)
	r, err := coupleAPIErrors(c.R(ctx).
		SetResult(Instance{}).
		Get(e))
	if err != nil {
		return nil, err
	}
	return r.Result().(*Instance).fixDates(), nil
}

// CreateInstance creates a Linode instance
func (c *Client) CreateInstance(ctx context.Context, instance *InstanceCreateOptions) (*Instance, error) {
	var body string
	e, err := c.Instances.Endpoint()
	if err != nil {
		return nil, err
	}

	req := c.R(ctx).SetResult(&Instance{})

	if bodyData, err := json.Marshal(instance); err == nil {
		body = string(bodyData)
	} else {
		return nil, NewError(err)
	}

	r, err := coupleAPIErrors(req.
		SetBody(body).
		Post(e))

	if err != nil {
		return nil, err
	}
	return r.Result().(*Instance).fixDates(), nil
}

// UpdateInstance creates a Linode instance
func (c *Client) UpdateInstance(ctx context.Context, id int, instance *InstanceUpdateOptions) (*Instance, error) {
	var body string
	e, err := c.Instances.Endpoint()
	if err != nil {
		return nil, err
	}
	e = fmt.Sprintf("%s/%d", e, id)

	req := c.R(ctx).SetResult(&Instance{})

	if bodyData, err := json.Marshal(instance); err == nil {
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
	return r.Result().(*Instance).fixDates(), nil
}

// RenameInstance renames an Instance
func (c *Client) RenameInstance(ctx context.Context, linodeID int, label string) (*Instance, error) {
	return c.UpdateInstance(ctx, linodeID, &InstanceUpdateOptions{Label: label})
}

// DeleteInstance deletes a Linode instance
func (c *Client) DeleteInstance(ctx context.Context, id int) error {
	e, err := c.Instances.Endpoint()
	if err != nil {
		return err
	}
	e = fmt.Sprintf("%s/%d", e, id)

	if _, err := coupleAPIErrors(c.R(ctx).Delete(e)); err != nil {
		return err
	}

	return nil
}

// BootInstance will boot a Linode instance
// A configID of 0 will cause Linode to choose the last/best config
func (c *Client) BootInstance(ctx context.Context, id int, configID int) (bool, error) {
	bodyStr := ""

	if configID != 0 {
		bodyMap := map[string]int{"config_id": configID}
		bodyJSON, err := json.Marshal(bodyMap)
		if err != nil {
			return false, NewError(err)
		}
		bodyStr = string(bodyJSON)
	}

	e, err := c.Instances.Endpoint()
	if err != nil {
		return false, err
	}

	e = fmt.Sprintf("%s/%d/boot", e, id)
	r, err := coupleAPIErrors(c.R(ctx).
		SetBody(bodyStr).
		Post(e))

	return settleBoolResponseOrError(r, err)
}

// CloneInstance clones a Linode instance
func (c *Client) CloneInstance(ctx context.Context, id int, options *InstanceCloneOptions) (*Instance, error) {
	var body string
	e, err := c.Instances.Endpoint()
	if err != nil {
		return nil, err
	}
	e = fmt.Sprintf("%s/%d/clone", e, id)

	req := c.R(ctx).SetResult(&Instance{})

	if bodyData, err := json.Marshal(options); err == nil {
		body = string(bodyData)
	} else {
		return nil, NewError(err)
	}

	r, err := coupleAPIErrors(req.
		SetBody(body).
		Post(e))

	if err != nil {
		return nil, err
	}

	return r.Result().(*Instance).fixDates(), nil
}

// RebootInstance reboots a Linode instance
// A configID of 0 will cause Linode to choose the last/best config
func (c *Client) RebootInstance(ctx context.Context, id int, configID int) (bool, error) {
	bodyStr := "{}"

	if configID != 0 {
		bodyMap := map[string]int{"config_id": configID}
		bodyJSON, err := json.Marshal(bodyMap)
		if err != nil {
			return false, NewError(err)
		}
		bodyStr = string(bodyJSON)
	}

	e, err := c.Instances.Endpoint()
	if err != nil {
		return false, err
	}

	e = fmt.Sprintf("%s/%d/reboot", e, id)

	r, err := coupleAPIErrors(c.R(ctx).
		SetBody(bodyStr).
		Post(e))

	return settleBoolResponseOrError(r, err)
}

// MutateInstance Upgrades a Linode to its next generation.
func (c *Client) MutateInstance(ctx context.Context, id int) (bool, error) {
	e, err := c.Instances.Endpoint()
	if err != nil {
		return false, err
	}
	e = fmt.Sprintf("%s/%d/mutate", e, id)

	r, err := coupleAPIErrors(c.R(ctx).Post(e))
	return settleBoolResponseOrError(r, err)
}

// RebuildInstanceOptions is a struct representing the options to send to the rebuild linode endpoint
type RebuildInstanceOptions struct {
	Image           string            `json:"image"`
	RootPass        string            `json:"root_pass"`
	AuthorizedKeys  []string          `json:"authorized_keys"`
	StackscriptID   int               `json:"stackscript_id"`
	StackscriptData map[string]string `json:"stackscript_data"`
	Booted          bool              `json:"booted"`
}

// RebuildInstance Deletes all Disks and Configs on this Linode,
// then deploys a new Image to this Linode with the given attributes.
func (c *Client) RebuildInstance(ctx context.Context, id int, opts *RebuildInstanceOptions) (*Instance, error) {
	o, err := json.Marshal(opts)
	if err != nil {
		return nil, NewError(err)
	}
	b := string(o)
	e, err := c.Instances.Endpoint()
	if err != nil {
		return nil, err
	}
	e = fmt.Sprintf("%s/%d/rebuild", e, id)
	r, err := coupleAPIErrors(c.R(ctx).
		SetBody(b).
		SetResult(&Instance{}).
		Post(e))
	if err != nil {
		return nil, err
	}
	return r.Result().(*Instance).fixDates(), nil
}

// ResizeInstance resizes an instance to new Linode type
func (c *Client) ResizeInstance(ctx context.Context, id int, linodeType string) (bool, error) {
	body := fmt.Sprintf("{\"type\":\"%s\"}", linodeType)

	e, err := c.Instances.Endpoint()
	if err != nil {
		return false, err
	}
	e = fmt.Sprintf("%s/%d/resize", e, id)

	r, err := coupleAPIErrors(c.R(ctx).
		SetBody(body).
		Post(e))

	return settleBoolResponseOrError(r, err)
}

// ShutdownInstance - Shutdown an instance
func (c *Client) ShutdownInstance(ctx context.Context, id int) (bool, error) {
	e, err := c.Instances.Endpoint()
	if err != nil {
		return false, err
	}
	e = fmt.Sprintf("%s/%d/shutdown", e, id)
	return settleBoolResponseOrError(coupleAPIErrors(c.R(ctx).Post(e)))
}

func settleBoolResponseOrError(resp *resty.Response, err error) (bool, error) {
	if err != nil {
		return false, err
	}
	return true, nil
}
