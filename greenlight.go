package main

import (
	"context"
	"errors"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"strconv"

	"github.com/getAlby/nostr-wallet-connect/glalby" // TODO: import from other repository
	"github.com/getAlby/nostr-wallet-connect/models/lnclient"
)

type GreenlightService struct {
	workdir string
	client  *glalby.BlockingGreenlightAlbyClient
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
		//listener: &listener,
		//svc:      svc,
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

func (gs *GreenlightService) register(inviteCode string) error {
	/*output, err := gs.execCommand("scheduler", "register", "--network=bitcoin", fmt.Sprintf("--invite=%s", inviteCode))
	log.Printf("scheduler register: %v %v", string(output), err)
	return err*/
	return errors.New("TODO")
}

func (gs *GreenlightService) Shutdown() error {
	log.Println("TODO: shut down greenlight client")

	return nil
	//return bs.svc.Disconnect()
}

func (gs *GreenlightService) SendPaymentSync(ctx context.Context, payReq string) (preimage string, err error) {
	//glcli pay BOLT11_INVOICE_HERE

	/*log.Printf("SendPaymentSync %v", payReq)
	payResponse := models.PayResponse{}
	err = gs.execJSONCommand(&payResponse, "pay", payReq)
	if err != nil {
		log.Printf("SendPaymentSync failed: %v", err)
		return "", err
	}
	log.Printf("SendPaymentSync succeeded: %v", payResponse.Preimage)

	return payResponse.Preimage, nil*/
	return "", errors.New("TODO")
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
	/*channels, err := gs.ListChannels(ctx)

	if err != nil {
		return 0, err
	}

	balance = 0
	for _, channel := range channels {
		balance += channel.LocalBalance
	}

	return balance, nil*/
	return 0, nil
}

func (gs *GreenlightService) MakeInvoice(ctx context.Context, amount int64, description string, descriptionHash string, expiry int64) (transaction *Nip47Transaction, err error) {
	invoice, err := gs.client.MakeInvoice(glalby.GreenlightInvoiceRequest{
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
		log.Printf("GetInfo failed: %v", err)
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
	//glcli listfunds

	/*listFundsResponse := models.ListFundsResponse{}
	err := gs.execJSONCommand(&listFundsResponse, "listfunds")
	if err != nil {
		log.Printf("ListChannels failed: %v", err)
		return nil, err
	}

	glChannels := listFundsResponse.Channels
	channels := []lnclient.Channel{}

	for _, glChannel := range glChannels {
		channels = append(channels, lnclient.Channel{
			LocalBalance:  glChannel.OurAmountMsat.Msat,
			RemoteBalance: glChannel.AmountMsat.Msat - glChannel.OurAmountMsat.Msat,
			RemotePubkey:  glChannel.PeerId,
			Id:            glChannel.Id,
			Active:        glChannel.State == 2,
		})
	}*/

	channels := []lnclient.Channel{}
	return channels, nil
}

func (gs *GreenlightService) GetNodeConnectionInfo(ctx context.Context) (nodeConnectionInfo *lnclient.NodeConnectionInfo, err error) {
	// glcli scheduler schedule
	/*scheduleResponse := models.ScheduleResponse{}
	err = gs.execJSONCommand(&scheduleResponse, "scheduler", "schedule")
	if err != nil {
		log.Printf("GetNodeConnectionInfo failed: %v", err)
		return nil, err
	}

	return &lnclient.NodeConnectionInfo{
		Pubkey:  scheduleResponse.NodeId,
		Address: strings.ReplaceAll(scheduleResponse.GrpcUri, "https://", ""),
		Port:    9735, // TODO: why doesn't greenlight return this?
	}, nil*/
	return nil, errors.New("TODO")
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
	//glcli listfunds

	/*listFundsResponse := models.ListFundsResponse{}
	err := gs.execJSONCommand(&listFundsResponse, "listfunds")
	if err != nil {
		log.Printf("GetOnchainBalance failed: %v", err)
		return 0, err
	}

	var balance int64 = 0
	for _, output := range listFundsResponse.Outputs {
		balance += output.AmountMsat.Msat
	}

	return balance / 1000, nil*/
	return 0, nil
}
