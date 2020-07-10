package common

import (
	"encoding/binary"
	"fmt"

	"github.com/dapperlabs/flow-go/ledger"
	"github.com/dapperlabs/flow-go/ledger/utils"
)

// EncodingDecodingVersion encoder/decoder code only supports
// decoding data with version smaller or equal to this value
// bumping this number prevents older versions of the code
// to deal with the newer version of data
// codes should be always updated with backward compatibility
const EncodingDecodingVersion = uint64(0)

// EncodingType capture the type of encoded entity
type EncodingType uint16

const (
	// EncodingTypeStateCommitment - encoding type for StateCommitments
	EncodingTypeStateCommitment = iota
	// EncodingTypeKeyPart - encoding type for KeyParts (a subset of key)
	EncodingTypeKeyPart
	// EncodingTypeKey - encoding type for Keys (unique identifier to reference a location in ledger)
	EncodingTypeKey
	// EncodingTypeValue - encoding type for Ledger Values
	EncodingTypeValue
	// EncodingTypePath - encoding type for Paths (trie storage location of a key value pair)
	EncodingTypePath
	// EncodingTypePayload - encoding type for Payloads (stored at trie nodes including key value pair )
	EncodingTypePayload
	// EncodingTypeProof encoding type for Proofs
	// (all data needed to verify a key value pair at specific stateCommitment)
	EncodingTypeProof
	// EncodingTypeBatchProof - encoding type for BatchProofs
	EncodingTypeBatchProof
	// encodingTypeUnknown - unknown encoding type - Warning this should always be the last item in the list
	encodingTypeUnknown
)

func (e EncodingType) String() string {
	return [...]string{"StateCommitment", "KeyPart", "Key", "Value", "Path", "Payload", "Proof", "BatchProof", "Unknown"}[e]
}

// checkEncDecVer extracts encoding bytes from a raw encoded message
// checks it against the supported versions and returns the rest of rawInput (excluding encDecVersion bytes)
func checkEncDecVer(rawInput []byte) (rest []byte, version uint64, err error) {
	version, rest, err = readUint64(rawInput)
	if err != nil {
		return rest, version, fmt.Errorf("error checking the encoding decoding version: %w", err)
	}
	// error on versions coming from future till a time-machine is invented
	if version > EncodingDecodingVersion {
		return rest, version, fmt.Errorf("incompatible encoding decoding version (%d > %d): %w", version, EncodingDecodingVersion, err)
	}
	// return the rest of bytes
	return rest, version, nil
}

// checkEncodingType extracts encoding byte from a raw encoded message
// checks it against the supported versions and returns the rest of rawInput (excluding encDecVersion bytes)
func checkEncodingType(rawInput []byte, expectedType uint16) (rest []byte, err error) {
	t, r, err := readUint16(rawInput)
	if err != nil {
		return r, fmt.Errorf("error checking type of the encoded entity: %w", err)
	}

	// error if type is known for this code
	if t >= encodingTypeUnknown {
		return r, fmt.Errorf("unknown entity type in the encoded data (%d > %d)", t, encodingTypeUnknown)
	}

	// error if type is known for this code
	if t != expectedType {
		return r, fmt.Errorf("unexpected entity type, got (%v) but (%v) was expected", EncodingType(t), EncodingType(expectedType))
	}

	// return the rest of bytes
	return r, nil
}

// EncodeKeyPart encodes a key part into a byte slice
func EncodeKeyPart(kp *ledger.KeyPart) []byte {
	// EncodingDecodingType
	buffer := appendUint64([]byte{}, EncodingDecodingVersion)

	// encode key part entity type
	buffer = appendUint16(buffer, EncodingTypeKeyPart)

	// encode the key part content
	buffer = append(buffer, encodeKeyPart(kp)...)
	return buffer
}

func encodeKeyPart(kp *ledger.KeyPart) []byte {
	buffer := make([]byte, 0)

	// encode "Type" field of the key part
	buffer = appendUint16(buffer, kp.Type)

	// encode "Value" field of the key part
	buffer = append(buffer, kp.Value...)
	return buffer
}

// DecodeKeyPart constructs a key part from an encoded key part
func DecodeKeyPart(encodedKeyPart []byte) (*ledger.KeyPart, error) {
	// currently we ignore the version but in the future we
	// can do switch case based on the version if needed
	rest, _, err := checkEncDecVer(encodedKeyPart)
	if err != nil {
		return nil, fmt.Errorf("error decoding key part: %w", err)
	}

	// check the encoding type
	rest, err = checkEncodingType(rest, EncodingTypeKeyPart)
	if err != nil {
		return nil, fmt.Errorf("error decoding key part: %w", err)
	}

	// decode the key part content
	key, err := decodeKeyPart(rest)
	if err != nil {
		return nil, fmt.Errorf("error decoding key part: %w", err)
	}

	return key, nil
}

func decodeKeyPart(inp []byte) (*ledger.KeyPart, error) {
	// read key part type and the rest is the key item part
	kpt, kpv, err := readUint16(inp)
	if err != nil {
		return nil, fmt.Errorf("error decoding key part (content): %w", err)
	}
	return &ledger.KeyPart{Type: kpt, Value: kpv}, nil
}

// EncodeKey encodes a key into a byte slice
func EncodeKey(k *ledger.Key) []byte {

	// encode EncodingDecodingType
	buffer := appendUint64([]byte{}, EncodingDecodingVersion)
	// encode key entity type
	buffer = appendUint16(buffer, EncodingTypeKey)
	// encode key content
	buffer = append(buffer, encodeKey(k)...)

	return buffer
}

// encodeKey encodes a key into a byte slice
func encodeKey(k *ledger.Key) []byte {
	buffer := make([]byte, 0)
	// encode number of key parts
	buffer = appendUint16(buffer, uint16(len(k.KeyParts)))
	// iterate over key parts
	for _, kp := range k.KeyParts {
		// encode the key part
		encKP := encodeKeyPart(&kp)
		// encode the len of the encoded key part
		buffer = appendUint64(buffer, uint64(len(encKP)))
		// append the encoded key part
		buffer = append(buffer, encKP...)
	}
	return buffer
}

// DecodeKey constructs a key from an encoded key part
func DecodeKey(encodedKey []byte) (*ledger.Key, error) {
	// check the enc dec version
	rest, _, err := checkEncDecVer(encodedKey)
	if err != nil {
		return nil, fmt.Errorf("error decoding key: %w", err)
	}
	// check the encoding type
	rest, err = checkEncodingType(rest, EncodingTypeKey)
	if err != nil {
		return nil, fmt.Errorf("error decoding key: %w", err)
	}

	// decode the key content
	key, err := decodeKey(rest)
	if err != nil {
		return nil, fmt.Errorf("error decoding key: %w", err)
	}
	return key, nil
}

func decodeKey(inp []byte) (*ledger.Key, error) {
	key := &ledger.Key{}
	numOfParts, rest, err := readUint16(inp)
	if err != nil {
		return nil, fmt.Errorf("error decoding key (content): %w", err)
	}

	for i := 0; i < int(numOfParts); i++ {
		var kpEncSize uint64
		var kpEnc []byte
		// read encoded key part size
		kpEncSize, rest, err = readUint64(rest)
		if err != nil {
			return nil, fmt.Errorf("error decoding key (content): %w", err)
		}

		// read encoded key part
		kpEnc, rest, err = readSlice(rest, int(kpEncSize))
		if err != nil {
			return nil, fmt.Errorf("error decoding key (content): %w", err)
		}

		// decode encoded key part
		kp, err := decodeKeyPart(kpEnc)
		if err != nil {
			return nil, fmt.Errorf("error decoding key (content): %w", err)
		}
		key.KeyParts = append(key.KeyParts, *kp)
	}
	return key, nil
}

// EncodeValue encodes a value into a byte slice
func EncodeValue(v ledger.Value) []byte {
	// encode EncodingDecodingType
	buffer := appendUint64([]byte{}, EncodingDecodingVersion)

	// encode key entity type
	buffer = appendUint16(buffer, EncodingTypeValue)

	// encode value
	buffer = append(buffer, encodeValue(v)...)

	return buffer
}

func encodeValue(v ledger.Value) []byte {
	return v
}

// DecodeValue constructs a ledger value using an encoded byte slice
func DecodeValue(encodedValue []byte) (ledger.Value, error) {
	// check enc dec version
	rest, _, err := checkEncDecVer(encodedValue)
	if err != nil {
		return nil, err
	}

	// check the encoding type
	rest, err = checkEncodingType(rest, EncodingTypeValue)
	if err != nil {
		return nil, err
	}

	return decodeValue(rest)
}

func decodeValue(inp []byte) (ledger.Value, error) {
	return ledger.Value(inp), nil
}

// EncodePayload encodes a ledger payload
func EncodePayload(p *ledger.Payload) []byte {
	// encode EncodingDecodingType
	buffer := appendUint64([]byte{}, EncodingDecodingVersion)

	// encode key entity type
	buffer = appendUint16(buffer, EncodingTypePayload)

	// append encoded payload content
	buffer = append(buffer, encodePayload(p)...)

	return buffer
}

func encodePayload(p *ledger.Payload) []byte {
	buffer := make([]byte, 0)

	// encode key
	encK := encodeKey(&p.Key)

	// encode encoded key size
	buffer = appendUint64(buffer, uint64(len(encK)))

	// append encoded key content
	buffer = append(buffer, encK...)

	// encode value
	encV := encodeValue(p.Value)

	// encode encoded value size
	buffer = appendUint64(buffer, uint64(len(encV)))

	// append encoded key content
	buffer = append(buffer, encV...)

	return buffer
}

// DecodePayload construct a payload from an encoded byte slice
func DecodePayload(encodedPayload []byte) (*ledger.Payload, error) {
	// check the enc dec version
	rest, _, err := checkEncDecVer(encodedPayload)
	if err != nil {
		return nil, fmt.Errorf("error decoding payload: %w", err)
	}
	// check the encoding type
	rest, err = checkEncodingType(rest, EncodingTypePayload)
	if err != nil {
		return nil, fmt.Errorf("error decoding payload: %w", err)
	}
	return decodePayload(rest)
}

func decodePayload(inp []byte) (*ledger.Payload, error) {

	// read encoded key size
	encKeySize, rest, err := readUint64(inp)
	if err != nil {
		return nil, fmt.Errorf("error decoding payload: %w", err)
	}

	// read encoded key
	encKey, rest, err := readSlice(rest, int(encKeySize))
	if err != nil {
		return nil, fmt.Errorf("error decoding payload: %w", err)
	}

	// decode the key
	key, err := decodeKey(encKey)
	if err != nil {
		return nil, fmt.Errorf("error decoding payload: %w", err)
	}

	// read encoded value size
	encValeSize, rest, err := readUint64(rest)
	if err != nil {
		return nil, fmt.Errorf("error decoding payload: %w", err)
	}

	// read encoded value
	encValue, rest, err := readSlice(rest, int(encValeSize))
	if err != nil {
		return nil, fmt.Errorf("error decoding payload: %w", err)
	}

	// decode value
	value, err := decodeValue(encValue)
	if err != nil {
		return nil, fmt.Errorf("error decoding payload: %w", err)
	}

	return &ledger.Payload{Key: *key, Value: value}, nil
}

// EncodeProof encodes the content of a proof into a byte slice
func EncodeProof(p *ledger.Proof) []byte {
	// encode EncodingDecodingType
	buffer := appendUint64([]byte{}, EncodingDecodingVersion)

	// encode key entity type
	buffer = appendUint16(buffer, EncodingTypeProof)

	// append encoded proof content
	buffer = append(buffer, encodeProof(p)...)

	return buffer
}

func encodeProof(p *ledger.Proof) []byte {
	// first byte is reserved for inclusion flag
	buffer := make([]byte, 1)
	if p.Inclusion {
		// set the first bit to 1 if it is an inclusion proof
		buffer[0] |= 1 << 7
	}

	// steps are encoded as a single byte
	buffer = appendUint8(buffer, p.Steps)

	// include flags size and content
	buffer = appendUint8(buffer, uint8(len(p.Flags)))
	buffer = append(buffer, p.Flags...)

	// include path size and content
	buffer = appendUint16(buffer, uint16(p.Path.Size()))
	buffer = append(buffer, p.Path...)

	// include encoded payload size and content
	encPayload := encodePayload(&p.Payload)
	buffer = appendUint64(buffer, uint64(len(encPayload)))
	buffer = append(buffer, encPayload...)

	// and finally include all interims (hash values)
	// number of interims
	buffer = appendUint8(buffer, uint8(len(p.Interims)))
	for _, inter := range p.Interims {
		buffer = appendUint16(buffer, uint16(len(inter)))
		buffer = append(buffer, inter...)
	}

	return buffer
}

// DecodeProof construct a proof from an encoded byte slice
func DecodeProof(encodedProof []byte) (*ledger.Proof, error) {
	// check the enc dec version
	rest, _, err := checkEncDecVer(encodedProof)
	if err != nil {
		return nil, fmt.Errorf("error decoding proof: %w", err)
	}
	// check the encoding type
	rest, err = checkEncodingType(rest, EncodingTypeProof)
	if err != nil {
		return nil, fmt.Errorf("error decoding proof: %w", err)
	}
	return decodeProof(rest)
}

func decodeProof(inp []byte) (*ledger.Proof, error) {
	pInst := ledger.NewProof()

	// Inclusion flag
	byteInclusion, rest, err := readSlice(inp, 1)
	pInst.Inclusion = utils.IsBitSet(byteInclusion, 0)

	// read steps
	steps, rest, err := readUint8(rest)
	pInst.Steps = steps

	// read flags
	flagsSize, rest, err := readUint8(rest)
	if err != nil {
		return nil, fmt.Errorf("error decoding proof: %w", err)
	}
	flags, rest, err := readSlice(rest, int(flagsSize))
	if err != nil {
		return nil, fmt.Errorf("error decoding proof: %w", err)
	}
	pInst.Flags = flags

	// read path
	pathSize, rest, err := readUint16(rest)
	if err != nil {
		return nil, fmt.Errorf("error decoding proof: %w", err)
	}
	path, rest, err := readSlice(rest, int(pathSize))
	if err != nil {
		return nil, fmt.Errorf("error decoding proof: %w", err)
	}
	pInst.Path = path

	// read payload
	encPayloadSize, rest, err := readUint64(rest)
	if err != nil {
		return nil, fmt.Errorf("error decoding proof: %w", err)
	}
	encPayload, rest, err := readSlice(rest, int(encPayloadSize))
	if err != nil {
		return nil, fmt.Errorf("error decoding proof: %w", err)
	}
	payload, err := decodePayload(encPayload)
	if err != nil {
		return nil, fmt.Errorf("error decoding proof: %w", err)
	}
	pInst.Payload = *payload

	// read interims
	interimsLen, rest, err := readUint8(rest)
	if err != nil {
		return nil, fmt.Errorf("error decoding proof: %w", err)
	}
	interims := make([][]byte, 0)

	for i := 0; i < int(interimsLen); i++ {
		interimSize, rest, err := readUint16(rest)
		if err != nil {
			return nil, fmt.Errorf("error decoding proof: %w", err)
		}

		interim, rest, err := readSlice(rest, int(interimSize))
		if err != nil {
			return nil, fmt.Errorf("error decoding proof: %w", err)
		}
		interims = append(interims, interim)
	}
	pInst.Interims = interims

	return pInst, nil
}

// // EncodeBatchProof encodes all the content of a batch proof into a slice of byte slices and total len
// // TODO change this to only an slice of bytes
// func EncodeBatchProof(bp *ledger.BatchProof) ([][]byte, int) {
// 	proofs := make([][]byte, 0)
// 	totalLength := 0
// 	// for each proof we create a byte array
// 	for _, p := range bp.Proofs {
// 		proof := EncodeProof(p)
// 		totalLength += len(proof)
// 		proofs = append(proofs, proof)
// 	}
// 	return proofs, totalLength
// }

// // DecodeBatchProof takes in an encodes array of byte arrays an converts them into a BatchProof
// // TODO create proof decoder
// func DecodeBatchProof(proofs [][]byte) (*ledger.BatchProof, error) {
// 	bp := ledger.NewBatchProof()
// 	// The decode logic is as follows:
// 	// The first byte in the array is the inclusion flag, with the first bit set as the inclusion (1 = inclusion, 0 = non-inclusion)
// 	// The second byte is size, needs to be converted to uint8
// 	// The next 32 bytes are the flag
// 	// Each subsequent 32 bytes are the proofs needed for the verifier
// 	// Each result is put into their own array and put into a BatchProof
// 	for _, proof := range proofs {
// 		if len(proof) < 4 {
// 			return nil, fmt.Errorf("error decoding the proof: proof size too small")
// 		}
// 		pInst := ledger.NewProof()
// 		byteInclusion := proof[0:1]
// 		pInst.Inclusion = utils.IsBitSet(byteInclusion, 0)
// 		step := proof[1:2]
// 		pInst.Steps = step[0]
// 		flagSize := int(proof[2])
// 		if flagSize < 1 {
// 			return nil, fmt.Errorf("error decoding the proof: flag size should be greater than 0")
// 		}
// 		pInst.Flags = proof[3 : flagSize+3]
// 		byteProofs := make([][]byte, 0)
// 		for i := flagSize + 3; i < len(proof); i += 32 {
// 			// TODO understand the logic here
// 			if i+32 <= len(proof) {
// 				byteProofs = append(byteProofs, proof[i:i+32])
// 			} else {
// 				byteProofs = append(byteProofs, proof[i:])
// 			}
// 		}
// 		pInst.Interims = byteProofs
// 		bp.AppendProof(pInst)
// 	}
// 	return bp, nil
// }

// TODO RAMTIN FIX proof encoder decoder

func readSlice(input []byte, size int) (value []byte, rest []byte, err error) {
	if len(input) < size {
		return nil, input, fmt.Errorf("input size is too small to be splited %d < %d ", len(input), size)
	}
	return input[:size], input[size:], nil
}

func readUint8(input []byte) (value uint8, rest []byte, err error) {
	if len(input) < 1 {
		return 0, input, fmt.Errorf("input size (%d) is too small to read a uint8", len(input))
	}
	return uint8(input[0]), input[1:], nil
}

func readUint16(input []byte) (value uint16, rest []byte, err error) {
	if len(input) < 2 {
		return 0, input, fmt.Errorf("input size (%d) is too small to read a uint16", len(input))
	}
	return binary.BigEndian.Uint16(input[:2]), input[2:], nil
}

func readUint32(input []byte) (value uint32, rest []byte, err error) {
	if len(input) < 4 {
		return 0, input, fmt.Errorf("input size (%d) is too small to read a uint32", len(input))
	}
	return binary.BigEndian.Uint32(input[:4]), input[4:], nil
}

func readUint64(input []byte) (value uint64, rest []byte, err error) {
	if len(input) < 8 {
		return 0, input, fmt.Errorf("input size (%d) is too small to read a uint64", len(input))
	}
	return binary.BigEndian.Uint64(input[:8]), input[8:], nil
}

func writeUint8(buffer []byte, loc int, value uint8) int {
	buffer[loc] = byte(value)
	return loc + 1
}

func writeUint16(buffer []byte, loc int, value uint16) int {
	binary.BigEndian.PutUint16(buffer[loc:], value)
	return loc + 2
}

func writeUint32(buffer []byte, loc int, value uint32) int {
	binary.BigEndian.PutUint32(buffer[loc:], value)
	return loc + 4
}

func writeUint64(buffer []byte, loc int, value uint64) int {
	binary.BigEndian.PutUint64(buffer[loc:], value)
	return loc + 8
}

func appendUint8(input []byte, value uint8) []byte {
	return append(input, byte(value))
}

func appendUint16(input []byte, value uint16) []byte {
	buffer := make([]byte, 2)
	binary.BigEndian.PutUint16(buffer, value)
	return append(input, buffer...)
}

func appendUint32(input []byte, value uint32) []byte {
	buffer := make([]byte, 4)
	binary.BigEndian.PutUint32(buffer, value)
	return append(input, buffer...)
}

func appendUint64(input []byte, value uint64) []byte {
	buffer := make([]byte, 8)
	binary.BigEndian.PutUint64(buffer, value)
	return append(input, buffer...)
}