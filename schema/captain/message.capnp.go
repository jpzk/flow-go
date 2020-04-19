// Code generated by capnpc-go. DO NOT EDIT.

package captain

import (
	strconv "strconv"
	capnp "zombiezen.com/go/capnproto2"
	text "zombiezen.com/go/capnproto2/encoding/text"
	schemas "zombiezen.com/go/capnproto2/schemas"
)

type Message struct{ capnp.Struct }
type Message_Which uint16

const (
	Message_Which_auth                Message_Which = 0
	Message_Which_ping                Message_Which = 1
	Message_Which_pong                Message_Which = 2
	Message_Which_announce            Message_Which = 3
	Message_Which_request             Message_Which = 4
	Message_Which_response            Message_Which = 5
	Message_Which_collectionGuarantee Message_Which = 6
	Message_Which_blockProposal       Message_Which = 7
	Message_Which_blockVote           Message_Which = 8
	Message_Which_blockCommit         Message_Which = 9
)

func (w Message_Which) String() string {
	const s = "authpingpongannouncerequestresponsecollectionGuaranteeblockProposalblockVoteblockCommit"
	switch w {
	case Message_Which_auth:
		return s[0:4]
	case Message_Which_ping:
		return s[4:8]
	case Message_Which_pong:
		return s[8:12]
	case Message_Which_announce:
		return s[12:20]
	case Message_Which_request:
		return s[20:27]
	case Message_Which_response:
		return s[27:35]
	case Message_Which_collectionGuarantee:
		return s[35:54]
	case Message_Which_blockProposal:
		return s[54:67]
	case Message_Which_blockVote:
		return s[67:76]
	case Message_Which_blockCommit:
		return s[76:87]

	}
	return "Message_Which(" + strconv.FormatUint(uint64(w), 10) + ")"
}

// Message_TypeID is the unique identifier for the type Message.
const Message_TypeID = 0xaf7a8da44e30bf62

func NewMessage(s *capnp.Segment) (Message, error) {
	st, err := capnp.NewStruct(s, capnp.ObjectSize{DataSize: 8, PointerCount: 1})
	return Message{st}, err
}

func NewRootMessage(s *capnp.Segment) (Message, error) {
	st, err := capnp.NewRootStruct(s, capnp.ObjectSize{DataSize: 8, PointerCount: 1})
	return Message{st}, err
}

func ReadRootMessage(msg *capnp.Message) (Message, error) {
	root, err := msg.RootPtr()
	return Message{root.Struct()}, err
}

func (s Message) String() string {
	str, _ := text.Marshal(0xaf7a8da44e30bf62, s.Struct)
	return str
}

func (s Message) Which() Message_Which {
	return Message_Which(s.Struct.Uint16(0))
}
func (s Message) Auth() (Auth, error) {
	if s.Struct.Uint16(0) != 0 {
		panic("Which() != auth")
	}
	p, err := s.Struct.Ptr(0)
	return Auth{Struct: p.Struct()}, err
}

func (s Message) HasAuth() bool {
	if s.Struct.Uint16(0) != 0 {
		return false
	}
	p, err := s.Struct.Ptr(0)
	return p.IsValid() || err != nil
}

func (s Message) SetAuth(v Auth) error {
	s.Struct.SetUint16(0, 0)
	return s.Struct.SetPtr(0, v.Struct.ToPtr())
}

// NewAuth sets the auth field to a newly
// allocated Auth struct, preferring placement in s's segment.
func (s Message) NewAuth() (Auth, error) {
	s.Struct.SetUint16(0, 0)
	ss, err := NewAuth(s.Struct.Segment())
	if err != nil {
		return Auth{}, err
	}
	err = s.Struct.SetPtr(0, ss.Struct.ToPtr())
	return ss, err
}

func (s Message) Ping() (Ping, error) {
	if s.Struct.Uint16(0) != 1 {
		panic("Which() != ping")
	}
	p, err := s.Struct.Ptr(0)
	return Ping{Struct: p.Struct()}, err
}

func (s Message) HasPing() bool {
	if s.Struct.Uint16(0) != 1 {
		return false
	}
	p, err := s.Struct.Ptr(0)
	return p.IsValid() || err != nil
}

func (s Message) SetPing(v Ping) error {
	s.Struct.SetUint16(0, 1)
	return s.Struct.SetPtr(0, v.Struct.ToPtr())
}

// NewPing sets the ping field to a newly
// allocated Ping struct, preferring placement in s's segment.
func (s Message) NewPing() (Ping, error) {
	s.Struct.SetUint16(0, 1)
	ss, err := NewPing(s.Struct.Segment())
	if err != nil {
		return Ping{}, err
	}
	err = s.Struct.SetPtr(0, ss.Struct.ToPtr())
	return ss, err
}

func (s Message) Pong() (Pong, error) {
	if s.Struct.Uint16(0) != 2 {
		panic("Which() != pong")
	}
	p, err := s.Struct.Ptr(0)
	return Pong{Struct: p.Struct()}, err
}

func (s Message) HasPong() bool {
	if s.Struct.Uint16(0) != 2 {
		return false
	}
	p, err := s.Struct.Ptr(0)
	return p.IsValid() || err != nil
}

func (s Message) SetPong(v Pong) error {
	s.Struct.SetUint16(0, 2)
	return s.Struct.SetPtr(0, v.Struct.ToPtr())
}

// NewPong sets the pong field to a newly
// allocated Pong struct, preferring placement in s's segment.
func (s Message) NewPong() (Pong, error) {
	s.Struct.SetUint16(0, 2)
	ss, err := NewPong(s.Struct.Segment())
	if err != nil {
		return Pong{}, err
	}
	err = s.Struct.SetPtr(0, ss.Struct.ToPtr())
	return ss, err
}

func (s Message) Announce() (Announce, error) {
	if s.Struct.Uint16(0) != 3 {
		panic("Which() != announce")
	}
	p, err := s.Struct.Ptr(0)
	return Announce{Struct: p.Struct()}, err
}

func (s Message) HasAnnounce() bool {
	if s.Struct.Uint16(0) != 3 {
		return false
	}
	p, err := s.Struct.Ptr(0)
	return p.IsValid() || err != nil
}

func (s Message) SetAnnounce(v Announce) error {
	s.Struct.SetUint16(0, 3)
	return s.Struct.SetPtr(0, v.Struct.ToPtr())
}

// NewAnnounce sets the announce field to a newly
// allocated Announce struct, preferring placement in s's segment.
func (s Message) NewAnnounce() (Announce, error) {
	s.Struct.SetUint16(0, 3)
	ss, err := NewAnnounce(s.Struct.Segment())
	if err != nil {
		return Announce{}, err
	}
	err = s.Struct.SetPtr(0, ss.Struct.ToPtr())
	return ss, err
}

func (s Message) Request() (Request, error) {
	if s.Struct.Uint16(0) != 4 {
		panic("Which() != request")
	}
	p, err := s.Struct.Ptr(0)
	return Request{Struct: p.Struct()}, err
}

func (s Message) HasRequest() bool {
	if s.Struct.Uint16(0) != 4 {
		return false
	}
	p, err := s.Struct.Ptr(0)
	return p.IsValid() || err != nil
}

func (s Message) SetRequest(v Request) error {
	s.Struct.SetUint16(0, 4)
	return s.Struct.SetPtr(0, v.Struct.ToPtr())
}

// NewRequest sets the request field to a newly
// allocated Request struct, preferring placement in s's segment.
func (s Message) NewRequest() (Request, error) {
	s.Struct.SetUint16(0, 4)
	ss, err := NewRequest(s.Struct.Segment())
	if err != nil {
		return Request{}, err
	}
	err = s.Struct.SetPtr(0, ss.Struct.ToPtr())
	return ss, err
}

func (s Message) Response() (Response, error) {
	if s.Struct.Uint16(0) != 5 {
		panic("Which() != response")
	}
	p, err := s.Struct.Ptr(0)
	return Response{Struct: p.Struct()}, err
}

func (s Message) HasResponse() bool {
	if s.Struct.Uint16(0) != 5 {
		return false
	}
	p, err := s.Struct.Ptr(0)
	return p.IsValid() || err != nil
}

func (s Message) SetResponse(v Response) error {
	s.Struct.SetUint16(0, 5)
	return s.Struct.SetPtr(0, v.Struct.ToPtr())
}

// NewResponse sets the response field to a newly
// allocated Response struct, preferring placement in s's segment.
func (s Message) NewResponse() (Response, error) {
	s.Struct.SetUint16(0, 5)
	ss, err := NewResponse(s.Struct.Segment())
	if err != nil {
		return Response{}, err
	}
	err = s.Struct.SetPtr(0, ss.Struct.ToPtr())
	return ss, err
}

func (s Message) CollectionGuarantee() (CollectionGuarantee, error) {
	if s.Struct.Uint16(0) != 6 {
		panic("Which() != collectionGuarantee")
	}
	p, err := s.Struct.Ptr(0)
	return CollectionGuarantee{Struct: p.Struct()}, err
}

func (s Message) HasCollectionGuarantee() bool {
	if s.Struct.Uint16(0) != 6 {
		return false
	}
	p, err := s.Struct.Ptr(0)
	return p.IsValid() || err != nil
}

func (s Message) SetCollectionGuarantee(v CollectionGuarantee) error {
	s.Struct.SetUint16(0, 6)
	return s.Struct.SetPtr(0, v.Struct.ToPtr())
}

// NewCollectionGuarantee sets the collectionGuarantee field to a newly
// allocated CollectionGuarantee struct, preferring placement in s's segment.
func (s Message) NewCollectionGuarantee() (CollectionGuarantee, error) {
	s.Struct.SetUint16(0, 6)
	ss, err := NewCollectionGuarantee(s.Struct.Segment())
	if err != nil {
		return CollectionGuarantee{}, err
	}
	err = s.Struct.SetPtr(0, ss.Struct.ToPtr())
	return ss, err
}

func (s Message) BlockProposal() (BlockProposal, error) {
	if s.Struct.Uint16(0) != 7 {
		panic("Which() != blockProposal")
	}
	p, err := s.Struct.Ptr(0)
	return BlockProposal{Struct: p.Struct()}, err
}

func (s Message) HasBlockProposal() bool {
	if s.Struct.Uint16(0) != 7 {
		return false
	}
	p, err := s.Struct.Ptr(0)
	return p.IsValid() || err != nil
}

func (s Message) SetBlockProposal(v BlockProposal) error {
	s.Struct.SetUint16(0, 7)
	return s.Struct.SetPtr(0, v.Struct.ToPtr())
}

// NewBlockProposal sets the blockProposal field to a newly
// allocated BlockProposal struct, preferring placement in s's segment.
func (s Message) NewBlockProposal() (BlockProposal, error) {
	s.Struct.SetUint16(0, 7)
	ss, err := NewBlockProposal(s.Struct.Segment())
	if err != nil {
		return BlockProposal{}, err
	}
	err = s.Struct.SetPtr(0, ss.Struct.ToPtr())
	return ss, err
}

func (s Message) BlockVote() (BlockVote, error) {
	if s.Struct.Uint16(0) != 8 {
		panic("Which() != blockVote")
	}
	p, err := s.Struct.Ptr(0)
	return BlockVote{Struct: p.Struct()}, err
}

func (s Message) HasBlockVote() bool {
	if s.Struct.Uint16(0) != 8 {
		return false
	}
	p, err := s.Struct.Ptr(0)
	return p.IsValid() || err != nil
}

func (s Message) SetBlockVote(v BlockVote) error {
	s.Struct.SetUint16(0, 8)
	return s.Struct.SetPtr(0, v.Struct.ToPtr())
}

// NewBlockVote sets the blockVote field to a newly
// allocated BlockVote struct, preferring placement in s's segment.
func (s Message) NewBlockVote() (BlockVote, error) {
	s.Struct.SetUint16(0, 8)
	ss, err := NewBlockVote(s.Struct.Segment())
	if err != nil {
		return BlockVote{}, err
	}
	err = s.Struct.SetPtr(0, ss.Struct.ToPtr())
	return ss, err
}

func (s Message) BlockCommit() (BlockCommit, error) {
	if s.Struct.Uint16(0) != 9 {
		panic("Which() != blockCommit")
	}
	p, err := s.Struct.Ptr(0)
	return BlockCommit{Struct: p.Struct()}, err
}

func (s Message) HasBlockCommit() bool {
	if s.Struct.Uint16(0) != 9 {
		return false
	}
	p, err := s.Struct.Ptr(0)
	return p.IsValid() || err != nil
}

func (s Message) SetBlockCommit(v BlockCommit) error {
	s.Struct.SetUint16(0, 9)
	return s.Struct.SetPtr(0, v.Struct.ToPtr())
}

// NewBlockCommit sets the blockCommit field to a newly
// allocated BlockCommit struct, preferring placement in s's segment.
func (s Message) NewBlockCommit() (BlockCommit, error) {
	s.Struct.SetUint16(0, 9)
	ss, err := NewBlockCommit(s.Struct.Segment())
	if err != nil {
		return BlockCommit{}, err
	}
	err = s.Struct.SetPtr(0, ss.Struct.ToPtr())
	return ss, err
}

// Message_List is a list of Message.
type Message_List struct{ capnp.List }

// NewMessage creates a new list of Message.
func NewMessage_List(s *capnp.Segment, sz int32) (Message_List, error) {
	l, err := capnp.NewCompositeList(s, capnp.ObjectSize{DataSize: 8, PointerCount: 1}, sz)
	return Message_List{l}, err
}

func (s Message_List) At(i int) Message { return Message{s.List.Struct(i)} }

func (s Message_List) Set(i int, v Message) error { return s.List.SetStruct(i, v.Struct) }

func (s Message_List) String() string {
	str, _ := text.MarshalList(0xaf7a8da44e30bf62, s.List)
	return str
}

// Message_Promise is a wrapper for a Message promised by a client call.
type Message_Promise struct{ *capnp.Pipeline }

func (p Message_Promise) Struct() (Message, error) {
	s, err := p.Pipeline.Struct()
	return Message{s}, err
}

func (p Message_Promise) Auth() Auth_Promise {
	return Auth_Promise{Pipeline: p.Pipeline.GetPipeline(0)}
}

func (p Message_Promise) Ping() Ping_Promise {
	return Ping_Promise{Pipeline: p.Pipeline.GetPipeline(0)}
}

func (p Message_Promise) Pong() Pong_Promise {
	return Pong_Promise{Pipeline: p.Pipeline.GetPipeline(0)}
}

func (p Message_Promise) Announce() Announce_Promise {
	return Announce_Promise{Pipeline: p.Pipeline.GetPipeline(0)}
}

func (p Message_Promise) Request() Request_Promise {
	return Request_Promise{Pipeline: p.Pipeline.GetPipeline(0)}
}

func (p Message_Promise) Response() Response_Promise {
	return Response_Promise{Pipeline: p.Pipeline.GetPipeline(0)}
}

func (p Message_Promise) CollectionGuarantee() CollectionGuarantee_Promise {
	return CollectionGuarantee_Promise{Pipeline: p.Pipeline.GetPipeline(0)}
}

func (p Message_Promise) BlockProposal() BlockProposal_Promise {
	return BlockProposal_Promise{Pipeline: p.Pipeline.GetPipeline(0)}
}

func (p Message_Promise) BlockVote() BlockVote_Promise {
	return BlockVote_Promise{Pipeline: p.Pipeline.GetPipeline(0)}
}

func (p Message_Promise) BlockCommit() BlockCommit_Promise {
	return BlockCommit_Promise{Pipeline: p.Pipeline.GetPipeline(0)}
}

const schema_cc8ede639915bf22 = "x\xdaT\xceMh\xd3`\x1c\xc7\xf1\xdf?I\x9bU" +
	"6\x9a\x91\x0c*R6\xc5U\x1dc\x9b\xcc\xe1;\x14" +
	"u*\xe2\xa45\"\x1e\x04I\xc3\xc3V\xda&\xb1I" +
	"/\x82\x08^\x8b\xa2\x87!\xeaN2\xf1$E\xf0$" +
	"H\xe9A\x04_\x0e\xde\xc4\x83\xce\x83o\xe0A\x10\x9c" +
	"N\xe7_\x9e\xd2\x8a\x1e\x9f\xdf\xe7\xff\x85g\xe2\x06e" +
	"\x95\xad\xb1k:\x90?\x15\x8bs\xa19ql\xf1\xd2" +
	"\xb9\x06\xf2i\"\xde\xd0\x1c\xb8\xee\xbe\xbe\xfc\x0c\xd3\xa4" +
	"\xaf\x01&\xef\xa9#d\xb6T\x1d\x98|\xa8nW\xf0" +
	"\x91CwNT\x9cqWu\x82\xc8)z\xe3\x15\x11" +
	"\x86\xce\xac\x18s\x9d\xc0\x0bv\xcd\x880)\x9f9\xa2" +
	"\xfcfU\xebe\xd6\x080\x134\x02\xd8\x1a\xa9d\x1b" +
	"\xa4P\x1f\xfdf\x8b$\xf4\xb5\xa1G\x82%AYe" +
	"\x8b\x14\xc0\xecoC\xaf\x84\x94\x04\xf5\x17[\xa4\x02\xe6" +
	"\x00\x1d\x01lK\xc2\x90\x04\xed'[\xa4\x01f\x9a\xf6" +
	"\x01vJ\xc2F\x09\xb1\x15\xb6(\x06\x98\xeb\xdb\xc5\x90" +
	"\x84Q\x09\xf1\x1flQ\x1c0\xb7\xd0-\xc0\x1e\x95\xb0" +
	"C\x82\xfe\x9d-\xd2\x01s\x8a\xaa\x80\xbdMBVB" +
	"\xcf2[\xd4\x03\x98{\xe98`\xef\x91pXB\xe2" +
	"\x1b[\x94\x00\xcci*\x00\xf6\x01\x099R(\xe9\xd4" +
	"\xa292xS\xe3\xa9\xfea\xf9\xf9K\x80\xc8\x00%" +
	"\x83\xa27K\x06_|\xb2V\xcb\x9e\xd9]\xff;\xfb" +
	"\xed\xb9?\xf35\xc1\x14oufv<\xcf\xafy\xae" +
	"\x00@\x06\xa7\x1b\xf5V]\x94^u\xf4BU\x9c\xad" +
	"\x890\"\x83?\xbd=:\xf3\xa5\x99y\xd7\xed\xaa\"" +
	"\x0c|/\xect\xeb\x96\xde<\x9e\x1a\xf6\x97\xba\xea\xfa" +
	"\xe5\xb2p\xa3\"\xf9\xde\xa1\x9aSu<=\x12\x82\x0c" +
	"\x1e\x9e\x7f\xbfR\x1a\\x\xd4\xbd+\x94}\xb7\x94\xab" +
	"\xfa\x18\x0c\xfc\xd0)\x93\xc1cw\xee\x7f>x\xba\xb4" +
	"\xf0\xdf\xc5I?\x02\x092V\x17S\x0f\xce\xcf\xdf\xbd" +
	"\xf2\xaf\xed\xf7+\xd0+E\xf9\xc7\x9b\x99\x9d'^\\" +
	"\xa5\xdb\x9d\xf6O\x00\x00\x00\xff\xff\xe6\xb0\xa4a"

func init() {
	schemas.Register(schema_cc8ede639915bf22,
		0xaf7a8da44e30bf62)
}