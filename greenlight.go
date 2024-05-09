package main

import (
	"context"
	"encoding/hex"
	"errors"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	//"github.com/getAlby/nostr-wallet-connect/glalby" // for local development only

	"github.com/getAlby/glalby-go/glalby"
	decodepay "github.com/nbd-wtf/ln-decodepay"
	"github.com/sirupsen/logrus"

	"github.com/getAlby/nostr-wallet-connect/models/lnclient"
)

type GreenlightService struct {
	workdir string
	client  *glalby.BlockingGreenlightAlbyClient
	svc     *Service
}

const DEVICE_CREDENTIALS_KEY = "GreenlightCreds"

func NewGreenlightService(svc *Service, mnemonic, inviteCode, workDir, encryptionKey string) (result lnclient.LNClient, err error) {
	if mnemonic == "" || inviteCode == "" || workDir == "" {
		return nil, errors.New("one or more required greenlight configuration are missing")
	}

	//create dir if not exists
	newpath := filepath.Join(workDir)
	err = os.MkdirAll(newpath, os.ModePerm)
	if err != nil {
		log.Printf("Failed to create greenlight working dir: %v", err)
		return nil, err
	}

	var credentials *glalby.GreenlightCredentials
	existingDeviceCreds, _ := svc.cfg.Get(DEVICE_CREDENTIALS_KEY, encryptionKey)

	if existingDeviceCreds != "" {
		credentials = &glalby.GreenlightCredentials{
			GlCreds: existingDeviceCreds,
		}
		svc.Logger.Info("Using saved greenlight credentials")
	}

	if credentials == nil {
		svc.Logger.Info("No greenlight credentials found, attempting to recover existing node")
		recoveredCredentials, err := glalby.Recover(mnemonic)
		credentials = &recoveredCredentials

		if err != nil {
			svc.Logger.Errorf("Failed to recover node: %v", err)
			svc.Logger.Infof("Trying to register instead...")
			recoveredCredentials, err := glalby.Register(mnemonic, inviteCode)
			credentials = &recoveredCredentials

			if err != nil {
				svc.Logger.Fatalf("Failed to register new node")
			}
		}

		if credentials == nil || credentials.GlCreds == "" {
			return nil, errors.New("unexpected response from Recover")
		}
		svc.cfg.SetUpdate(DEVICE_CREDENTIALS_KEY, credentials.GlCreds, encryptionKey)
	}

	client, err := glalby.NewBlockingGreenlightAlbyClient(mnemonic, *credentials)

	if err != nil {
		log.Printf("Failed to create greenlight alby client: %v", err)
		return nil, err
	}
	if client == nil {
		log.Fatalf("unexpected response from NewBlockingGreenlightAlbyClient")
	}

	gs := GreenlightService{
		workdir: newpath,
		client:  client,
		svc:     svc,
	}

	nodeInfo, err := client.GetInfo()

	if err != nil {
		return nil, err
	}

	log.Printf("Node info: %v", nodeInfo)

	return &gs, nil
}

func (gs *GreenlightService) Shutdown() error {
	_, err := gs.client.Shutdown()
	if err != nil {
		gs.svc.Logger.WithError(err).Error("Failed to shutdown greenlight node")
		return err
	}
	return nil
}

func (gs *GreenlightService) SendPaymentSync(ctx context.Context, payReq string) (*lnclient.Nip47PayInvoiceResponse, error) {
	response, err := gs.client.Pay(glalby.PayRequest{
		Bolt11: payReq,
	})

	if err != nil {
		gs.svc.Logger.Errorf("Failed to send payment: %v", err)
		return nil, err
	}
	log.Printf("SendPaymentSync succeeded: %v", response.Preimage)
	return &lnclient.Nip47PayInvoiceResponse{
		Preimage: response.Preimage,
	}, nil
}

func (gs *GreenlightService) SendKeysend(ctx context.Context, amount int64, destination, preimage string, custom_records []lnclient.TLVRecord) (preImage string, err error) {

	extraTlvs := []glalby.TlvEntry{}

	for _, customRecord := range custom_records {
		extraTlvs = append(extraTlvs, glalby.TlvEntry{
			Ty:    customRecord.Type,
			Value: hex.EncodeToString([]byte(customRecord.Value)),
		})
	}

	amount_u64 := uint64(amount)
	// TODO: support passing custom preimage
	response, err := gs.client.KeySend(glalby.KeySendRequest{
		Destination: destination,
		AmountMsat:  &amount_u64,
		ExtraTlvs:   &extraTlvs,
	})

	if err != nil {
		gs.svc.Logger.Errorf("Failed to send keysend payment: %v", err)
		return "", err
	}

	return response.PaymentPreimage, nil
}

func (gs *GreenlightService) GetBalance(ctx context.Context) (balance int64, err error) {
	response, err := gs.client.ListFunds(glalby.ListFundsRequest{})

	if err != nil {
		gs.svc.Logger.Errorf("Failed to list funds: %v", err)
		return 0, err
	}

	balance = 0
	for _, channel := range response.Channels {
		if channel.OurAmountMsat != nil && channel.Connected {
			balance += int64(*channel.OurAmountMsat)
		}
	}

	return balance, nil
}

func (gs *GreenlightService) MakeInvoice(ctx context.Context, amount int64, description string, descriptionHash string, expiry int64) (transaction *Nip47Transaction, err error) {
	uexpiry := uint64(expiry)
	// TODO: it seems description hash cannot be passed to greenlight
	invoice, err := gs.client.MakeInvoice(glalby.MakeInvoiceRequest{
		AmountMsat:  uint64(amount),
		Description: description,
		Label:       "label_" + strconv.Itoa(rand.Int()),
		Expiry:      &uexpiry,
	})

	if err != nil {
		gs.svc.Logger.Errorf("MakeInvoice failed: %v", err)
		return nil, err
	}

	paymentRequest, err := decodepay.Decodepay(strings.ToLower(invoice.Bolt11))
	if err != nil {
		gs.svc.Logger.WithFields(logrus.Fields{
			"invoice": invoice.Bolt11,
		}).Errorf("Failed to decode bolt11 invoice: %v", invoice.Bolt11)
		return nil, err
	}

	description = paymentRequest.Description
	descriptionHash = paymentRequest.DescriptionHash
	expiresAt := int64(invoice.ExpiresAt)
	transaction = &Nip47Transaction{
		Type:            "incoming",
		Invoice:         invoice.Bolt11,
		Description:     description,
		DescriptionHash: descriptionHash,
		PaymentHash:     invoice.PaymentHash,
		ExpiresAt:       &expiresAt,
		Amount:          amount,
		CreatedAt:       int64(paymentRequest.CreatedAt),
	}

	return transaction, nil
}

func (gs *GreenlightService) LookupInvoice(ctx context.Context, paymentHash string) (transaction *Nip47Transaction, err error) {
	response, err := gs.client.ListInvoices(glalby.ListInvoicesRequest{
		PaymentHash: &paymentHash,
	})

	if err != nil {
		gs.svc.Logger.Errorf("ListInvoices failed: %v", err)
		return nil, err
	}

	if len(response.Invoices) == 0 {
		return nil, errors.New("invoice not found")
	}
	invoice := response.Invoices[0]

	if invoice.Bolt11 == nil {
		return nil, errors.New("not a Bolt11 invoice")
	}

	transaction, err = gs.greenlightInvoiceToTransaction(&invoice)

	if err != nil {
		gs.svc.Logger.Errorf("Failed to map invoice: %v", err)
		return nil, err
	}

	return transaction, nil
}

func (gs *GreenlightService) ListTransactions(ctx context.Context, from, until, limit, offset uint64, unpaid bool, invoiceType string) (transactions []Nip47Transaction, err error) {
	listInvoicesResponse, err := gs.client.ListInvoices(glalby.ListInvoicesRequest{})

	if err != nil {
		gs.svc.Logger.Errorf("ListInvoices failed: %v", err)
		return nil, err
	}

	transactions = []Nip47Transaction{}

	if err != nil {
		log.Printf("ListInvoices failed: %v", err)
		return nil, err
	}

	for _, invoice := range listInvoicesResponse.Invoices {
		if invoice.PaidAt == nil {
			// skip unpaid invoices
			continue
		}

		transaction, err := gs.greenlightInvoiceToTransaction(&invoice)
		if err != nil {
			continue
		}

		transactions = append(transactions, *transaction)
	}

	listPaymentsResponse, err := gs.client.ListPayments(glalby.ListPaymentsRequest{})

	if err != nil {
		gs.svc.Logger.Errorf("ListPayments failed: %v", err)
		return nil, err
	}

	for _, payment := range listPaymentsResponse.Payments {
		description := ""
		descriptionHash := ""
		bolt11 := ""
		var expiresAt *int64
		if payment.Bolt11 != nil {
			// TODO: Greenlight should provide these details so we don't need to manually decode the invoice
			bolt11 := *payment.Bolt11
			paymentRequest, err := decodepay.Decodepay(strings.ToLower(bolt11))
			if err != nil {
				log.Printf("Failed to decode bolt11 invoice: %v", bolt11)
				return nil, err
			}

			description = paymentRequest.Description
			descriptionHash = paymentRequest.DescriptionHash
			expiresAtUnix := time.UnixMilli(int64(paymentRequest.CreatedAt) * 1000).Add(time.Duration(paymentRequest.Expiry) * time.Second).Unix()
			expiresAt = &expiresAtUnix
		}

		preimage := ""
		if payment.Preimage != nil {
			preimage = *payment.Preimage
		}
		var amount int64 = 0
		var fee int64 = 0
		if payment.AmountSentMsat != nil {
			amount = int64(*payment.AmountSentMsat)
			if payment.AmountMsat != nil {
				fee = amount - int64(*payment.AmountMsat)
			}
		}
		var settledAt *int64
		if payment.CompletedAt != nil {
			completedAt := int64(*payment.CompletedAt)
			settledAt = &completedAt
		}

		transactions = append(transactions, lnclient.Transaction{
			Type:            "outgoing",
			Invoice:         bolt11,
			Description:     description,
			DescriptionHash: descriptionHash,
			Preimage:        preimage,
			PaymentHash:     payment.PaymentHash,
			ExpiresAt:       expiresAt,
			Amount:          amount,
			FeesPaid:        fee,
			CreatedAt:       int64(payment.CreatedAt),
			SettledAt:       settledAt,
		})
	}

	// sort by created date descending
	sort.SliceStable(transactions, func(i, j int) bool {
		return transactions[i].CreatedAt > transactions[j].CreatedAt
	})

	return transactions, nil
}

func (gs *GreenlightService) GetInfo(ctx context.Context) (info *lnclient.NodeInfo, err error) {
	nodeInfo, err := gs.client.GetInfo()

	if err != nil {
		gs.svc.Logger.Errorf("GetInfo failed: %v", err)
		return nil, err
	}

	return &lnclient.NodeInfo{
		Alias:       nodeInfo.Alias,
		Color:       nodeInfo.Color,
		Pubkey:      nodeInfo.Pubkey,
		Network:     nodeInfo.Network,
		BlockHeight: nodeInfo.BlockHeight,
		BlockHash:   "",
	}, nil
}

func (gs *GreenlightService) ListChannels(ctx context.Context) ([]lnclient.Channel, error) {
	response, err := gs.client.ListFunds(glalby.ListFundsRequest{})

	if err != nil {
		gs.svc.Logger.Errorf("Failed to list funds: %v", err)
		return nil, err
	}

	channels := []lnclient.Channel{}

	for _, glChannel := range response.Channels {
		if glChannel.ChannelId == nil {
			continue
		}

		var localAmount uint64
		if glChannel.OurAmountMsat != nil {
			localAmount = *glChannel.OurAmountMsat
		}
		var totalAmount uint64
		if glChannel.AmountMsat != nil {
			totalAmount = *glChannel.AmountMsat
		}

		channels = append(channels, lnclient.Channel{
			LocalBalance:  int64(localAmount),
			RemoteBalance: int64(totalAmount - localAmount),
			RemotePubkey:  glChannel.PeerId,
			Id:            *glChannel.ChannelId,
			Active:        glChannel.State == 2,
			FundingTxId:   glChannel.FundingTxid,
			// TODO: add Public property
		})
	}

	return channels, nil
}

func (gs *GreenlightService) GetNodeConnectionInfo(ctx context.Context) (nodeConnectionInfo *lnclient.NodeConnectionInfo, err error) {
	info, err := gs.GetInfo(ctx)
	if err != nil {
		gs.svc.Logger.Errorf("GetInfo failed: %v", err)
		return nil, err
	}
	return &lnclient.NodeConnectionInfo{
		Pubkey: info.Pubkey,
	}, nil
}

func (gs *GreenlightService) ConnectPeer(ctx context.Context, connectPeerRequest *lnclient.ConnectPeerRequest) error {
	var host *string
	if connectPeerRequest.Address != "" {
		host = &connectPeerRequest.Address
	}
	var port *uint16
	if connectPeerRequest.Port > 0 {
		port = &connectPeerRequest.Port
	}
	_, err := gs.client.ConnectPeer(glalby.ConnectPeerRequest{
		Id:   connectPeerRequest.Pubkey,
		Host: host,
		Port: port,
	})
	if err != nil {
		gs.svc.Logger.Errorf("ConnectPeer failed: %v", err)
		return err
	}
	return nil
}

func (gs *GreenlightService) OpenChannel(ctx context.Context, openChannelRequest *lnclient.OpenChannelRequest) (*lnclient.OpenChannelResponse, error) {

	amountMsat := uint64(openChannelRequest.Amount) * 1000
	// minConf := uint32(0) //
	response, err := gs.client.FundChannel(glalby.FundChannelRequest{
		Id:         openChannelRequest.Pubkey,
		AmountMsat: &amountMsat,
		Announce:   &openChannelRequest.Public,
		// Minconf:    &minConf,
	})
	if err != nil {
		gs.svc.Logger.Errorf("OpenChannel failed: %v", err)
		return nil, err
	}

	return &lnclient.OpenChannelResponse{
		FundingTxId: response.Txid,
	}, nil
}

func (gs *GreenlightService) CloseChannel(ctx context.Context, closeChannelRequest *lnclient.CloseChannelRequest) (*lnclient.CloseChannelResponse, error) {
	_, err := gs.client.Close(glalby.CloseRequest{
		Id: closeChannelRequest.ChannelId,
	})
	if err != nil {
		gs.svc.Logger.WithError(err).Error("CloseChannel failed")
		return nil, err
	}

	return &lnclient.CloseChannelResponse{}, nil
}

func (gs *GreenlightService) GetNewOnchainAddress(ctx context.Context) (string, error) {

	newAddressResponse, err := gs.client.NewAddress(glalby.NewAddressRequest{})
	if err != nil {
		gs.svc.Logger.Errorf("NewAddress failed: %v", err)
		return "", err
	}
	if newAddressResponse.Bech32 == nil {
		return "", errors.New("no Bech32 in new address response")
	}

	return *newAddressResponse.Bech32, nil
}

func (gs *GreenlightService) GetOnchainBalance(ctx context.Context) (*lnclient.OnchainBalanceResponse, error) {
	response, err := gs.client.ListFunds(glalby.ListFundsRequest{})
	gs.svc.Logger.WithField("response", response).Info("Listed funds")

	if err != nil {
		gs.svc.Logger.Errorf("Failed to list funds: %v", err)
		return nil, err
	}

	var spendableBalance int64 = 0
	var pendingBalance int64 = 0
	for _, output := range response.Outputs {
		if output.AmountMsat != nil {
			if output.Status == 1 {
				spendableBalance += int64(*output.AmountMsat)
			} else {
				pendingBalance += int64(*output.AmountMsat)
			}
		}
	}

	return &lnclient.OnchainBalanceResponse{
		Spendable: spendableBalance / 1000,
		Total:     (spendableBalance + pendingBalance) / 1000,
	}, nil
}

func (gs *GreenlightService) RedeemOnchainFunds(ctx context.Context, toAddress string) (string, error) {
	amountAll := glalby.AmountOrAll(glalby.AmountOrAllAll{})
	txId, err := gs.client.Withdraw(glalby.WithdrawRequest{
		Destination: toAddress,
		Amount:      &amountAll,
	})
	if err != nil {
		gs.svc.Logger.WithError(err).Error("Withdraw failed")
		return "", err
	}
	gs.svc.Logger.WithField("txId", txId).Info("Redeeming On-Chain funds")

	return txId.Txid, nil
}

func (gs *GreenlightService) SendPaymentProbes(ctx context.Context, invoice string) error {
	return nil
}

func (gs *GreenlightService) SendSpontaneousPaymentProbes(ctx context.Context, amountMsat uint64, nodeId string) error {
	return nil
}

func (gs *GreenlightService) ListPeers(ctx context.Context) ([]lnclient.PeerDetails, error) {
	return nil, nil
}

func (gs *GreenlightService) GetLogOutput(ctx context.Context, maxLen int) ([]byte, error) {
	return []byte{}, nil
}

func (gs *GreenlightService) SignMessage(ctx context.Context, message string) (string, error) {
	response, err := gs.client.SignMessage(glalby.SignMessageRequest{
		Message: message,
	})

	if err != nil {
		gs.svc.Logger.Errorf("SignMessage failed: %v", err)
		return "", err
	}

	return response.Zbase, nil
}

func (gs *GreenlightService) greenlightInvoiceToTransaction(invoice *glalby.ListInvoicesInvoice) (*Nip47Transaction, error) {
	description := ""
	descriptionHash := ""
	if invoice.Description != nil {
		description = *invoice.Description
	}
	bolt11 := *invoice.Bolt11
	paymentRequest, err := decodepay.Decodepay(strings.ToLower(bolt11))
	if err != nil {
		gs.svc.Logger.WithFields(logrus.Fields{
			"invoice": bolt11,
		}).Errorf("Failed to decode bolt11 invoice: %v", bolt11)
		return nil, err
	}
	if description == "" {
		description = paymentRequest.Description
	}
	descriptionHash = paymentRequest.DescriptionHash

	expiresAt := int64(invoice.ExpiresAt)

	var amount int64 = 0
	var fee int64 = 0
	if invoice.AmountReceivedMsat != nil {
		amount = int64(*invoice.AmountReceivedMsat)
		if invoice.AmountMsat != nil {
			fee = int64(*invoice.AmountMsat) - amount
		}
	}

	preimage := ""
	var settledAt *int64
	if invoice.PaidAt != nil {
		preimage = *invoice.PaymentPreimage
		paidAt := int64(*invoice.PaidAt)
		settledAt = &paidAt
	}

	transaction := &Nip47Transaction{
		Type:            "incoming",
		Invoice:         bolt11,
		Description:     description,
		DescriptionHash: descriptionHash,
		PaymentHash:     invoice.PaymentHash,
		ExpiresAt:       &expiresAt,
		Amount:          amount,
		FeesPaid:        fee,
		CreatedAt:       int64(paymentRequest.CreatedAt),
		Preimage:        preimage,
		SettledAt:       settledAt,
	}
	return transaction, nil
}

func (gs *GreenlightService) ResetRouter(ctx context.Context) error {
	return nil
}

func (gs *GreenlightService) GetBalances(ctx context.Context) (*lnclient.BalancesResponse, error) {
	onchainBalance, err := gs.GetOnchainBalance(ctx)
	if err != nil {
		gs.svc.Logger.WithError(err).Error("Failed to retrieve onchain balance")
		return nil, err
	}

	response, err := gs.client.ListFunds(glalby.ListFundsRequest{})

	if err != nil {
		gs.svc.Logger.Errorf("Failed to list funds: %v", err)
		return nil, err
	}

	var totalReceivable int64 = 0
	var totalSpendable int64 = 0
	var nextMaxReceivable int64 = 0
	var nextMaxSpendable int64 = 0
	var nextMaxReceivableMPP int64 = 0
	var nextMaxSpendableMPP int64 = 0
	for _, channel := range response.Channels {
		if channel.OurAmountMsat != nil && channel.AmountMsat != nil && channel.Connected {

			channelReceivable := int64(*channel.AmountMsat - *channel.OurAmountMsat)
			channelSpendable := int64(*channel.OurAmountMsat)

			nextMaxReceivable = max(nextMaxReceivable, channelReceivable)
			nextMaxSpendable = max(nextMaxSpendable, channelSpendable)

			nextMaxReceivableMPP += channelReceivable
			nextMaxSpendableMPP += channelSpendable

			totalReceivable += channelReceivable
			totalSpendable += channelSpendable
		}
	}

	return &lnclient.BalancesResponse{
		Onchain: *onchainBalance,
		Lightning: lnclient.LightningBalanceResponse{
			TotalSpendable:       totalSpendable,
			TotalReceivable:      totalReceivable,
			NextMaxSpendable:     nextMaxSpendable,
			NextMaxReceivable:    nextMaxReceivable,
			NextMaxSpendableMPP:  nextMaxSpendableMPP,
			NextMaxReceivableMPP: nextMaxReceivableMPP,
		},
	}, nil

}

func (gs *GreenlightService) GetNodeStatus(ctx context.Context) (nodeStatus *lnclient.NodeStatus, err error) {
	return nil, nil
}
