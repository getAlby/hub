package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"time"

	"github.com/breez/breez-sdk-go/breez_sdk"
	models "github.com/getAlby/nostr-wallet-connect/models/greenlight"
)

type GreenlightService struct {
	workdir string
	hsmdCmd *exec.Cmd
	//svc      *breez_sdk.BlockingGreenlightServices
}

func NewGreenlightService(mnemonic, inviteCode, workDir string) (result LNClient, err error) {
	if mnemonic == "" || inviteCode == "" || workDir == "" {
		return nil, errors.New("One or more required greenlight configuration are missing")
	}

	//create dir if not exists
	newpath := filepath.Join(".", workDir)
	err = os.MkdirAll(newpath, os.ModePerm)
	if err != nil {
		log.Printf("Failed to create greenlight working dir: %v", err)
		return nil, err
	}
	seed, err := breez_sdk.MnemonicToSeed(mnemonic)
	if err != nil {
		log.Printf("Failed to convert mnemonic to seed: %v", err)
		return nil, err
	}

	hsmSecretPath := filepath.Join(newpath, "hsm_secret")
	err = os.WriteFile(hsmSecretPath, seed[0:32], 0644)
	if err != nil {
		log.Printf("Failed to write hsm secret: %v", err)
		return nil, err
	}

	gs := GreenlightService{
		workdir: newpath,
		//listener: &listener,
		//svc:      svc,
	}

	err = gs.recover()

	if err != nil {
		log.Printf("Failed to recover node: %v", err)
		log.Print("Trying to register instead...")
		err = gs.register(inviteCode)
		if err != nil {
			log.Fatalf("Failed to register new node")
		}
	}

	nodeInfo := models.NodeInfo{}
	err = gs.execJSONCommand(&nodeInfo, "getinfo")
	if err != nil {
		return nil, err
	}
	if err == nil {
		log.Printf("Node info: %v", nodeInfo)
	}

	// TODO: schedule

	gs.hsmdCmd = gs.createCommand("hsmd")

	if err := gs.hsmdCmd.Start(); err != nil {
		log.Fatalf("Failed to start hsmd: %v", err)
	}

	return &gs, nil
}

func (gs *GreenlightService) recover() error {
	output, err := gs.execCommand("scheduler", "recover")
	log.Printf("scheduler recover: %v %v", string(output), err)
	return err
}

func (gs *GreenlightService) register(inviteCode string) error {
	output, err := gs.execCommand("scheduler", "register", "--network=bitcoin", fmt.Sprintf("--invite=%s", inviteCode))
	log.Printf("scheduler register: %v %v", string(output), err)
	return err
}

func (gs *GreenlightService) createCommand(args ...string) *exec.Cmd {
	cmd := exec.Command("glcli", args...)
	cmd.Dir = gs.workdir
	return cmd
}

func (gs *GreenlightService) execCommand(args ...string) ([]byte, error) {
	cmd := gs.createCommand(args...)

	var outputBuffer bytes.Buffer
	var errorBuffer bytes.Buffer

	cmd.Stdout = &outputBuffer
	cmd.Stderr = &errorBuffer
	err := cmd.Run()

	if err != nil {
		errorOutput := errorBuffer.String()
		log.Printf("Failed to exec command %v: %v %v", args, err, errorOutput)
		return nil, err
	}
	output := outputBuffer.Bytes()
	return output, err
}

func (gs *GreenlightService) execJSONCommand(dest any, args ...string) error {
	output, err := gs.execCommand(args...)
	if err != nil {
		return err
	}

	err = json.Unmarshal(output, dest)
	if err != nil {
		log.Printf("Failed to unmarshal command output %v: %v", string(output), err)
	}
	return err
}

func (gs *GreenlightService) Shutdown() error {
	if gs.hsmdCmd != nil {
		if err := gs.hsmdCmd.Process.Kill(); err != nil {
			log.Printf("Failed to kill hsmd process: %v", err)
			return err
		}
	}

	return nil
	//return bs.svc.Disconnect()
}

func (gs *GreenlightService) SendPaymentSync(ctx context.Context, payReq string) (preimage string, err error) {
	//glcli pay BOLT11_INVOICE_HERE

	payResponse := models.PayResponse{}
	err = gs.execJSONCommand(&payResponse, "pay", payReq)
	if err != nil {
		log.Printf("SendPaymentSync failed: %v", err)
		return "", err
	}

	return payResponse.Preimage, nil
}

func (gs *GreenlightService) SendKeysend(ctx context.Context, amount int64, destination, preimage string, custom_records []TLVRecord) (preImage string, err error) {
	log.Println("TODO: SendKeysend")
	return "", nil
}

func (gs *GreenlightService) GetBalance(ctx context.Context) (balance int64, err error) {
	/*info, err := bs.svc.NodeInfo()
	if err != nil {
		return 0, err
	}
	return int64(info.ChannelsBalanceMsat) / 1000, nil*/
	log.Println("TODO: GetBalance")
	return 0, nil
}

func (gs *GreenlightService) MakeInvoice(ctx context.Context, amount int64, description string, descriptionHash string, expiry int64) (transaction *Nip47Transaction, err error) {
	//glcli invoice example_label3 21000msat

	invoice := models.Invoice{}
	err = gs.execJSONCommand(&invoice, "invoice", "label_"+strconv.Itoa(rand.Int()), strconv.FormatInt(amount, 10)+"msat")
	if err != nil {
		log.Println("MakeInvoice failed: %v", err)
		return nil, err
	}

	transaction = &Nip47Transaction{
		Type:        "incoming",
		Invoice:     invoice.Bolt11,
		PaymentHash: invoice.PaymentHash,
		Amount:      amount,
		CreatedAt:   time.Now().Unix(),
		ExpiresAt:   &invoice.ExpiresAt,
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
	return transactions, nil
}

func (gs *GreenlightService) GetInfo(ctx context.Context) (info *NodeInfo, err error) {
	log.Println("TODO: GetInfo")
	return &NodeInfo{
		Alias:       "greenlight",
		Color:       "",
		Pubkey:      "",
		Network:     "mainnet",
		BlockHeight: 0,
		BlockHash:   "",
	}, nil
}
