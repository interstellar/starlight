package enclaveagent

import (
	"context"

	i10rjson "github.com/interstellar/starlight/encoding/json"
	"github.com/interstellar/starlight/errors"
)

type CreateManagedKeyRequest struct {
	ControlPredicate     i10rjson.HexBytes `json:"control_predicate"`
	RecoveryPredicate    i10rjson.HexBytes `json:"recovery_predicate"`
	CredentialLifetimeMS uint64            `json:"credential_lifetime_ms"`
	SealedClusterRecord  i10rjson.HexBytes `json:"sealed_cluster_record"`
}

type createManagedKeyResponse struct {
	Record i10rjson.HexBytes `json:"record"`
}

type RecoverManagedKeyRequest struct {
	RecoveryRequest        i10rjson.HexBytes `json:"recovery_request"`
	SealedManagedKeyRecord i10rjson.HexBytes `json:"sealed_managed_key_record"`
	SealedClusterRecord    i10rjson.HexBytes `json:"sealed_cluster_record"`
}

type IssueCredentialRequest struct {
	Params     i10rjson.HexBytes `json:"params"`
	Signatures i10rjson.HexBytes `json:"signatures"`

	SealedManagedKeyRecord i10rjson.HexBytes `json:"sealed_managed_key_record"`
	SealedClusterRecord    i10rjson.HexBytes `json:"sealed_cluster_record"`
}

type issueCredentialResponse struct {
	Credential i10rjson.HexBytes `json:"credential"`
}

type recoverManagedKeyResponse struct {
	Record i10rjson.HexBytes `json:"record"`
}

type SignTransactionRequest struct {
	SignRequest            i10rjson.HexBytes `json:"sign_request"`
	Signature              i10rjson.HexBytes `json:"signature"`
	Credential             i10rjson.HexBytes `json:"credential"`
	SealedManagedKeyRecord i10rjson.HexBytes `json:"sealed_managed_key_record"`
	SealedClusterRecord    i10rjson.HexBytes `json:"sealed_cluster_record"`
}

type signTransactionResponse struct {
	Signature i10rjson.HexBytes `json:"signature"`
}

func (c *Client) CreateManagedKey(ctx context.Context, req CreateManagedKeyRequest) ([]byte, error) {
	var resp createManagedKeyResponse
	err := c.post(ctx, "/stellar/create-managed-key", req, &resp)
	if err != nil {
		return nil, errors.Wrap(err, "creating managed key")
	}
	return resp.Record, nil
}

func (c *Client) RecoverManagedKey(ctx context.Context, req RecoverManagedKeyRequest) ([]byte, error) {
	var resp recoverManagedKeyResponse
	err := c.post(ctx, "/stellar/recover-managed-key", req, &resp)
	if err != nil {
		return nil, errors.Wrap(err, "recovering managed key")
	}
	return resp.Record, nil
}

func (c *Client) IssueCredential(ctx context.Context, req IssueCredentialRequest) (i10rjson.HexBytes, error) {
	var resp issueCredentialResponse
	err := c.post(ctx, "/stellar/issue-credential", req, &resp)
	if err != nil {
		return nil, errors.Wrap(err, "issuing credential")
	}
	return resp.Credential, nil
}

func (c *Client) SignTransaction(ctx context.Context, req SignTransactionRequest) ([]byte, error) {
	var resp signTransactionResponse
	err := c.post(ctx, "/stellar/sign-transaction", req, &resp)
	if err != nil {
		return nil, errors.Wrap(err, "signing transaction")
	}
	return resp.Signature, nil
}
