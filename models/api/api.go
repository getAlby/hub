package api

import "time"

type App struct {
	// ID          uint      `json:"id"` // ID unused - pubkey is used as ID
	Name        string    `json:"name"`
	Description string    `json:"description"`
	NostrPubkey string    `json:"nostrPubkey"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`

	LastEventAt    *time.Time `json:"lastEventAt"`
	ExpiresAt      *time.Time `json:"expiresAt"`
	RequestMethods []string   `json:"requestMethods"`
	MaxAmount      int        `json:"maxAmount"`
	BudgetUsage    int64      `json:"budgetUsage"`
	BudgetRenewal  string     `json:"budgetRenewal"`
}

// type AppPermission struct {
// 	ID            uint      `json:"id"`
// 	AppId         uint      `json:"appId" validate:"required"`
// 	App           App       `json:"app"`
// 	RequestMethod string    `json:"requestMethod" validate:"required"`
// 	MaxAmount     int       `json:"maxAmount"`
// 	BudgetRenewal string    `json:"budgetRenewal"`
// 	ExpiresAt     time.Time `json:"expiresAt"`
// 	CreatedAt     time.Time `json:"createdAt"`
// 	UpdatedAt     time.Time `json:"updatedAt"`
// }

type ListAppsResponse struct {
	Apps []App `json:"apps"`
}

type CreateAppResponse struct {
	PairingUri    string `json:"pairingUri"`
	PairingSecret string `json:"pairingSecretKey"`
	Pubkey        string `json:"pairingPublicKey"`
	Name          string `json:"name"`
	ReturnTo      string `json:"returnTo"`
}

type User struct {
	Email string `json:"email"`
}

type InfoResponse struct {
	User        *User  `json:"user"`
	BackendType string `json:"backendType"`
}
