package key

import (
	"github.com/stellar/go/exp/crypto/derivation"
	"github.com/stellar/go/keypair"
	"github.com/stellar/go/strkey"
	"github.com/stellar/go/xdr"
)

const PrimaryAccountIndex uint32 = 0

// DeriveAccountPrimary derives the primary account key
// from from the root key determined by seed.
func DeriveAccountPrimary(seed []byte) *keypair.Full {
	return DeriveAccount(seed, PrimaryAccountIndex)
}

// DeriveAccount derives the account key at index i
// from from the root key determined by seed.
//
// Only "hardened" derivation is supported.
// For convenience, indexes below derivation.FirstHardenedIndex
// are converted to hardened indexes.
func DeriveAccount(seed []byte, i uint32) *keypair.Full {
	// These path elements are from StellarAccountPathFormat
	// in package github.com/stellar/go/exp/crypto/derivation.
	return derive(seed, 44, 148, i)
}

// derive derives the child key at path
// from from the root key determined by seed.
//
// Only "hardened" derivation is supported.
// For convenience, indexes below derivation.FirstHardenedIndex
// are allowed, and converted to hardened indexes.
func derive(seed []byte, path ...uint32) *keypair.Full {
	key, err := derivation.NewMasterKey(seed)
	if err != nil {
		// TODO(kr): find a way to have package derivation do this
		panic(err) // can't happen (error from Hash.Write)
	}

	for _, i := range path {
		if i < derivation.FirstHardenedIndex {
			i += derivation.FirstHardenedIndex
		}
		key, err = key.Derive(i)
		if err != nil {
			panic(err) // can't happen (error from Hash.Write or non-hardened index)
		}
	}

	kp, err := keypair.FromRawSeed(key.RawSeed())
	if err != nil {
		panic(err) // can't happen (error from bytes.Buffer or strkey version byte)
	}

	return kp
}

// PublicKeyXDR returns the public-key part of k in XDR format.
func PublicKeyXDR(kp *keypair.Full) xdr.PublicKey {
	b := strkey.MustDecode(strkey.VersionByteAccountID, kp.Address())
	var u256 xdr.Uint256
	copy(u256[:], b)
	pk, err := xdr.NewPublicKey(xdr.PublicKeyTypePublicKeyTypeEd25519, u256)
	if err != nil {
		panic(err) // mustn't happen
	}
	return pk
}
