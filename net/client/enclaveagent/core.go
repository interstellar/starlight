package enclaveagent

import (
	"context"

	chainjson "github.com/interstellar/starlight/encoding/json"
	"github.com/interstellar/starlight/errors"
)

type deviceIdentityRequest struct {
	Nonce chainjson.HexBytes `json:"nonce"`
}

type DeviceIdentity struct {
	DeviceIdentity chainjson.HexBytes `json:"device_identity"`
}

type Cluster struct {
	SealedClusterRecord []byte
	ClusterCertificate  []byte
}

type CreateClusterRequest struct {
	EscrowPredicate       chainjson.HexBytes   `json:"escrow_predicate"`
	ExtraDeviceIdentities []chainjson.HexBytes `json:"extra_device_identities"`
}

type createClusterResponse struct {
	SealedClusterRecord chainjson.HexBytes   `json:"sealed_cluster_record"`
	ClusterCertificate  chainjson.HexBytes   `json:"cluster_certificate"`
	ExportedRecords     []chainjson.HexBytes `json:"exported_records"`
}

type ExtendClusterExportRequest struct {
	ExtensionRequest    chainjson.HexBytes   `json:"extension_request"`
	SealedClusterRecord chainjson.HexBytes   `json:"sealed_cluster_record"`
	ExportedRecords     []chainjson.HexBytes `json:"exported_records"`
}

type extendClusterExportResponse struct {
	ExportedRecords []chainjson.HexBytes `json:"exported_records"`
}

type extendClusterImportRequest struct {
	ExportedRecord chainjson.HexBytes `json:"exported_record"`
}

type extendClusterImportResponse struct {
	ClusterCertificate  chainjson.HexBytes `json:"cluster_certificate"`
	SealedClusterRecord chainjson.HexBytes `json:"sealed_cluster_record"`
}

type MigrateClusterExportRequest struct {
	MigrationRequest          chainjson.HexBytes `json:"migration_request"`
	SealedSourceClusterRecord chainjson.HexBytes `json:"sealed_source_cluster_record"`
}

type migrateClusterExportResponse struct {
	Record chainjson.HexBytes `json:"record"`
}

type MigrateClusterImportRequest struct {
	MigratedClusterRecord     chainjson.HexBytes `json:"migrated_cluster_record"`
	SealedTargetClusterRecord chainjson.HexBytes `json:"sealed_target_cluster_record"`
}

type migrateClusterImportResponse struct {
	SealedClusterRecord chainjson.HexBytes `json:"sealed_cluster_record"`
	ClusterCertificate  chainjson.HexBytes `json:"cluster_certificate"`
}

func (c *Client) DeviceIdentity(ctx context.Context, nonce [32]byte) (*DeviceIdentity, error) {
	var identity DeviceIdentity
	err := c.post(ctx, "/core/device-identity", deviceIdentityRequest{
		Nonce: nonce[:],
	}, &identity)
	if err != nil {
		return nil, errors.Wrap(err, "querying device identity")
	}
	return &identity, nil
}

func (c *Client) CreateCluster(ctx context.Context, req CreateClusterRequest) (*Cluster, []chainjson.HexBytes, error) {
	var resp createClusterResponse
	err := c.post(ctx, "/core/create-cluster", &req, &resp)
	if err != nil {
		return nil, nil, errors.Wrap(err, "creating a cluster")
	}
	cluster := &Cluster{
		SealedClusterRecord: resp.SealedClusterRecord,
		ClusterCertificate:  resp.ClusterCertificate,
	}
	return cluster, resp.ExportedRecords, nil
}

func (c *Client) ExtendClusterExport(ctx context.Context, req ExtendClusterExportRequest) ([]chainjson.HexBytes, error) {
	var resp extendClusterExportResponse
	err := c.post(ctx, "/core/extend-cluster/export", &req, &resp)
	if err != nil {
		return nil, errors.Wrap(err, "exporting for cluster extension")
	}
	return resp.ExportedRecords, nil
}

func (c *Client) ExtendClusterImport(ctx context.Context, exportedRecord []byte) (*Cluster, error) {
	var resp extendClusterImportResponse
	err := c.post(ctx, "/core/extend-cluster/import", extendClusterImportRequest{
		ExportedRecord: exportedRecord,
	}, &resp)
	if err != nil {
		return nil, errors.Wrap(err, "importing cluster extension")
	}
	return &Cluster{
		SealedClusterRecord: resp.SealedClusterRecord,
		ClusterCertificate:  resp.ClusterCertificate,
	}, nil
}

func (c *Client) MigrateClusterExport(ctx context.Context, req MigrateClusterExportRequest) (record []byte, err error) {
	var resp migrateClusterExportResponse
	err = c.post(ctx, "/core/migrate-cluster/export", &req, &resp)
	if err != nil {
		return nil, errors.Wrap(err, "exporting cluster migration")
	}
	return resp.Record, nil
}

func (c *Client) MigrateClusterImport(ctx context.Context, req MigrateClusterImportRequest) (*Cluster, error) {
	var resp migrateClusterImportResponse
	err := c.post(ctx, "/core/migrate-cluster/import", &req, &resp)
	if err != nil {
		return nil, errors.Wrap(err, "importing cluster migration")
	}
	return &Cluster{
		SealedClusterRecord: resp.SealedClusterRecord,
		ClusterCertificate:  resp.ClusterCertificate,
	}, nil
}
