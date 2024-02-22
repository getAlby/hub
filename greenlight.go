package main

import (
	"context"
	"errors"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"strconv"

	//"github.com/getAlby/nostr-wallet-connect/glalby" // for local development only
	"github.com/getAlby/glalby/glalby"

	"github.com/getAlby/nostr-wallet-connect/models/lnclient"
)

type GreenlightService struct {
	workdir string
	client  *glalby.BlockingGreenlightAlbyClient
	svc     *Service
	// hsmdCmd *exec.Cmd
}

func NewGreenlightService(svc *Service, mnemonic, inviteCode, workDir string) (result lnclient.LNClient, err error) {
	if mnemonic == "" || inviteCode == "" || workDir == "" {
		return nil, errors.New("one or more required greenlight configuration are missing")
	}

	//create dir if not exists
	newpath := filepath.Join(".", workDir)
	err = os.MkdirAll(newpath, os.ModePerm)
	if err != nil {
		log.Printf("Failed to create greenlight working dir: %v", err)
		return nil, err
	}

	var credentials *glalby.GreenlightCredentials
	// NOTE: mnemonic used for encryption
	existingDeviceCert, _ := svc.cfg.Get("GreenlightDeviceCert", mnemonic)
	existingDeviceKey, _ := svc.cfg.Get("GreenlightDeviceKey", mnemonic)

	if existingDeviceCert != "" && existingDeviceKey != "" {
		credentials = &glalby.GreenlightCredentials{
			DeviceCert: existingDeviceCert,
			DeviceKey:  existingDeviceKey,
		}
		svc.Logger.Info("Using saved greenlight credentials")
	}

	if credentials == nil {
		svc.Logger.Info("No greenlight credentials found, attempting to recover existing node")
		recoveredCredentials, err := glalby.Recover(mnemonic)
		credentials = &recoveredCredentials

		if err != nil {
			log.Printf("Failed to recover node: %v", err)

			log.Print("Trying to register instead...")
			log.Fatalf("TODO")
			/*err = gs.register(inviteCode)
			if err != nil {
				log.Fatalf("Failed to register new node")
			}*/
		}

		if credentials == nil || credentials.DeviceCert == "" || credentials.DeviceKey == "" {
			log.Fatalf("unexpected response from Recover")
		}
		// NOTE: mnemonic used for encryption
		svc.cfg.SetUpdate("GreenlightDeviceCert", credentials.DeviceCert, mnemonic)
		svc.cfg.SetUpdate("GreenlightDeviceKey", credentials.DeviceKey, mnemonic)
	}

	client, err := glalby.NewBlockingGreenlightAlbyClient(mnemonic, *credentials)

	if err != nil {
		log.Printf("Failed to create greenlight alby client: %v", err)
	}
	if client == nil {
		log.Fatalf("unexpected response from NewBlockingGreenlightAlbyClient")
	}

	gs := GreenlightService{
		workdir: newpath,
		client:  client,
		svc:     svc,
	}

	//gs.hsmdCmd = gs.createCommand("hsmd")

	// if err := gs.hsmdCmd.Start(); err != nil {
	// 	log.Fatalf("Failed to start hsmd: %v", err)
	// }

	/*nodeInfo := models.NodeInfo{}
	err = gs.execJSONCommand(&nodeInfo, "getinfo")*/
	nodeInfo, err := client.GetInfo()

	if err != nil {
		return nil, err
	}

	log.Printf("Node info: %v", nodeInfo)

	return &gs, nil
}

func (gs *GreenlightService) Shutdown() error {
	return nil
}

func (gs *GreenlightService) SendPaymentSync(ctx context.Context, payReq string) (preimage string, err error) {
	response, err := gs.client.Pay(glalby.PayRequest{
		Bolt11: payReq,
	})

	if err != nil {
		gs.svc.Logger.Errorf("Failed to send payment: %v", err)
		return "", err
	}
	log.Printf("SendPaymentSync succeeded: %v", response.PaymentPreimage)
	return response.PaymentPreimage, nil
}

func (gs *GreenlightService) SendKeysend(ctx context.Context, amount int64, destination, preimage string, custom_records []lnclient.TLVRecord) (preImage string, err error) {
	/*if len(custom_records) > 0 {
		log.Printf("FIXME: TLVs not supported with CLI")
	}

	payResponse := models.PayResponse{}
	err = gs.execJSONCommand(&payResponse, "keysend", destination, strconv.FormatInt(amount, 10)+"msat")
	if err != nil {
		log.Printf("Keysend failed: %v", err)
		return "", err
	}
	log.Printf("Keysend succeeded: %v", payResponse.Preimage)

	return payResponse.Preimage, nil*/
	return "", errors.New("TODO")
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
	invoice, err := gs.client.MakeInvoice(glalby.MakeInvoiceRequest{
		AmountMsat:  uint64(amount),
		Description: description,
		Label:       "label_" + strconv.Itoa(rand.Int()),
		// TODO: other fields
	})

	if err != nil {
		log.Printf("MakeInvoice failed: %v", err)
		return nil, err
	}

	// TODO: add missing fields
	transaction = &Nip47Transaction{
		Type:    "incoming",
		Invoice: invoice.Bolt11,
		//PaymentHash: invoice.PaymentHash,
		Amount: amount,
		//CreatedAt:   time.Now().Unix(),
		//ExpiresAt:   &invoice.ExpiresAt,
	}

	return transaction, nil
}

func (gs *GreenlightService) LookupInvoice(ctx context.Context, paymentHash string) (transaction *Nip47Transaction, err error) {
	log.Println("TODO: LookupInvoice")
	return nil, errors.New("TODO: LookupInvoice")
}

func (gs *GreenlightService) ListTransactions(ctx context.Context, from, until, limit, offset uint64, unpaid bool, invoiceType string) (transactions []Nip47Transaction, err error) {
	log.Println("TODO: ListTransactions")

	transactions = []Nip47Transaction{}

	/*//glcli listinvoices
	listInvoicesResponse := models.ListInvoicesResponse{}
	err = gs.execJSONCommand(&listInvoicesResponse, "listinvoices")
	if err != nil {
		log.Printf("ListInvoices failed: %v", err)
		return nil, err
	}

	for _, glInvoice := range listInvoicesResponse.Invoices {
		var createdAt int64
		description := glInvoice.Description
		descriptionHash := ""

		if glInvoice.Bolt11 != "" {
			// TODO: Greenlight should provide these details so we don't need to manually decode the invoice
			paymentRequest, err := decodepay.Decodepay(strings.ToLower(glInvoice.Bolt11))
			if err != nil {
				log.Printf("Failed to decode bolt11 invoice: %v", glInvoice.Bolt11)
				return nil, err
			}

			createdAt = int64(paymentRequest.CreatedAt)
			description = paymentRequest.Description
			descriptionHash = paymentRequest.DescriptionHash
		}

		var fee int64 = 0
		if glInvoice.AmountReceivedMsat != nil {
			fee = glInvoice.AmountMsat.Msat - glInvoice.AmountReceivedMsat.Msat
		}

		transactions = append(transactions, lnclient.Transaction{
			Type:            "incoming",
			Invoice:         glInvoice.Bolt11,
			Description:     description,
			DescriptionHash: descriptionHash,
			Preimage:        glInvoice.Preimage,
			PaymentHash:     glInvoice.PaymentHash,
			ExpiresAt:       &glInvoice.ExpiresAt,
			Amount:          glInvoice.AmountMsat.Msat,
			FeesPaid:        fee,
			CreatedAt:       createdAt,
			SettledAt:       glInvoice.PaidAt,
		})
	}

	//glcli listpays
	listPaymentsResponse := models.ListPaymentsResponse{}
	err = gs.execJSONCommand(&listPaymentsResponse, "listpays")
	if err != nil {
		log.Printf("ListPayments failed: %v", err)
		return nil, err
	}

	for _, glPayment := range listPaymentsResponse.Payments {
		description := ""
		descriptionHash := ""
		var expiresAt *int64
		if glPayment.Bolt11 != "" {
			// TODO: Greenlight should provide these details so we don't need to manually decode the invoice
			paymentRequest, err := decodepay.Decodepay(strings.ToLower(glPayment.Bolt11))
			if err != nil {
				log.Printf("Failed to decode bolt11 invoice: %v", glPayment.Bolt11)
				return nil, err
			}

			description = paymentRequest.Description
			descriptionHash = paymentRequest.DescriptionHash
			expiresAtUnix := time.UnixMilli(int64(paymentRequest.CreatedAt) * 1000).Add(time.Duration(paymentRequest.Expiry) * time.Second).Unix()
			expiresAt = &expiresAtUnix
		}

		transactions = append(transactions, lnclient.Transaction{
			Type:            "outgoing",
			Invoice:         glPayment.Bolt11,
			Description:     description,
			DescriptionHash: descriptionHash,
			Preimage:        glPayment.Preimage,
			PaymentHash:     glPayment.PaymentHash,
			ExpiresAt:       expiresAt,
			Amount:          glPayment.AmountMsat.Msat,
			FeesPaid:        glPayment.AmountSentMsat.Msat - glPayment.AmountMsat.Msat,
			CreatedAt:       glPayment.CreatedAt,
			SettledAt:       &glPayment.CompletedAt,
		})
	}

	// sort by created date descending
	sort.SliceStable(transactions, func(i, j int) bool {
		return transactions[i].CreatedAt > transactions[j].CreatedAt
	})*/

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
	// glcli connect pubkey host port
	/*connectPeerResponse := models.ConnectPeerResponse{}
	err := gs.execJSONCommand(&connectPeerResponse, "connect", connectPeerRequest.Pubkey, connectPeerRequest.Address, strconv.Itoa(connectPeerRequest.Port))
	if err != nil {
		log.Printf("ConnectPeer failed: %v", err)
		return err
	}

	return nil*/
	return errors.New("TODO")
}

func (gs *GreenlightService) OpenChannel(ctx context.Context, openChannelRequest *lnclient.OpenChannelRequest) (*lnclient.OpenChannelResponse, error) {

	// glcli fundchannel nodeid amount

	/*openChannelResponse := models.OpenChannelResponse{}
	err := gs.execJSONCommand(&openChannelResponse, "fundchannel", openChannelRequest.Pubkey, strconv.FormatInt(openChannelRequest.Amount*1000, 10)+"msat")
	if err != nil {
		log.Printf("OpenChannel failed: %v", err)
		return nil, err
	}

	return &lnclient.OpenChannelResponse{
		FundingTxId: openChannelResponse.TxId,
	}, nil*/
	return nil, errors.New("TODO")
}

func (gs *GreenlightService) GetNewOnchainAddress(ctx context.Context) (string, error) {
	// glcli newaddr

	/*newAddressResponse := models.NewAddressResponse{}
	err := gs.execJSONCommand(&newAddressResponse, "newaddr")
	if err != nil {
		log.Printf("GetNewOnchainAddress failed: %v", err)
		return "", err
	}

	return newAddressResponse.Bech32, nil*/
	return "", errors.New("TODO")
}

func (gs *GreenlightService) GetOnchainBalance(ctx context.Context) (int64, error) {
	response, err := gs.client.ListFunds(glalby.ListFundsRequest{})

	if err != nil {
		gs.svc.Logger.Errorf("Failed to list funds: %v", err)
		return 0, err
	}

	var balance int64 = 0
	for _, output := range response.Outputs {
		if output.AmountMsat != nil {
			balance += int64(*output.AmountMsat)
		}
	}

	return balance / 1000, nil
}
