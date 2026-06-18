package http

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"

	"github.com/getAlby/hub/logger"
)

// ai_chat.go is a proof-of-concept read-only wallet assistant. It forwards the
// conversation to llm402.ai (an OpenAI-compatible LLM gateway).
//
// Payment uses the L402 (Lightning) flow: the gateway answers an unauthenticated
// request with 402 Payment Required + a bolt11 invoice and a macaroon. The Hub
// pays the invoice (its core capability), then retries with
// "Authorization: L402 <macaroon>:<preimage>". No API key or account is needed —
// the user's own node pays a few sats per inference call.
//
// llm402 does NOT support OpenAI function-calling (the `tools` parameter), so
// instead of letting the model call wallet tools we pre-fetch a read-only
// snapshot (balance, recent transactions, exchange rate) and inject it into the
// prompt. The only spend is the per-request inference fee, capped by config.
//
// For productionizing this should move into the api/ package (behind the API
// interface), stream responses, and meter inference spend as a budgeted app.

const aiChatSystemPrompt = `You are the assistant built into Alby Hub, a self-custodial Lightning wallet. Alby Hub IS the user's wallet — never tell them to "connect a wallet" or explain what Lightning is.

Answer questions about the user's wallet using the WALLET DATA provided in the next message. Do not invent balances or transactions — only use the data given. Amounts are in satoshis (sats) unless stated otherwise. Be concise, use the wallet's own vocabulary (sats, connections, apps), and format sat amounts with thousands separators.

You are READ-ONLY: you cannot send payments or create invoices. If the user asks you to send money, explain that this assistant can't move funds yet and they should use the Send screen.`

type aiChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type aiChatRequest struct {
	Messages []aiChatMessage `json:"messages"`
}

type aiChatResponse struct {
	Message string `json:"message"`
}

// l402Challenge is the JSON body returned alongside a 402 Payment Required.
type l402Challenge struct {
	Price    uint64 `json:"price"` // sats
	Invoice  string `json:"invoice"`
	Macaroon string `json:"macaroon"`
}

// --- OpenAI-compatible wire types ---

type openAIMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatCompletionRequest struct {
	Model    string          `json:"model"`
	Messages []openAIMessage `json:"messages"`
}

type chatCompletionResponse struct {
	Choices []struct {
		Message      openAIMessage `json:"message"`
		FinishReason string        `json:"finish_reason"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

func (httpSvc *HttpService) aiChatHandler(c echo.Context) error {
	ctx := c.Request().Context()

	var chatRequest aiChatRequest
	if err := c.Bind(&chatRequest); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Message: fmt.Sprintf("Bad request: %s", err.Error()),
		})
	}

	model := httpSvc.cfg.GetEnv().LLM402Model
	baseURL := httpSvc.cfg.GetEnv().LLM402BaseUrl

	messages := []openAIMessage{
		{Role: "system", Content: aiChatSystemPrompt},
		{Role: "system", Content: httpSvc.walletContext(ctx)},
	}
	for _, m := range chatRequest.Messages {
		messages = append(messages, openAIMessage{Role: m.Role, Content: m.Content})
	}

	completion, err := httpSvc.callLLM402Chat(ctx, baseURL, model, messages)
	if err != nil {
		logger.Logger.WithError(err).Error("AI chat completion request failed")
		return c.JSON(http.StatusBadGateway, ErrorResponse{
			Message: "Couldn't get a response from the AI provider.",
		})
	}
	if len(completion.Choices) == 0 {
		return c.JSON(http.StatusBadGateway, ErrorResponse{
			Message: "AI provider returned no response.",
		})
	}

	return c.JSON(http.StatusOK, aiChatResponse{
		Message: completion.Choices[0].Message.Content,
	})
}

// walletContext builds a compact read-only snapshot of the wallet for the model.
// Each piece is best-effort: a failure is noted rather than aborting the chat.
func (httpSvc *HttpService) walletContext(ctx context.Context) string {
	snapshot := map[string]interface{}{}

	if balances, err := httpSvc.api.GetBalances(ctx); err == nil {
		snapshot["balances"] = balances
	} else {
		snapshot["balancesError"] = err.Error()
	}

	if transactions, err := httpSvc.api.ListTransactions(ctx, nil, 5, 0); err == nil {
		snapshot["recentTransactions"] = transactions
	} else {
		snapshot["transactionsError"] = err.Error()
	}

	currency := httpSvc.cfg.GetCurrency()
	if currency == "" {
		currency = "USD"
	}
	if rate, err := httpSvc.albySvc.GetBitcoinRate(ctx, currency); err == nil && rate.RateFloat > 0 {
		snapshot["bitcoinPrice"] = map[string]interface{}{
			"currency":    currency,
			"pricePerBtc": rate.RateFloat,
			"satsPerUnit": 100_000_000 / rate.RateFloat,
		}
	}

	return "WALLET DATA (JSON):\n" + toJSONString(snapshot)
}

func toJSONString(v interface{}) string {
	bytes, err := json.Marshal(v)
	if err != nil {
		return fmt.Sprintf(`{"error":%q}`, err.Error())
	}
	return string(bytes)
}

// callLLM402Chat performs one chat completion using the L402 Lightning flow:
// try unauthenticated, pay the returned invoice, then retry with the preimage.
func (httpSvc *HttpService) callLLM402Chat(ctx context.Context, baseURL, model string, messages []openAIMessage) (*chatCompletionResponse, error) {
	requestBody, err := json.Marshal(chatCompletionRequest{
		Model:    model,
		Messages: messages,
	})
	if err != nil {
		return nil, err
	}

	// If a prepaid-balance token is configured, use it (handy for demoing on a
	// node that can't pay mainnet invoices). Otherwise pay per-request via L402.
	initialAuth := ""
	if token := httpSvc.cfg.GetEnv().LLM402Token; token != "" {
		initialAuth = "Bearer " + token
	}

	res, body, err := doChatRequest(ctx, baseURL, requestBody, initialAuth)
	if err != nil {
		return nil, err
	}

	if res.StatusCode == http.StatusPaymentRequired {
		authHeader, err := httpSvc.payL402Challenge(ctx, body)
		if err != nil {
			return nil, err
		}
		res, body, err = doChatRequest(ctx, baseURL, requestBody, authHeader)
		if err != nil {
			return nil, err
		}
	}

	if res.StatusCode >= 300 {
		logger.Logger.WithFields(logrus.Fields{
			"status_code": res.StatusCode,
			"body":        string(body),
		}).Error("AI provider returned non-success code")
		return nil, fmt.Errorf("AI provider returned status %d", res.StatusCode)
	}

	var completion chatCompletionResponse
	if err := json.Unmarshal(body, &completion); err != nil {
		return nil, fmt.Errorf("failed to decode AI response: %w", err)
	}
	if completion.Error != nil {
		return nil, fmt.Errorf("AI provider error: %s", completion.Error.Message)
	}

	return &completion, nil
}

// payL402Challenge parses the 402 body, pays the invoice (respecting the
// configured per-call cap), and returns the "L402 <macaroon>:<preimage>" header.
func (httpSvc *HttpService) payL402Challenge(ctx context.Context, body []byte) (string, error) {
	var challenge l402Challenge
	if err := json.Unmarshal(body, &challenge); err != nil {
		return "", fmt.Errorf("failed to parse 402 payment challenge: %w", err)
	}
	if challenge.Invoice == "" || challenge.Macaroon == "" {
		return "", fmt.Errorf("incomplete 402 payment challenge")
	}

	maxSats := httpSvc.cfg.GetEnv().LLM402MaxSatsPerCall
	if maxSats > 0 && challenge.Price > maxSats {
		return "", fmt.Errorf("inference price %d sats exceeds the configured cap of %d sats", challenge.Price, maxSats)
	}

	payment, err := httpSvc.api.SendPayment(ctx, challenge.Invoice, nil, map[string]interface{}{
		"comment": "AI assistant inference (llm402.ai)",
	}, nil)
	if err != nil {
		return "", fmt.Errorf("failed to pay inference invoice: %w", err)
	}
	if payment.Preimage == nil || *payment.Preimage == "" {
		return "", fmt.Errorf("inference payment did not return a preimage")
	}

	return fmt.Sprintf("L402 %s:%s", challenge.Macaroon, *payment.Preimage), nil
}

func doChatRequest(ctx context.Context, baseURL string, requestBody []byte, authHeader string) (*http.Response, []byte, error) {
	client := &http.Client{Timeout: 90 * time.Second}
	url := fmt.Sprintf("%s/chat/completions", baseURL)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	if authHeader != "" {
		req.Header.Set("Authorization", authHeader)
	}

	res, err := client.Do(req)
	if err != nil {
		return nil, nil, err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read response body: %w", err)
	}

	return res, body, nil
}
