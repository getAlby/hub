package lnd

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	decodepay "github.com/nbd-wtf/ln-decodepay"
	"google.golang.org/grpc/status"

	"github.com/getAlby/hub/events"
	"github.com/getAlby/hub/lnclient"
	"github.com/getAlby/hub/lnclient/lnd/wrapper"
	"github.com/getAlby/hub/logger"

	"github.com/sirupsen/logrus"
	// "gorm.io/gorm"

	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/lightningnetwork/lnd/lnrpc/routerrpc"
)

type LNDService struct {
	client *wrapper.LNDWrapper
	pubkey string
	cancel context.CancelFunc
}

func (svc *LNDService) GetBalance(ctx context.Context) (balance int64, err error) {
	resp, err := svc.client.ChannelBalance(ctx, &lnrpc.ChannelBalanceRequest{})
	if err != nil {
		return 0, err
	}
	return int64(resp.LocalBalance.Msat), nil
}

// FIXME: this always returns limit * 2 transactions and offset is not used correctly
func (svc *LNDService) ListTransactions(ctx context.Context, from, until, limit, offset uint64, unpaid bool, invoiceType string) (transactions []lnclient.Transaction, err error) {
	// Fetch invoices
	var invoices []*lnrpc.Invoice
	if invoiceType == "" || invoiceType == "incoming" {
		incomingResp, err := svc.client.ListInvoices(ctx, &lnrpc.ListInvoiceRequest{Reversed: true, NumMaxInvoices: limit, IndexOffset: offset})
		if err != nil {
			return nil, err
		}
		invoices = incomingResp.Invoices
	}
	for _, invoice := range invoices {
		// this will cause retrieved amount to be less than limit if unpaid is false
		if !unpaid && invoice.State != lnrpc.Invoice_SETTLED {
			continue
		}

		transaction := lndInvoiceToTransaction(invoice)
		transactions = append(transactions, *transaction)
	}
	// Fetch payments
	var payments []*lnrpc.Payment
	if invoiceType == "" || invoiceType == "outgoing" {
		// Not just pending but failed payments will also be included because of IncludeIncomplete
		outgoingResp, err := svc.client.ListPayments(ctx, &lnrpc.ListPaymentsRequest{Reversed: true, MaxPayments: limit, IndexOffset: offset, IncludeIncomplete: unpaid})
		if err != nil {
			return nil, err
		}
		payments = outgoingResp.Payments
	}
	for _, payment := range payments {
		if payment.Status == lnrpc.Payment_FAILED {
			// don't return failed payments for now
			// this will cause retrieved amount to be less than limit
			continue
		}

		transaction, err := lndPaymentToTransaction(payment)
		if err != nil {
			return nil, err
		}

		transactions = append(transactions, *transaction)
	}

	// sort by created date descending
	sort.SliceStable(transactions, func(i, j int) bool {
		return transactions[i].CreatedAt > transactions[j].CreatedAt
	})

	return transactions, nil
}

func (svc *LNDService) GetInfo(ctx context.Context) (info *lnclient.NodeInfo, err error) {
	resp, err := svc.client.GetInfo(ctx, &lnrpc.GetInfoRequest{})
	if err != nil {
		return nil, err
	}
	network := resp.Chains[0].Network
	if network == "mainnet" {
		network = "bitcoin"
	}
	return &lnclient.NodeInfo{
		Alias:       resp.Alias,
		Color:       resp.Color,
		Pubkey:      resp.IdentityPubkey,
		Network:     network,
		BlockHeight: resp.BlockHeight,
		BlockHash:   resp.BlockHash,
	}, nil
}

func (svc *LNDService) parseChannelPoint(channelPointStr string) (*lnrpc.ChannelPoint, error) {
	channelPointParts := strings.Split(channelPointStr, ":")

	if len(channelPointParts) == 2 {
		channelPoint := &lnrpc.ChannelPoint{}
		channelPoint.FundingTxid = &lnrpc.ChannelPoint_FundingTxidStr{
			FundingTxidStr: channelPointParts[0],
		}

		outputIndex, err := strconv.ParseUint(channelPointParts[1], 10, 32)
		if err != nil {
			return nil, err
		}
		channelPoint.OutputIndex = uint32(outputIndex)

		return channelPoint, nil
	}

	return nil, errors.New("invalid channel point")
}

func (svc *LNDService) ListChannels(ctx context.Context) ([]lnclient.Channel, error) {
	activeResp, err := svc.client.ListChannels(ctx, &lnrpc.ListChannelsRequest{})
	if err != nil {
		return nil, err
	}
	pendingResp, err := svc.client.PendingChannels(ctx, &lnrpc.PendingChannelsRequest{})
	if err != nil {
		return nil, err
	}

	nodeInfo, err := svc.GetInfo(ctx)
	if err != nil {
		return nil, err
	}

	// hardcoding required confirmations as there seems to be no way to get the number of required confirmations in LND
	var confirmationsRequired uint32 = 6
	// get recent transactions to check how many confirmations pending channel(s) have
	recentOnchainTransactions, err := svc.client.GetTransactions(ctx, &lnrpc.GetTransactionsRequest{
		StartHeight: int32(nodeInfo.BlockHeight - confirmationsRequired),
	})
	if err != nil {
		return nil, err
	}

	channels := make([]lnclient.Channel, len(activeResp.Channels)+len(pendingResp.PendingOpenChannels))

	for i, lndChannel := range activeResp.Channels {
		channelPoint, err := svc.parseChannelPoint(lndChannel.ChannelPoint)
		if err != nil {
			return nil, err
		}

		// first 3 bytes of the channel ID are the block height
		channelOpeningBlockHeight := lndChannel.ChanId >> 40
		confirmations := nodeInfo.BlockHeight - uint32(channelOpeningBlockHeight)

		channels[i] = lnclient.Channel{
			InternalChannel:                          lndChannel,
			LocalBalance:                             lndChannel.LocalBalance * 1000,
			LocalSpendableBalance:                    int64(math.Max(float64((lndChannel.LocalBalance-int64(lndChannel.LocalConstraints.ChanReserveSat))*1000), float64(0))),
			RemoteBalance:                            lndChannel.RemoteBalance * 1000,
			RemotePubkey:                             lndChannel.RemotePubkey,
			Id:                                       strconv.FormatUint(lndChannel.ChanId, 10),
			Active:                                   lndChannel.Active,
			Public:                                   !lndChannel.Private,
			FundingTxId:                              channelPoint.GetFundingTxidStr(),
			Confirmations:                            &confirmations,
			ConfirmationsRequired:                    &confirmationsRequired,
			UnspendablePunishmentReserve:             lndChannel.LocalConstraints.ChanReserveSat,
			CounterpartyUnspendablePunishmentReserve: lndChannel.RemoteConstraints.ChanReserveSat,
		}
	}

	for j, lndChannel := range pendingResp.PendingOpenChannels {
		channelPoint, err := svc.parseChannelPoint(lndChannel.Channel.ChannelPoint)
		if err != nil {
			return nil, err
		}
		fundingTxId := channelPoint.GetFundingTxidStr()

		var confirmations *uint32
		for _, t := range recentOnchainTransactions.Transactions {
			if t.TxHash == fundingTxId {
				confirmations32 := uint32(t.NumConfirmations)
				confirmations = &confirmations32
			}
		}

		channels[j+len(activeResp.Channels)] = lnclient.Channel{
			InternalChannel:       lndChannel,
			LocalBalance:          lndChannel.Channel.LocalBalance * 1000,
			RemoteBalance:         lndChannel.Channel.RemoteBalance * 1000,
			RemotePubkey:          lndChannel.Channel.RemoteNodePub,
			Public:                !lndChannel.Channel.Private,
			FundingTxId:           fundingTxId,
			Active:                false,
			Confirmations:         confirmations,
			ConfirmationsRequired: &confirmationsRequired,
		}
	}

	return channels, nil
}

func (svc *LNDService) MakeInvoice(ctx context.Context, amount int64, description string, descriptionHash string, expiry int64) (transaction *lnclient.Transaction, err error) {
	var descriptionHashBytes []byte

	if descriptionHash != "" {
		descriptionHashBytes, err = hex.DecodeString(descriptionHash)

		if err != nil || len(descriptionHashBytes) != 32 {
			logger.Logger.WithFields(logrus.Fields{
				"amount":          amount,
				"description":     description,
				"descriptionHash": descriptionHash,
				"expiry":          expiry,
			}).Errorf("Invalid description hash")
			return nil, errors.New("description hash must be 32 bytes hex")
		}
	}

	if expiry == 0 {
		expiry = lnclient.DEFAULT_INVOICE_EXPIRY
	}

	channels, err := svc.ListChannels(ctx)
	if err != nil {
		return nil, err
	}

	hasPublicChannels := false
	for _, channel := range channels {
		if channel.Active && channel.Public {
			hasPublicChannels = true
		}
	}

	addInvoiceRequest := &lnrpc.Invoice{
		ValueMsat:       amount,
		Memo:            description,
		DescriptionHash: descriptionHashBytes,
		Expiry:          expiry,
		Private:         !hasPublicChannels, // use private channel hints in the invoice
	}

	resp, err := svc.client.AddInvoice(ctx, addInvoiceRequest)

	if err != nil {
		return nil, err
	}

	inv, err := svc.client.LookupInvoice(ctx, &lnrpc.PaymentHash{RHash: resp.RHash})
	if err != nil {
		return nil, err
	}

	transaction = lndInvoiceToTransaction(inv)
	return transaction, nil
}

func (svc *LNDService) LookupInvoice(ctx context.Context, paymentHash string) (transaction *lnclient.Transaction, err error) {
	paymentHashBytes, err := hex.DecodeString(paymentHash)

	if err != nil || len(paymentHashBytes) != 32 {
		logger.Logger.WithFields(logrus.Fields{
			"paymentHash": paymentHash,
		}).Errorf("Invalid payment hash")
		return nil, errors.New("Payment hash must be 32 bytes hex")
	}

	lndInvoice, err := svc.client.LookupInvoice(ctx, &lnrpc.PaymentHash{RHash: paymentHashBytes})
	if err != nil {
		return nil, err
	}

	transaction = lndInvoiceToTransaction(lndInvoice)
	return transaction, nil
}

func (svc *LNDService) SendPaymentSync(ctx context.Context, payReq string) (*lnclient.PayInvoiceResponse, error) {
	resp, err := svc.client.SendPaymentSync(ctx, &lnrpc.SendRequest{PaymentRequest: payReq})
	if err != nil {
		return nil, err
	}

	if resp.PaymentError != "" {
		return nil, errors.New(resp.PaymentError)
	}

	if resp.PaymentPreimage == nil {
		return nil, errors.New("no preimage in response")
	}

	var fee uint64 = 0
	if resp.PaymentRoute != nil {
		fee = uint64(resp.PaymentRoute.TotalFeesMsat)
	}

	return &lnclient.PayInvoiceResponse{
		Preimage: hex.EncodeToString(resp.PaymentPreimage),
		Fee:      fee,
	}, nil
}

func (svc *LNDService) SendKeysend(ctx context.Context, amount uint64, destination string, custom_records []lnclient.TLVRecord, preimage string) (*lnclient.PayKeysendResponse, error) {
	destBytes, err := hex.DecodeString(destination)
	if err != nil {
		return nil, err
	}
	preImageBytes, err := hex.DecodeString(preimage)
	if err != nil || len(preImageBytes) != 32 {
		logger.Logger.WithFields(logrus.Fields{
			"preimage": preimage,
		}).WithError(err).Error("Invalid preimage")
		return nil, err
	}

	paymentHash256 := sha256.New()
	paymentHash256.Write(preImageBytes)
	paymentHashBytes := paymentHash256.Sum(nil)
	paymentHash := hex.EncodeToString(paymentHashBytes)

	destCustomRecords := map[uint64][]byte{}
	for _, record := range custom_records {
		decodedValue, err := hex.DecodeString(record.Value)
		if err != nil {
			return nil, err
		}
		destCustomRecords[record.Type] = decodedValue
	}
	const KEYSEND_CUSTOM_RECORD = 5482373484
	destCustomRecords[KEYSEND_CUSTOM_RECORD] = preImageBytes
	sendPaymentRequest := &lnrpc.SendRequest{
		Dest:              destBytes,
		AmtMsat:           int64(amount),
		PaymentHash:       paymentHashBytes,
		DestFeatures:      []lnrpc.FeatureBit{lnrpc.FeatureBit_TLV_ONION_REQ},
		DestCustomRecords: destCustomRecords,
	}

	resp, err := svc.client.SendPaymentSync(ctx, sendPaymentRequest)
	if err != nil {
		logger.Logger.WithFields(logrus.Fields{
			"amount":        amount,
			"payeePubkey":   destination,
			"paymentHash":   paymentHash,
			"preimage":      preimage,
			"customRecords": custom_records,
			"error":         err,
		}).Errorf("Failed to send keysend payment")
		return nil, err
	}
	if resp.PaymentError != "" {
		logger.Logger.WithFields(logrus.Fields{
			"amount":        amount,
			"payeePubkey":   destination,
			"paymentHash":   paymentHash,
			"preimage":      preimage,
			"customRecords": custom_records,
			"paymentError":  resp.PaymentError,
		}).Errorf("Keysend payment has payment error")
		return nil, errors.New(resp.PaymentError)
	}
	respPreimage := hex.EncodeToString(resp.PaymentPreimage)
	if respPreimage != preimage {
		logger.Logger.WithFields(logrus.Fields{
			"amount":        amount,
			"payeePubkey":   destination,
			"paymentHash":   paymentHash,
			"preimage":      preimage,
			"customRecords": custom_records,
			"paymentError":  resp.PaymentError,
		}).Errorf("Preimage in keysend response does not match")
		return nil, errors.New("preimage in keysend response does not match")
	}
	logger.Logger.WithFields(logrus.Fields{
		"amount":        amount,
		"payeePubkey":   destination,
		"paymentHash":   paymentHash,
		"preimage":      preimage,
		"customRecords": custom_records,
		"respPreimage":  respPreimage,
	}).Info("Keysend payment successful")

	return &lnclient.PayKeysendResponse{
		Fee: uint64(resp.PaymentRoute.TotalFeesMsat),
	}, nil
}

func NewLNDService(ctx context.Context, eventPublisher events.EventPublisher, lndAddress, lndCertHex, lndMacaroonHex string) (result lnclient.LNClient, err error) {
	if lndAddress == "" || lndCertHex == "" || lndMacaroonHex == "" {
		return nil, errors.New("one or more required LND configuration are missing")
	}

	lndClient, err := wrapper.NewLNDclient(wrapper.LNDoptions{
		Address:     lndAddress,
		CertHex:     lndCertHex,
		MacaroonHex: lndMacaroonHex,
	})
	if err != nil {
		logger.Logger.Errorf("Failed to create new LND client %v", err)
		return nil, err
	}
	info, err := lndClient.GetInfo(ctx, &lnrpc.GetInfoRequest{})
	if err != nil {
		return nil, err
	}

	lndCtx, cancel := context.WithCancel(ctx)

	lndService := &LNDService{client: lndClient, pubkey: info.IdentityPubkey, cancel: cancel}

	// Subscribe to payments
	go func() {
		for {
			paymentStream, err := lndClient.SubscribePayments(lndCtx, &routerrpc.TrackPaymentsRequest{
				NoInflightUpdates: true,
			})
			if err != nil {
				logger.Logger.WithError(err).Error("Error subscribing to payments")
				continue
			}
			for {
				select {
				case <-lndCtx.Done():
					return
				default:
					payment, err := paymentStream.Recv()
					if err != nil {
						logger.Logger.WithError(err).Error("Failed to receive payment")
						continue
					}

					switch payment.Status {
					case lnrpc.Payment_FAILED:
						logger.Logger.WithFields(logrus.Fields{
							"payment": payment,
						}).Info("Received payment failed notification")

						transaction, err := lndPaymentToTransaction(payment)
						if err != nil {
							continue
						}
						eventPublisher.Publish(&events.Event{
							Event: "nwc_payment_failed_async",
							Properties: &events.PaymentFailedAsyncProperties{
								Transaction: transaction,
								Reason:      payment.FailureReason.String(),
							},
						})
					case lnrpc.Payment_SUCCEEDED:
						logger.Logger.WithFields(logrus.Fields{
							"payment": payment,
						}).Info("Received payment sent notification")

						transaction, err := lndPaymentToTransaction(payment)
						if err != nil {
							continue
						}
						eventPublisher.Publish(&events.Event{
							Event:      "nwc_payment_sent",
							Properties: transaction,
						})
					default:
						continue
					}
				}
			}
		}
	}()

	// Subscribe to invoices
	go func() {
		for {
			invoiceStream, err := lndClient.SubscribeInvoices(lndCtx, &lnrpc.InvoiceSubscription{})
			if err != nil {
				logger.Logger.WithError(err).Error("Error subscribing to invoices")
				continue
			}
			for {
				select {
				case <-lndCtx.Done():
					return
				default:
					invoice, err := invoiceStream.Recv()
					if err != nil {
						logger.Logger.WithError(err).Error("Failed to receive invoice")
						continue
					}
					if invoice.State != lnrpc.Invoice_SETTLED {
						continue
					}

					logger.Logger.WithFields(logrus.Fields{
						"invoice": invoice,
					}).Info("Received new invoice")

					eventPublisher.Publish(&events.Event{
						Event:      "nwc_payment_received",
						Properties: lndInvoiceToTransaction(invoice),
					})
				}
			}
		}
	}()

	logger.Logger.Infof("Connected to LND - alias %s", info.Alias)

	return lndService, nil
}

func (svc *LNDService) Shutdown() error {
	logger.Logger.Info("cancelling LND context")
	svc.cancel()
	return nil
}

func (svc *LNDService) GetNodeConnectionInfo(ctx context.Context) (nodeConnectionInfo *lnclient.NodeConnectionInfo, err error) {
	info, err := svc.client.GetInfo(ctx, &lnrpc.GetInfoRequest{})
	if err != nil {
		return nil, err
	}

	return &lnclient.NodeConnectionInfo{
		Pubkey: info.IdentityPubkey,
		//Address: address,
		//Port:    port,
	}, nil
}

func (svc *LNDService) ConnectPeer(ctx context.Context, connectPeerRequest *lnclient.ConnectPeerRequest) error {
	_, err := svc.client.ConnectPeer(ctx, &lnrpc.ConnectPeerRequest{
		Addr: &lnrpc.LightningAddress{
			Pubkey: connectPeerRequest.Pubkey,
			Host:   connectPeerRequest.Address + ":" + strconv.Itoa(int(connectPeerRequest.Port)),
		},
	})

	if grpcErr, ok := status.FromError(err); ok {
		if strings.HasPrefix(grpcErr.Message(), "already connected to peer") {
			return nil
		}
	}
	return err
}

func (svc *LNDService) OpenChannel(ctx context.Context, openChannelRequest *lnclient.OpenChannelRequest) (*lnclient.OpenChannelResponse, error) {
	peers, err := svc.ListPeers(ctx)
	var foundPeer *lnclient.PeerDetails
	for _, peer := range peers {
		if peer.NodeId == openChannelRequest.Pubkey {

			foundPeer = &peer
			break
		}
	}

	if foundPeer == nil {
		return nil, errors.New("node is not peered yet")
	}

	logger.Logger.WithField("peer_id", foundPeer.NodeId).Info("Opening channel")

	nodePub, err := hex.DecodeString(openChannelRequest.Pubkey)
	if err != nil {
		return nil, errors.New("failed to decode pubkey")
	}

	channel, err := svc.client.OpenChannelSync(ctx, &lnrpc.OpenChannelRequest{
		NodePubkey:         nodePub,
		Private:            !openChannelRequest.Public,
		LocalFundingAmount: openChannelRequest.Amount,
		// set a super-high forwarding fee of 100K sats by default to disable unwanted routing
		BaseFee: 100_000_000,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to open channel with %s: %s", foundPeer.NodeId, err)
	}

	fundingTxidBytes := channel.GetFundingTxidBytes()

	// we get the funding transaction id bytes in reverse
	for i, j := 0, len(fundingTxidBytes)-1; i < j; i, j = i+1, j-1 {
		fundingTxidBytes[i], fundingTxidBytes[j] = fundingTxidBytes[j], fundingTxidBytes[i]
	}

	return &lnclient.OpenChannelResponse{
		FundingTxId: hex.EncodeToString(fundingTxidBytes),
	}, err
}

func (svc *LNDService) CloseChannel(ctx context.Context, closeChannelRequest *lnclient.CloseChannelRequest) (*lnclient.CloseChannelResponse, error) {
	logger.Logger.WithFields(logrus.Fields{
		"request": closeChannelRequest,
	}).Info("Closing Channel")

	resp, err := svc.client.ListChannels(ctx, &lnrpc.ListChannelsRequest{})
	if err != nil {
		return nil, err
	}

	var foundChannel *lnrpc.Channel
	for _, channel := range resp.Channels {
		if strconv.FormatUint(channel.ChanId, 10) == closeChannelRequest.ChannelId {

			foundChannel = channel
			break
		}
	}

	if foundChannel == nil {
		return nil, errors.New("no channel exists with the given id")
	}

	channelPoint, err := svc.parseChannelPoint(foundChannel.ChannelPoint)
	if err != nil {
		return nil, err
	}

	stream, err := svc.client.CloseChannel(ctx, &lnrpc.CloseChannelRequest{
		ChannelPoint: channelPoint,
		Force:        closeChannelRequest.Force,
	})
	if err != nil {
		return nil, err
	}

	for {
		resp, err := stream.Recv()
		if err != nil {
			return nil, err
		}

		switch update := resp.Update.(type) {
		case *lnrpc.CloseStatusUpdate_ClosePending:
			closingHash := update.ClosePending.Txid
			txid, err := chainhash.NewHash(closingHash)
			if err != nil {
				return nil, err
			}
			logger.Logger.WithFields(logrus.Fields{
				"closingTxid": txid.String(),
			}).Info("Channel close pending")
			// TODO: return the closing tx id or fire an event
			return &lnclient.CloseChannelResponse{}, nil
		}
	}
}

func (svc *LNDService) GetNewOnchainAddress(ctx context.Context) (string, error) {
	resp, err := svc.client.NewAddress(ctx, &lnrpc.NewAddressRequest{
		Type: lnrpc.AddressType_WITNESS_PUBKEY_HASH,
	})
	if err != nil {
		logger.Logger.WithError(err).Error("NewOnchainAddress failed")
		return "", err
	}
	return resp.Address, nil
}

func (svc *LNDService) GetOnchainBalance(ctx context.Context) (*lnclient.OnchainBalanceResponse, error) {
	balances, err := svc.client.WalletBalance(ctx, &lnrpc.WalletBalanceRequest{})
	if err != nil {
		return nil, err
	}
	logger.Logger.WithFields(logrus.Fields{
		"balances": balances,
	}).Debug("Listed Balances")
	return &lnclient.OnchainBalanceResponse{
		Spendable: int64(balances.ConfirmedBalance),
		Total:     int64(balances.TotalBalance),
		Reserved:  int64(balances.ReservedBalanceAnchorChan),
	}, nil
}

func (svc *LNDService) RedeemOnchainFunds(ctx context.Context, toAddress string) (txId string, err error) {
	return "", nil
}

func (svc *LNDService) SendPaymentProbes(ctx context.Context, invoice string) error {
	return nil
}

func (svc *LNDService) SendSpontaneousPaymentProbes(ctx context.Context, amountMsat uint64, nodeId string) error {
	return nil
}

func (svc *LNDService) ListPeers(ctx context.Context) ([]lnclient.PeerDetails, error) {
	resp, err := svc.client.ListPeers(ctx, &lnrpc.ListPeersRequest{})
	ret := make([]lnclient.PeerDetails, 0, len(resp.Peers))
	for _, peer := range resp.Peers {
		ret = append(ret, lnclient.PeerDetails{
			NodeId:      peer.PubKey,
			Address:     peer.Address,
			IsPersisted: true,
			IsConnected: true,
		})
	}
	return ret, err
}

func (svc *LNDService) GetLogOutput(ctx context.Context, maxLen int) ([]byte, error) {
	return []byte{}, nil
}

func (svc *LNDService) SignMessage(ctx context.Context, message string) (string, error) {
	resp, err := svc.client.SignMessage(ctx, &lnrpc.SignMessageRequest{Msg: []byte(message)})
	if err != nil {
		return "", err
	}

	return resp.Signature, nil
}

func (svc *LNDService) GetBalances(ctx context.Context) (*lnclient.BalancesResponse, error) {
	onchainBalance, err := svc.GetOnchainBalance(ctx)
	if err != nil {
		logger.Logger.WithError(err).Error("Failed to retrieve onchain balance")
		return nil, err
	}

	var totalReceivable int64 = 0
	var totalSpendable int64 = 0
	var nextMaxReceivable int64 = 0
	var nextMaxSpendable int64 = 0
	var nextMaxReceivableMPP int64 = 0
	var nextMaxSpendableMPP int64 = 0
	resp, err := svc.client.ListChannels(ctx, &lnrpc.ListChannelsRequest{})

	for _, channel := range resp.Channels {
		// Unnecessary since ListChannels only returns active channels
		if channel.Active {
			channelMinSpendable := channel.LocalBalance * 1000
			channelMinReceivable := channel.RemoteBalance * 1000

			nextMaxSpendable = max(nextMaxSpendable, channelMinSpendable)
			nextMaxReceivable = max(nextMaxReceivable, channelMinReceivable)

			totalSpendable += channelMinSpendable
			totalReceivable += channelMinReceivable
		}
	}

	return &lnclient.BalancesResponse{
		Onchain: *onchainBalance,
		Lightning: lnclient.LightningBalanceResponse{
			TotalSpendable:    totalSpendable,
			TotalReceivable:   totalReceivable,
			NextMaxSpendable:  nextMaxSpendable,
			NextMaxReceivable: nextMaxReceivable,
			// TODO: return actuall MPP instead of 0
			NextMaxSpendableMPP:  nextMaxSpendableMPP,
			NextMaxReceivableMPP: nextMaxReceivableMPP,
		},
	}, nil
}

func lndPaymentToTransaction(payment *lnrpc.Payment) (*lnclient.Transaction, error) {
	var expiresAt *int64
	var description string
	var descriptionHash string
	if payment.PaymentRequest != "" {
		paymentRequest, err := decodepay.Decodepay(strings.ToLower(payment.PaymentRequest))
		if err != nil {
			logger.Logger.WithFields(logrus.Fields{
				"bolt11": payment.PaymentRequest,
			}).Errorf("Failed to decode bolt11 invoice: %v", err)
			return nil, err
		}
		expiresAtUnix := time.UnixMilli(int64(paymentRequest.CreatedAt) * 1000).Add(time.Duration(paymentRequest.Expiry) * time.Second).Unix()
		expiresAt = &expiresAtUnix
		description = paymentRequest.Description
		descriptionHash = paymentRequest.DescriptionHash
	}

	var settledAt *int64
	if payment.Status == lnrpc.Payment_SUCCEEDED {
		// FIXME: how to get the actual settled at time?
		settledAtUnix := time.Unix(0, payment.CreationTimeNs).Unix()
		settledAt = &settledAtUnix
	}

	return &lnclient.Transaction{
		Type:            "outgoing",
		Invoice:         payment.PaymentRequest,
		Preimage:        payment.PaymentPreimage,
		PaymentHash:     payment.PaymentHash,
		Amount:          payment.ValueMsat,
		FeesPaid:        payment.FeeMsat,
		CreatedAt:       time.Unix(0, payment.CreationTimeNs).Unix(),
		Description:     description,
		DescriptionHash: descriptionHash,
		ExpiresAt:       expiresAt,
		SettledAt:       settledAt,
		//TODO: Metadata:  (e.g. keysend),
	}, nil
}

func lndInvoiceToTransaction(invoice *lnrpc.Invoice) *lnclient.Transaction {
	var settledAt *int64
	preimage := hex.EncodeToString(invoice.RPreimage)
	metadata := map[string]interface{}{}

	if invoice.State == lnrpc.Invoice_SETTLED {
		settledAt = &invoice.SettleDate
	}
	var expiresAt *int64
	if invoice.Expiry > 0 {
		expiresAtUnix := invoice.CreationDate + invoice.Expiry
		expiresAt = &expiresAtUnix
	}

	if invoice.IsKeysend {
		tlvRecords := []lnclient.TLVRecord{}
		for _, htlc := range invoice.Htlcs {
			for key, value := range htlc.CustomRecords {
				tlvRecords = append(tlvRecords, lnclient.TLVRecord{
					Type:  key,
					Value: hex.EncodeToString(value),
				})
			}
		}
		metadata["tlv_records"] = tlvRecords
	}

	return &lnclient.Transaction{
		Type:            "incoming",
		Invoice:         invoice.PaymentRequest,
		Description:     invoice.Memo,
		DescriptionHash: hex.EncodeToString(invoice.DescriptionHash),
		Preimage:        preimage,
		PaymentHash:     hex.EncodeToString(invoice.RHash),
		Amount:          invoice.ValueMsat,
		CreatedAt:       invoice.CreationDate,
		SettledAt:       settledAt,
		ExpiresAt:       expiresAt,
		Metadata:        metadata,
	}
}

func (svc *LNDService) ResetRouter(key string) error {
	return nil
}

func (svc *LNDService) GetStorageDir() (string, error) {
	return "", nil
}

func (svc *LNDService) GetNodeStatus(ctx context.Context) (nodeStatus *lnclient.NodeStatus, err error) {
	return nil, nil
}

func (svc *LNDService) GetNetworkGraph(nodeIds []string) (lnclient.NetworkGraphResponse, error) {
	return nil, nil
}

func (svc *LNDService) UpdateLastWalletSyncRequest() {}

func (svc *LNDService) UpdateChannel(ctx context.Context, updateChannelRequest *lnclient.UpdateChannelRequest) error {
	logger.Logger.WithFields(logrus.Fields{
		"request": updateChannelRequest,
	}).Info("Updating Channel")

	chanId64, err := strconv.ParseUint(updateChannelRequest.ChannelId, 10, 64)
	if err != nil {
		return err
	}

	channelEdge, err := svc.client.GetChanInfo(ctx, &lnrpc.ChanInfoRequest{
		ChanId: chanId64,
	})
	if err != nil {
		return err
	}

	channelPoint, err := svc.parseChannelPoint(channelEdge.ChanPoint)
	if err != nil {
		return err
	}

	var nodePolicy *lnrpc.RoutingPolicy
	if channelEdge.Node1Pub == svc.client.IdentityPubkey {
		nodePolicy = channelEdge.Node1Policy
	} else {
		nodePolicy = channelEdge.Node2Policy
	}

	_, err = svc.client.UpdateChannel(ctx, &lnrpc.PolicyUpdateRequest{
		Scope: &lnrpc.PolicyUpdateRequest_ChanPoint{
			ChanPoint: channelPoint,
		},
		BaseFeeMsat:   int64(updateChannelRequest.ForwardingFeeBaseMsat),
		FeeRatePpm:    uint32(nodePolicy.FeeRateMilliMsat),
		TimeLockDelta: nodePolicy.TimeLockDelta,
		MaxHtlcMsat:   nodePolicy.MaxHtlcMsat,
	})

	if err != nil {
		return err
	}

	return nil
}

func (svc *LNDService) DisconnectPeer(ctx context.Context, peerId string) error {
	return nil
}

func (svc *LNDService) GetSupportedNIP47Methods() []string {
	return []string{
		"pay_invoice", "pay_keysend", "get_balance", "get_info", "make_invoice", "lookup_invoice", "list_transactions", "multi_pay_invoice", "multi_pay_keysend", "sign_message",
	}
}

func (svc *LNDService) GetSupportedNIP47NotificationTypes() []string {
	return []string{"payment_received", "payment_sent"}
}

func (svc *LNDService) GetPubkey() string {
	return svc.pubkey
}
