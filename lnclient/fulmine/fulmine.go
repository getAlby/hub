package fulmine

import (
	"context"
	"errors"
	"strconv"

	pb "github.com/ArkLabsHQ/fulmine/api-spec/protobuf/gen/go/fulmine/v1" // Update with the actual import path

	"github.com/getAlby/hub/config"
	"github.com/getAlby/hub/lnclient"
	"github.com/getAlby/hub/logger"
	"google.golang.org/grpc"
)

const nodeCommandGetOnboardAddress = "get_onboard_address"
const nodeCommandGetAddress = "get_address"
const nodeCommandSend = "send"

type FulmineService struct {
	workDir string
	conn    *grpc.ClientConn
	client  pb.ServiceClient
}

// Experimental ArkLabs Fulmine client
func NewFulmineService(ctx context.Context, cfg config.Config, workDir string) (result lnclient.LNClient, err error) {
	if workDir == "" {
		return nil, errors.New("one or more required fulmine configuration are missing")
	}

	conn, err := grpc.NewClient("localhost:7000", grpc.WithInsecure()) // Use secure connection in production
	if err != nil {
		logger.Logger.Fatalf("did not connect: %v", err)
	}

	// Create a new client
	client := pb.NewServiceClient(conn) // Replace with the actual service client name

	svc := FulmineService{
		workDir: workDir,
		client:  client,
		conn:    conn,
	}

	return &svc, nil
}

func (svc *FulmineService) Shutdown() error {
	err := svc.conn.Close()
	if err != nil {
		return err
	}
	return nil
}

func (svc *FulmineService) SendPaymentSync(ctx context.Context, invoice string, amount *uint64) (response *lnclient.PayInvoiceResponse, err error) {
	return nil, errors.New("TODO")
}

func (svc *FulmineService) SendKeysend(ctx context.Context, amount uint64, destination string, custom_records []lnclient.TLVRecord, preimage string) (*lnclient.PayKeysendResponse, error) {
	return nil, errors.New("keysend not supported")
}

func (svc *FulmineService) MakeInvoice(ctx context.Context, amount int64, description string, descriptionHash string, expiry int64) (transaction *lnclient.Transaction, err error) {
	return nil, errors.New("TODO")
}

func (svc *FulmineService) LookupInvoice(ctx context.Context, paymentHash string) (transaction *lnclient.Transaction, err error) {
	return nil, errors.New("TODO")
}

func (svc *FulmineService) ListTransactions(ctx context.Context, from, until, limit, offset uint64, unpaid bool, invoiceType string) (transactions []lnclient.Transaction, err error) {
	return nil, errors.New("TODO")
}

func (svc *FulmineService) GetInfo(ctx context.Context) (info *lnclient.NodeInfo, err error) {
	return &lnclient.NodeInfo{}, nil
}

func (svc *FulmineService) ListChannels(ctx context.Context) ([]lnclient.Channel, error) {
	return nil, nil
}

func (svc *FulmineService) GetNodeConnectionInfo(ctx context.Context) (nodeConnectionInfo *lnclient.NodeConnectionInfo, err error) {
	return &lnclient.NodeConnectionInfo{}, nil
}

func (svc *FulmineService) ConnectPeer(ctx context.Context, connectPeerRequest *lnclient.ConnectPeerRequest) error {
	return nil
}

func (svc *FulmineService) OpenChannel(ctx context.Context, openChannelRequest *lnclient.OpenChannelRequest) (*lnclient.OpenChannelResponse, error) {
	return nil, nil
}

func (svc *FulmineService) CloseChannel(ctx context.Context, closeChannelRequest *lnclient.CloseChannelRequest) (*lnclient.CloseChannelResponse, error) {
	return nil, nil
}

func (svc *FulmineService) GetNewOnchainAddress(ctx context.Context) (string, error) {
	return "", nil
}

func (svc *FulmineService) GetOnchainBalance(ctx context.Context) (*lnclient.OnchainBalanceResponse, error) {
	return nil, errors.New("not supported")
}

func (svc *FulmineService) RedeemOnchainFunds(ctx context.Context, toAddress string, amount uint64, sendAll bool) (string, error) {
	return "", errors.New("not supported")
}

func (svc *FulmineService) ResetRouter(key string) error {
	return errors.New("not supported")
}

func (svc *FulmineService) SignMessage(ctx context.Context, message string) (string, error) {
	return "", errors.New("not supported")
}

func (svc *FulmineService) DisconnectPeer(ctx context.Context, peerId string) error {
	return errors.New("not supported")
}

func (svc *FulmineService) ListPeers(ctx context.Context) ([]lnclient.PeerDetails, error) {
	return nil, errors.New("not supported")
}
func (svc *FulmineService) GetLogOutput(ctx context.Context, maxLen int) ([]byte, error) {
	return nil, errors.New("not supported")
}

func (svc *FulmineService) GetStorageDir() (string, error) {
	return "", errors.New("not supported")
}
func (svc *FulmineService) GetNetworkGraph(ctx context.Context, nodeIds []string) (lnclient.NetworkGraphResponse, error) {
	return nil, errors.New("not supported")
}
func (svc *FulmineService) UpdateLastWalletSyncRequest() {}

func (svc *FulmineService) GetNodeStatus(ctx context.Context) (nodeStatus *lnclient.NodeStatus, err error) {
	return &lnclient.NodeStatus{
		IsReady: true,
	}, nil
}

func (svc *FulmineService) SendPaymentProbes(ctx context.Context, invoice string) error {
	return errors.New("not supported")
}

func (svc *FulmineService) SendSpontaneousPaymentProbes(ctx context.Context, amountMsat uint64, nodeId string) error {
	return errors.New("not supported")
}

func (svc *FulmineService) UpdateChannel(ctx context.Context, updateChannelRequest *lnclient.UpdateChannelRequest) error {
	return errors.New("not supported")
}

func (svc *FulmineService) GetBalances(ctx context.Context, includeInactiveChannels bool) (*lnclient.BalancesResponse, error) {
	// Example: Call a method (replace with actual method name and parameters)
	balance, err := svc.client.GetBalance(ctx, &pb.GetBalanceRequest{}) // Replace with actual request type
	if err != nil {
		return nil, err
	}

	balanceMsat := int64(balance.GetAmount()) * 1000

	return &lnclient.BalancesResponse{
		Onchain: lnclient.OnchainBalanceResponse{
			Spendable: int64(0),
			Total:     int64(0),
		},
		Lightning: lnclient.LightningBalanceResponse{
			TotalSpendable:       balanceMsat,
			TotalReceivable:      0,
			NextMaxSpendable:     balanceMsat,
			NextMaxReceivable:    0,
			NextMaxSpendableMPP:  balanceMsat,
			NextMaxReceivableMPP: 0,
		},
	}, nil
}

func (svc *FulmineService) GetSupportedNIP47Methods() []string {
	//return []string{"pay_invoice", "get_balance", "get_budget", "get_info", "make_invoice", "lookup_invoice", "list_transactions", "multi_pay_invoice"}
	return []string{"get_balance", "get_budget", "get_info"}
}

func (svc *FulmineService) GetSupportedNIP47NotificationTypes() []string {
	return []string{}
}

func (svc *FulmineService) GetPubkey() string {
	return ""
}

func (svc *FulmineService) GetCustomNodeCommandDefinitions() []lnclient.CustomNodeCommandDef {
	return []lnclient.CustomNodeCommandDef{
		{
			Name:        nodeCommandGetOnboardAddress,
			Description: "Return on-chain address to onboard.",
			Args:        nil,
		},
		{
			Name:        nodeCommandGetAddress,
			Description: "Return ark address.",
			Args:        nil,
		},
		{
			Name:        nodeCommandSend,
			Description: "Send funds natively on Ark.",
			Args: []lnclient.CustomNodeCommandArgDef{
				{
					Name:        "address",
					Description: "off-chain ark address",
				},
				{
					Name:        "amount",
					Description: "amount to send in sats",
				},
			},
		},
	}
}

func (svc *FulmineService) ExecuteCustomNodeCommand(ctx context.Context, command *lnclient.CustomNodeCommandRequest) (*lnclient.CustomNodeCommandResponse, error) {
	switch command.Name {
	case nodeCommandGetOnboardAddress:
		response, err := svc.client.GetOnboardAddress(ctx, &pb.GetOnboardAddressRequest{})
		if err != nil {
			return nil, err
		}
		return &lnclient.CustomNodeCommandResponse{
			Response: response,
		}, nil
	case nodeCommandGetAddress:
		response, err := svc.client.GetAddress(ctx, &pb.GetAddressRequest{})
		if err != nil {
			return nil, err
		}
		return &lnclient.CustomNodeCommandResponse{
			Response: response,
		}, nil

	case nodeCommandSend:
		var address string
		var amount uint64
		var err error
		for i := range command.Args {
			switch command.Args[i].Name {
			case "address":
				address = command.Args[i].Value
			case "amount":
				amount, err = strconv.ParseUint(string(command.Args[i].Value), 10, 64)
			}
		}
		if err != nil {
			return nil, err
		}

		response, err := svc.client.SendOffChain(ctx, &pb.SendOffChainRequest{
			Address: address,
			Amount:  amount,
		})

		if err != nil {
			return nil, err
		}

		return &lnclient.CustomNodeCommandResponse{
			Response: response,
		}, nil
	}

	return nil, lnclient.ErrUnknownCustomNodeCommand
}
