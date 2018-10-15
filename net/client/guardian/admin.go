package guardian

import (
	"context"

	"github.com/interstellar/starlight/enclave/guardian"
)

type CreateClusterRequest = guardian.CreateClusterRequest
type CreateClusterResponse = guardian.CreateClusterResponse

func (c *Client) CreateCluster(ctx context.Context, req CreateClusterRequest) (*CreateClusterResponse, error) {
	var resp CreateClusterResponse
	err := c.post(ctx, "/admin/create-cluster", req, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}
