package greenlight

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/encoding"
	"google.golang.org/protobuf/encoding/protowire"
	"google.golang.org/protobuf/proto"

	"github.com/getAlby/hub/events"
	"github.com/getAlby/hub/lnclient"
	"github.com/getAlby/hub/logger"
	"github.com/sirupsen/logrus"
)

// streamIncoming subscribes to the greenlight.Node StreamIncoming RPC, which
// delivers incoming off-chain payments: invoice payments AND spontaneous
// payments (keysend). We only publish keysends here (offchain.bolt11 == "")
// because invoice payments are already covered by the WaitAnyInvoice pump;
// publishing both would create duplicate transactions in the hub.
//
// The greenlight.Node service has no Go bindings in the hub, so the stream
// is opened with a raw codec and the messages are decoded with protowire.
// The wire shape is stable: package greenlight, service Node, method
// StreamIncoming, request StreamIncomingFilter (empty), response
// IncomingPayment{ oneof details { OffChainPayment offchain = 1 } }.
func (g *GreenlightService) streamIncoming(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		stream, err := g.conn.NewStream(ctx, &grpc.StreamDesc{
			ServerStreams: true,
		}, "/greenlight.Node/StreamIncoming", grpc.CallCustomCodec(rawCodec{}))
		if err != nil {
			logger.Logger.WithError(err).Error("Failed to open StreamIncoming")
			if !sleepOrDone(ctx, 10*time.Second) {
				return
			}
			continue
		}

		// empty StreamIncomingFilter
		if err := stream.SendMsg(&emptyMessage{}); err != nil {
			logger.Logger.WithError(err).Error("Failed to send StreamIncoming request")
			if !sleepOrDone(ctx, 10*time.Second) {
				return
			}
			continue
		}
		if err := stream.CloseSend(); err != nil {
			logger.Logger.WithError(err).Debug("StreamIncoming CloseSend failed")
		}

		logger.Logger.Info("Subscribed to StreamIncoming (keysend)")

		for {
			msg := &incomingPaymentMessage{}
			err := stream.RecvMsg(msg)
			if err != nil {
				if ctx.Err() == nil {
					logger.Logger.WithError(err).Error("StreamIncoming receive failed, resubscribing")
				}
				break
			}
			offchain := msg.offchain
			if offchain == nil {
				continue
			}
			// invoice payments are handled by the WaitAnyInvoice pump
			if offchain.Bolt11 != "" {
				continue
			}

			paymentHash := hex.EncodeToString(offchain.PaymentHash)
			preimage := hex.EncodeToString(offchain.Preimage)
			now := time.Now().Unix()

			var tlvRecords []lnclient.TLVRecord
			for _, tlv := range offchain.ExtraTLVs {
				tlvRecords = append(tlvRecords, lnclient.TLVRecord{
					Type:  tlv.Type,
					Value: hex.EncodeToString(tlv.Value),
				})
			}

			var amountMsat uint64
			if offchain.Amount != nil {
				amountMsat = offchain.Amount.Msat
			}

			transaction := &lnclient.Transaction{
				Type:         "incoming",
				Invoice:      offchain.Bolt11,
				Description:  "",
				Preimage:     preimage,
				PaymentHash:  paymentHash,
				AmountMsat:   int64(amountMsat),
				FeesPaidMsat: 0,
				CreatedAt:    now,
				SettledAt:    &now,
				Metadata: lnclient.Metadata{
					"tlv_records": tlvRecords,
				},
			}

			logger.Logger.WithFields(logrus.Fields{
				"payment_hash": paymentHash,
				"amount_msat":  amountMsat,
			}).Info("Incoming keysend payment")

			g.eventPublisher.Publish(&events.Event{
				Event:      "nwc_lnclient_payment_received",
				Properties: transaction,
			})
		}

		// stream ended: back off and resubscribe
		if !sleepOrDone(ctx, 10*time.Second) {
			return
		}
	}
}

func sleepOrDone(ctx context.Context, d time.Duration) bool {
	select {
	case <-ctx.Done():
		return false
	case <-time.After(d):
		return true
	}
}

// --- raw codec + hand-rolled message types (no generated bindings) ---

// rawCodec is a grpc codec that passes through raw bytes for the hand-rolled
// message types and falls back to the protobuf runtime for generated types.
// It is only attached to the StreamIncoming call via CallCustomCodec.
type rawCodec struct{}

func (rawCodec) Marshal(v interface{}) ([]byte, error) {
	switch m := v.(type) {
	case []byte:
		return m, nil
	case *emptyMessage:
		return nil, nil
	case *incomingPaymentMessage:
		return nil, errors.New("incomingPaymentMessage is receive-only")
	default:
		if pm, ok := v.(proto.Message); ok {
			return proto.Marshal(pm)
		}
		return nil, fmt.Errorf("rawCodec: cannot marshal %T", v)
	}
}

func (rawCodec) Unmarshal(data []byte, v interface{}) error {
	switch m := v.(type) {
	case *incomingPaymentMessage:
		return m.decode(data)
	case *emptyMessage:
		return nil
	default:
		if pm, ok := v.(proto.Message); ok {
			return proto.Unmarshal(data, pm)
		}
		return nil
	}
}

func (rawCodec) Name() string { return "proto" }

func (rawCodec) String() string { return "proto" }

// emptyMessage is the empty StreamIncomingFilter request.
type emptyMessage struct{}

func (*emptyMessage) Reset()         {}
func (*emptyMessage) String() string { return "empty" }
func (*emptyMessage) ProtoMessage()  {}

type incomingPaymentMessage struct {
	offchain *offChainPayment
}

func (*incomingPaymentMessage) Reset()         {}
func (*incomingPaymentMessage) String() string { return "IncomingPayment" }
func (*incomingPaymentMessage) ProtoMessage()  {}

func (i *incomingPaymentMessage) decode(data []byte) error {
	for len(data) > 0 {
		num, typ, n := protowire.ConsumeTag(data)
		if n < 0 {
			return protowire.ParseError(n)
		}
		data = data[n:]
		switch {
		case num == 1 && typ == protowire.BytesType:
			payload, n := protowire.ConsumeBytes(data)
			if n < 0 {
				return protowire.ParseError(n)
			}
			data = data[n:]
			op := &offChainPayment{}
			if err := op.decode(payload); err != nil {
				return err
			}
			i.offchain = op
		default:
			n := protowire.ConsumeFieldValue(num, typ, data)
			if n < 0 {
				return protowire.ParseError(n)
			}
			data = data[n:]
		}
	}
	return nil
}

// offChainPayment mirrors greenlight.OffChainPayment:
//
//	string label = 1;
//	bytes preimage = 2;
//	Amount amount = 3;        // { uint64 msat = 1; }
//	repeated TlvField extratlvs = 4;  // { uint64 type = 1; bytes value = 2; }
//	bytes payment_hash = 5;
//	string bolt11 = 6;
type offChainPayment struct {
	Label       string
	Preimage    []byte
	Amount      *greenlightAmount
	ExtraTLVs   []*tlvField
	PaymentHash []byte
	Bolt11      string
}

type greenlightAmount struct {
	Msat uint64
}

type tlvField struct {
	Type  uint64
	Value []byte
}

func (o *offChainPayment) decode(data []byte) error {
	for len(data) > 0 {
		num, typ, n := protowire.ConsumeTag(data)
		if n < 0 {
			return protowire.ParseError(n)
		}
		data = data[n:]
		switch {
		case num == 1 && typ == protowire.BytesType:
			v, n := protowire.ConsumeString(data)
			if n < 0 {
				return protowire.ParseError(n)
			}
			data = data[n:]
			o.Label = v
		case num == 2 && typ == protowire.BytesType:
			v, n := protowire.ConsumeBytes(data)
			if n < 0 {
				return protowire.ParseError(n)
			}
			data = data[n:]
			o.Preimage = v
		case num == 3 && typ == protowire.BytesType:
			payload, n := protowire.ConsumeBytes(data)
			if n < 0 {
				return protowire.ParseError(n)
			}
			data = data[n:]
			amt := &greenlightAmount{}
			if err := amt.decode(payload); err != nil {
				return err
			}
			o.Amount = amt
		case num == 4 && typ == protowire.BytesType:
			payload, n := protowire.ConsumeBytes(data)
			if n < 0 {
				return protowire.ParseError(n)
			}
			data = data[n:]
			tlv := &tlvField{}
			if err := tlv.decode(payload); err != nil {
				return err
			}
			o.ExtraTLVs = append(o.ExtraTLVs, tlv)
		case num == 5 && typ == protowire.BytesType:
			v, n := protowire.ConsumeBytes(data)
			if n < 0 {
				return protowire.ParseError(n)
			}
			data = data[n:]
			o.PaymentHash = v
		case num == 6 && typ == protowire.BytesType:
			v, n := protowire.ConsumeString(data)
			if n < 0 {
				return protowire.ParseError(n)
			}
			data = data[n:]
			o.Bolt11 = v
		default:
			n := protowire.ConsumeFieldValue(num, typ, data)
			if n < 0 {
				return protowire.ParseError(n)
			}
			data = data[n:]
		}
	}
	return nil
}

func (a *greenlightAmount) decode(data []byte) error {
	for len(data) > 0 {
		num, typ, n := protowire.ConsumeTag(data)
		if n < 0 {
			return protowire.ParseError(n)
		}
		data = data[n:]
		switch {
		case num == 1 && typ == protowire.VarintType:
			v, n := protowire.ConsumeVarint(data)
			if n < 0 {
				return protowire.ParseError(n)
			}
			data = data[n:]
			a.Msat = v
		default:
			n := protowire.ConsumeFieldValue(num, typ, data)
			if n < 0 {
				return protowire.ParseError(n)
			}
			data = data[n:]
		}
	}
	return nil
}

func (t *tlvField) decode(data []byte) error {
	for len(data) > 0 {
		num, typ, n := protowire.ConsumeTag(data)
		if n < 0 {
			return protowire.ParseError(n)
		}
		data = data[n:]
		switch {
		case num == 1 && typ == protowire.VarintType:
			v, n := protowire.ConsumeVarint(data)
			if n < 0 {
				return protowire.ParseError(n)
			}
			data = data[n:]
			t.Type = v
		case num == 2 && typ == protowire.BytesType:
			v, n := protowire.ConsumeBytes(data)
			if n < 0 {
				return protowire.ParseError(n)
			}
			data = data[n:]
			t.Value = v
		default:
			n := protowire.ConsumeFieldValue(num, typ, data)
			if n < 0 {
				return protowire.ParseError(n)
			}
			data = data[n:]
		}
	}
	return nil
}

var _ encoding.Codec = rawCodec{}
