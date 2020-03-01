// +build relic

package signature

import (
	"fmt"

	"github.com/dapperlabs/flow-go/crypto"
	model "github.com/dapperlabs/flow-go/model/hotstuff"
	"github.com/dapperlabs/flow-go/model/messages"
)

// RandomBeaconSigVerifier verifies signatures generated by the random beacon.
// Specifically, it verifies individual key shares and reconstructed threshold signatures.
type RandomBeaconSigVerifier struct {
	// the hasher for signer random beacon signature
	randomBeaconHasher crypto.Hasher
}

// NewRandomBeaconSigVerifier constructs a new RandomBeaconSigVerifier
func NewRandomBeaconSigVerifier() RandomBeaconSigVerifier {
	// The random beacon is only run by consensus nodes. Hence, the tag used for identifying the vote can be hard-coded.
	return RandomBeaconSigVerifier{
		randomBeaconHasher: crypto.NewBLS_KMAC(messages.RandomBeaconTag),
	}
}

// VerifyRandomBeaconSig verifies a single random beacon signature share for a block using the using signer's public key
// sig - the signature to be verified
// block - the block that the signature was signed for.
// randomBeaconSignerIndex - the signer index of signer's random beacon key share.
//
// Note: we are specifically choosing safety over performance here.
//   * The vote itself contains all the information for verifying the signature: the blockID and the block's view
//   * We could use the vote to verify that the signature is valid for the information contained in the vote's message
//   * However, for security, we are explicitly verifying that the vote matches the full block.
//     We do this by converting the block to the byte-sequence which we expect an honest voter to have signed
//     and then check the provided signature against this self-computed byte-sequence.
func (s *RandomBeaconSigVerifier) VerifyRandomBeaconSig(sig crypto.Signature, block *model.Block, signerPubKey crypto.PublicKey) (bool, error) {
	msg := BlockToBytesForSign(block)
	valid, err := signerPubKey.Verify(sig, msg, s.randomBeaconHasher)
	if err != nil {
		return false, fmt.Errorf("cannot verify random beacon signature: %w", err)
	}
	return valid, nil
}

// VerifyAggregatedRandomBeaconSignature verifies a random beacon threshold signature for a block
// sig - the signature to be verified
// block - the block that the signature was signed for.
//
// Note: we are specifically choosing safety over performance here.
//   * The vote itself contains all the information for verifying the signature: the blockID and the block's view
//   * We could use the vote to verify that the signature is valid for the information contained in the vote's message
//   * However, for security, we are explicitly verifying that the vote matches the full block.
//     We do this by converting the block to the byte-sequence which we expect an honest voter to have signed
//     and then check the provided signature against this self-computed byte-sequence.
func (s *RandomBeaconSigVerifier) VerifyAggregatedRandomBeaconSignature(sig crypto.Signature, block *model.Block, groupPubKey crypto.PublicKey) (bool, error) {
	msg := BlockToBytesForSign(block)
	// the reconstructed signature is also a BLS signature which can be verified by the group public key
	valid, err := groupPubKey.Verify(sig, msg, s.randomBeaconHasher)
	if err != nil {
		return false, fmt.Errorf("cannot verify reconstructed random beacon sig: %w", err)
	}
	return valid, nil
}
