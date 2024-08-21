import {
  Bell,
  CirclePlus,
  HandCoins,
  Info,
  LucideIcon,
  NotebookTabs,
  PenLine,
  Search,
  WalletMinimal,
} from "lucide-react";

export type BackendType =
  | "LND"
  | "BREEZ"
  | "GREENLIGHT"
  | "LDK"
  | "PHOENIX"
  | "CASHU";

export type Nip47RequestMethod =
  | "get_info"
  | "get_balance"
  | "make_invoice"
  | "pay_invoice"
  | "pay_keysend"
  | "lookup_invoice"
  | "list_transactions"
  | "sign_message"
  | "multi_pay_invoice"
  | "multi_pay_keysend";

export type BudgetRenewalType =
  | "daily"
  | "weekly"
  | "monthly"
  | "yearly"
  | "never"
  | "";

export type Scope =
  | "pay_invoice" // also used for pay_keysend, multi_pay_invoice, multi_pay_keysend
  | "get_balance"
  | "get_info"
  | "make_invoice"
  | "lookup_invoice"
  | "list_transactions"
  | "sign_message"
  | "notifications"; // covers all notification types

export type Nip47NotificationType = "payment_received" | "payment_sent";

export type ScopeIconMap = {
  [key in Scope]: LucideIcon;
};

export const scopeIconMap: ScopeIconMap = {
  get_balance: WalletMinimal,
  get_info: Info,
  list_transactions: NotebookTabs,
  lookup_invoice: Search,
  make_invoice: CirclePlus,
  pay_invoice: HandCoins,
  sign_message: PenLine,
  notifications: Bell,
};

export type WalletCapabilities = {
  methods: Nip47RequestMethod[];
  scopes: Scope[];
  notificationTypes: Nip47NotificationType[];
};

export const validBudgetRenewals: BudgetRenewalType[] = [
  "daily",
  "weekly",
  "monthly",
  "yearly",
  "never",
];

export const scopeDescriptions: Record<Scope, string> = {
  get_balance: "Read your balance",
  get_info: "Read your node info",
  list_transactions: "Read transaction history",
  lookup_invoice: "Lookup status of invoices",
  make_invoice: "Create invoices",
  pay_invoice: "Send payments",
  sign_message: "Sign messages",
  notifications: "Receive wallet notifications",
};

export const expiryOptions: Record<string, number> = {
  "1 week": 7,
  "1 month": 30,
  "1 year": 365,
  Never: 0,
};

export const budgetOptions: Record<string, number> = {
  "10k": 10_000,
  "100k": 100_000,
  "1M": 1_000_000,
  Unlimited: 0,
};

export interface ErrorResponse {
  message: string;
}

export interface App {
  id: number;
  name: string;
  description: string;
  nostrPubkey: string;
  createdAt: string;
  updatedAt: string;
  lastEventAt?: string;
  expiresAt?: string;
  isolated: boolean;
  balance: number;

  scopes: Scope[];
  maxAmount: number;
  budgetUsage: number;
  budgetRenewal: BudgetRenewalType;
}

export interface AppPermissions {
  scopes: Scope[];
  maxAmount: number;
  budgetRenewal: BudgetRenewalType;
  expiresAt?: Date;
  isolated: boolean;
}

export interface InfoResponse {
  backendType: BackendType;
  setupCompleted: boolean;
  oauthRedirect: boolean;
  albyAccountConnected: boolean;
  running: boolean;
  albyAuthUrl: string;
  nextBackupReminder: string;
  albyUserIdentifier: string;
  network?: Network;
  version: string;
  unlocked: boolean;
  enableAdvancedSetup: boolean;
}

export type Network = "bitcoin" | "testnet" | "signet";

export interface MnemonicResponse {
  mnemonic: string;
}

export interface CreateAppRequest {
  name: string;
  pubkey: string;
  maxAmount: number;
  budgetRenewal: string;
  expiresAt: string | undefined;
  scopes: Scope[];
  returnTo: string;
  isolated: boolean;
}

export interface CreateAppResponse {
  id: number;
  name: string;
  pairingUri: string;
  pairingPublicKey: string;
  pairingSecretKey: string;
  returnTo: string;
}

export type UpdateAppRequest = {
  maxAmount: number;
  budgetRenewal: string;
  expiresAt: string | undefined;
  scopes: Scope[];
};

export type Channel = {
  localBalance: number;
  localSpendableBalance: number;
  remoteBalance: number;
  remotePubkey: string;
  id: string;
  fundingTxId: string;
  active: boolean;
  public: boolean;
  confirmations?: number;
  confirmationsRequired?: number;
  forwardingFeeBaseMsat: number;
  unspendablePunishmentReserve: number;
  counterpartyUnspendablePunishmentReserve: number;
  error?: string;
  status: "online" | "opening" | "offline";
  isOutbound: boolean;
};

export type UpdateChannelRequest = {
  forwardingFeeBaseMsat: number;
};

export type Peer = {
  nodeId: string;
  address: string;
  isPersisted: boolean;
  isConnected: boolean;
};

export type NodeConnectionInfo = {
  pubkey: string;
  address: string;
  port: number;
};

export type ConnectPeerRequest = {
  pubkey: string;
  address: string;
  port: number;
};

export type SignMessageRequest = {
  message: string;
};

export type SignMessageResponse = {
  message: string;
  signature: string;
};

export type PayInvoiceResponse = {
  preimage: string;
  fee: number;
};

export type CreateInvoiceRequest = {
  amount: number;
  description: string;
};

export type OpenChannelRequest = {
  pubkey: string;
  amount: number;
  public: boolean;
};

export type OpenChannelResponse = {
  fundingTxId: string;
};

// eslint-disable-next-line @typescript-eslint/ban-types
export type CloseChannelResponse = {};

export type OnchainBalanceResponse = {
  spendable: number;
  total: number;
  reserved: number;
};

// from https://mempool.space/docs/api/rest#get-node-stats
export type Node = {
  alias: string;
  public_key: string;
  color: string;
  active_channel_count: number;
  sockets: string;
};
export type SetupNodeInfo = Partial<{
  backendType: BackendType;

  mnemonic?: string;
  nextBackupReminder?: string;
  greenlightInviteCode?: string;
  breezApiKey?: string;

  lndAddress?: string;
  lndCertHex?: string;
  lndMacaroonHex?: string;

  phoenixdAddress?: string;
  phoenixdAuthorization?: string;
}>;

export type LSPType = "LSPS1";

export type RecommendedChannelPeer = {
  network: Network;
  image: string;
  name: string;
  minimumChannelSize: number;
  maximumChannelSize: number;
} & (
  | {
      paymentMethod: "onchain";
      pubkey: string;
      host: string;
    }
  | {
      paymentMethod: "lightning";
      lspType: LSPType;
      lspUrl: string;
      pubkey?: string;
    }
);

// TODO: move to different file
export type AlbyMe = {
  identifier: string;
  nostr_pubkey: string;
  lightning_address: string;
  email: string;
  name: string;
  avatar: string;
  keysend_pubkey: string;
  shared_node: boolean;
  hub: {
    latest_version: string;
    name?: string;
  };
};

export type AlbyBalance = {
  sats: number;
};

export type LSPOrderRequest = {
  amount: number;
  lspType: LSPType;
  lspUrl: string;
  public: boolean;
};

export type LSPOrderResponse = {
  invoice: string;
  fee: number;
  invoiceAmount: number;
  incomingLiquidity: number;
  outgoingLiquidity: number;
};

export type AutoChannelRequest = {
  isPublic: boolean;
};
export type AutoChannelResponse = {
  invoice?: string;
  fee?: number;
  channelSize: number;
};

export type RedeemOnchainFundsResponse = {
  txId: string;
};

export type LightningBalanceResponse = {
  totalSpendable: number;
  totalReceivable: number;
  nextMaxSpendable: number;
  nextMaxReceivable: number;
  nextMaxSpendableMPP: number;
  nextMaxReceivableMPP: number;
};

export type BalancesResponse = {
  onchain: OnchainBalanceResponse;
  lightning: LightningBalanceResponse;
};

export type Transaction = {
  type: "incoming" | "outgoing";
  appId: number | undefined;
  invoice: string;
  description: string;
  descriptionHash: string;
  preimage: string | undefined;
  paymentHash: string;
  amount: number;
  feesPaid: number;
  createdAt: string;
  settledAt: string | undefined;
  metadata?: Record<string, unknown>;
  boostagram?: Boostagram;
};

export type Boostagram = {
  appName: string;
  name: string;
  podcast: string;
  url: string;
  episode?: string;
  feedId?: string;
  itemId?: string;
  ts?: number;
  message?: string;
  senderId: string;
  senderName: string;
  time: string;
  action: "boost";
  valueMsatTotal: number;
};

export type NewChannelOrderStatus = "pay" | "paid" | "success" | "opening";

export type NewChannelOrder = {
  amount: string;
  isPublic: boolean;
  status: NewChannelOrderStatus;
  fundingTxId?: string;
  prevChannelIds: string[];
} & (
  | {
      paymentMethod: "onchain";
      pubkey: string;
      host: string;
    }
  | {
      paymentMethod: "lightning";
      lspType: LSPType;
      lspUrl: string;
    }
);

export type AuthTokenResponse = {
  token: string;
};
