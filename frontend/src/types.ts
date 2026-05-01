import {
  BellIcon,
  CirclePlusIcon,
  CrownIcon,
  HandCoinsIcon,
  InfoIcon,
  LucideIcon,
  NotebookTabsIcon,
  PenLineIcon,
  SearchIcon,
  WalletMinimalIcon,
} from "lucide-react";

export type BackendType = "LND" | "LDK" | "PHOENIX" | "CASHU";

export type Nip47RequestMethod =
  | "get_info"
  | "get_balance"
  | "get_budget"
  | "make_invoice"
  | "pay_invoice"
  | "pay_keysend"
  | "lookup_invoice"
  | "list_transactions"
  | "sign_message"
  | "multi_pay_invoice"
  | "multi_pay_keysend"
  | "make_hold_invoice"
  | "settle_hold_invoice"
  | "cancel_hold_invoice";

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
  | "notifications" // covers all notification types
  | "superuser";

export type Nip47NotificationType = "payment_received" | "payment_sent";

export type ScopeIconMap = {
  [key in Scope]: LucideIcon;
};

export const scopeIconMap: ScopeIconMap = {
  get_balance: WalletMinimalIcon,
  get_info: InfoIcon,
  list_transactions: NotebookTabsIcon,
  lookup_invoice: SearchIcon,
  make_invoice: CirclePlusIcon,
  pay_invoice: HandCoinsIcon,
  sign_message: PenLineIcon,
  notifications: BellIcon,
  superuser: CrownIcon,
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
  superuser: "Create other app connections",
};

export const expiryOptions: Record<string, number> = {
  "1 week": 7,
  "1 month": 30,
  "1 year": 365,
};

export const budgetOptionsSat: Record<string, number> = {
  "10k": 10_000,
  "100k": 100_000,
  "1M": 1_000_000,
};

export interface ErrorResponse {
  message: string;
}

export interface App {
  id: number;
  name: string;
  description: string;
  appPubkey: string;
  uniqueWalletPubkey: boolean;
  walletPubkey: string;
  createdAt: string;
  updatedAt: string;
  lastUsedAt?: string;
  lastSettledTransactionAt?: string;
  expiresAt?: string;
  isolated: boolean;
  balanceSat: number;
  balanceMsat: number;

  scopes: Scope[];
  maxAmountSat: number;
  maxAmountMsat: number;
  budgetUsageSat: number;
  budgetUsageMsat: number;
  budgetRenewal: BudgetRenewalType;
  metadata?: AppMetadata;
}

export interface AppPermissions {
  scopes: Scope[];
  maxAmountSat: number;
  budgetRenewal: BudgetRenewalType;
  expiresAt?: Date;
  isolated: boolean;
}

export interface InfoResponse {
  backendType: BackendType;
  setupCompleted: boolean;
  oauthRedirect: boolean;
  albyAccountConnected: boolean;
  ldkVssEnabled: boolean;
  vssSupported: boolean;
  running: boolean;
  albyAuthUrl: string;
  nextBackupReminder: string;
  albyUserIdentifier: string;
  network?: Network;
  version: string;
  relays: { url: string; online: boolean }[];
  unlocked: boolean;
  enableAdvancedSetup: boolean;
  startupState: string;
  startupError: string;
  startupErrorTime: string;
  autoUnlockPasswordSupported: boolean;
  autoUnlockPasswordEnabled: boolean;
  currency: string;
  nodeAlias: string;
  mempoolUrl: string;
  bitcoinDisplayFormat: BitcoinDisplayFormat;
  chainDataSourceType?: string;
  chainDataSourceAddress?: string;
  hideUpdateBanner: boolean;
}

export type BitcoinDisplayFormat = "sats" | "bip177";

export type HealthAlarmKind =
  | "alby_service"
  | "node_not_ready"
  | "channels_offline"
  | "nostr_relay_offline"
  | "vss_no_subscription";

export type HealthAlarm = {
  kind: HealthAlarmKind;
  rawDetails?: unknown;
};
export type AlbyInfoIncident = {
  name: string;
  started: string;
  status: string;
  impact: string;
  url: string;
};

export type HealthResponse = {
  alarms: HealthAlarm[];
};

export type Network = "bitcoin" | "testnet" | "signet";

export type AppMetadata = {
  app_store_app_id?: string;
  lud16?: string;
} & Record<string, unknown>;

export type AutoSwapConfig = {
  type: "out";
  enabled: boolean;
  balanceThresholdSat: number;
  swapAmountSat: number;
  destination: string;
};

export type SwapInfo = {
  albyServiceFee: number;
  boltzServiceFee: number;
  boltzNetworkFeeSat: number;
  minAmountSat: number;
  maxAmountSat: number;
};

export type BaseSwap = {
  id: string;
  sendAmountSat: number;
  lockupAddress: string;
  paymentHash: string;
  invoice: string;
  autoSwap: boolean;
  usedXpub: boolean;
  boltzPubkey: string;
  createdAt: string;
  updatedAt: string;
  lockupTxId?: string;
  claimTxId?: string;
  receiveAmountSat?: number;
};

export type SwapIn = BaseSwap & {
  type: "in";
  state: "PENDING" | "SUCCESS" | "FAILED" | "REFUNDED";
  refundAddress?: string;
};

export type SwapOut = BaseSwap & {
  type: "out";
  state: "PENDING" | "SUCCESS" | "FAILED";
  destinationAddress: string;
};

export type Swap = SwapIn | SwapOut;

export type SwapResponse = {
  swapId: string;
  paymentHash: string;
};

export interface MnemonicResponse {
  mnemonic: string;
}

export interface CreateAppRequest {
  name: string;
  pubkey?: string;
  maxAmountSat?: number;
  maxAmountMsat?: number;
  budgetRenewal?: BudgetRenewalType;
  expiresAt?: string;
  scopes: Scope[];
  returnTo?: string;
  isolated?: boolean;
  metadata?: AppMetadata;
  unlockPassword?: string; // required to create superuser apps
}

export interface CreateAppResponse {
  id: number;
  name: string;
  pairingUri: string;
  pairingPublicKey: string;
  pairingSecretKey: string;
  relayUrls: string[];
  walletPubkey: string;
  lud16: string;
  returnTo: string;
}

export type UpdateAppRequest = {
  name?: string;
  maxAmountSat?: number;
  maxAmountMsat?: number;
  budgetRenewal?: string;
  expiresAt?: string | undefined;
  updateExpiresAt?: boolean;
  scopes?: Scope[];
  metadata?: AppMetadata;
  isolated?: boolean;
};

export type Channel = {
  localBalanceSat: number;
  localBalanceMsat: number;
  localSpendableBalanceSat: number;
  localSpendableBalanceMsat: number;
  remoteBalanceSat: number;
  remoteBalanceMsat: number;
  remotePubkey: string;
  id: string;
  fundingTxId: string;
  fundingTxVout: number;
  active: boolean;
  public: boolean;
  confirmations?: number;
  confirmationsRequired?: number;
  forwardingFeeBaseMsat: number;
  forwardingFeeProportionalMillionths: number;
  unspendablePunishmentReserveSat: number;
  counterpartyUnspendablePunishmentReserveSat: number;
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

export type PayInvoiceResponse = Transaction;

export type CreateOfferRequest = {
  description: string;
};

export type CreateInvoiceRequest = {
  amountSat?: number;
  amountMsat?: number;
  description: string;
};

export type PayInvoiceRequest = {
  amountSat?: number;
  amountMsat?: number;
  metadata?: Record<string, unknown>;
};

export type OpenChannelRequest = {
  pubkey: string;
  amountSats: number;
  public: boolean;
};

export type OpenChannelResponse = {
  fundingTxId: string;
};

// eslint-disable-next-line @typescript-eslint/no-empty-object-type
export type CloseChannelResponse = {};

export type PendingBalancesDetails = {
  channelId: string;
  nodeId: string;
  amountSat: number;
  fundingTxId: string;
  fundingTxVout: number;
};

export type OnchainBalanceResponse = {
  spendableSat: number;
  totalSat: number;
  reservedSat: number;
  pendingBalancesFromChannelClosuresSat: number;
  pendingBalancesDetails: PendingBalancesDetails[];
  pendingSweepBalancesDetails: PendingBalancesDetails[];
};

// from https://mempool.space/docs/api/rest#get-address-utxo
export type MempoolUtxo = {
  txid: string;
  vout: number;
  status: {
    confirmed: boolean;
    block_height?: number;
    block_hash?: string;
    block_time?: number;
  };
  value: number;
};

// from https://mempool.space/docs/api/rest#get-node-stats
export type MempoolNode = {
  alias: string;
  public_key: string;
  color: string;
  active_channel_count: number;
  sockets: string;
};

// from https://mempool.space/docs/api/rest#get-transaction
export type MempoolTransaction = {
  txid: string;
  //version: 1,
  //locktime: 0,
  // vin: [],
  //vout: [],
  size: number;
  weight: number;
  fee: number;
  status:
    | {
        confirmed: true;
        block_height: number;
        block_hash: string;
        block_time: number;
      }
    | { confirmed: false };
};

export type LongUnconfirmedZeroConfChannel = { id: string; message: string };

export type SetupNodeInfo = Partial<{
  backendType: BackendType;

  mnemonic?: string;
  nextBackupReminder?: string;

  lndAddress?: string;
  lndCertFile?: string;
  lndMacaroonFile?: string;

  phoenixdAddress?: string;
  phoenixdAuthorization?: string;
}>;

export type LSPType = "LSPS1";

export type LSPChannelOfferPaymentMethod =
  | "card"
  | "wallet"
  | "prepaid"
  | "included";

export type LSPChannelOffer = {
  lspName: string;
  lspDescription: string;
  lspContactUrl: string;
  lspBalanceSat: number;
  feeTotalSat: number;
  feeTotalUsd: number;
  currentPaymentMethod: LSPChannelOfferPaymentMethod;
  terms: string;
};

export type RecommendedChannelPeer = {
  network: Network;
  image: string;
  name: string;
  minimumChannelSizeSat: number;
  maximumChannelSizeSat: number;
  note: string;
  publicChannelsAllowed: boolean;
  description: string;
} & (
  | {
      paymentMethod: "onchain";
      pubkey: string;
      host: string;
    }
  | {
      paymentMethod: "lightning";
      type: LSPType;
      identifier: string;
      contactUrl: string;
      terms?: string;
      pubkey?: string;
      maximumChannelExpiryBlocks?: number;
      feeTotalSat1m?: number;
      feeTotalSat2m?: number;
      feeTotalSat3m?: number;
    }
);

export type AlbyInfo = {
  hub: {
    latestVersion: string;
    latestReleaseNotes: string;
  };
};

export type BitcoinRate = {
  code: string;
  symbol: string;
  rate: string;
  rate_float: number;
  rate_cents: number;
};

export type Currency = {
  iso_code: string;
  symbol: string;
  name: string;
  priority: number;
};

// TODO: use camel case (needs mapping in the Alby OAuth Service - see how AlbyInfo is done above)
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
    name?: string;
    config?: {
      region?: string;
    };
  };
  subscription: {
    plan_code: string;
  };
};

export type LSPOrderRequest = {
  amountSat?: number;
  lspType: LSPType;
  lspIdentifier: string;
  public: boolean;
};

export type RedeemOnchainFundsRequest = {
  toAddress: string;
  amountSat?: number;
  feeRate?: number;
  sendAll?: boolean;
};

export type AutoSwapRequest = {
  balanceThresholdSat?: number;
  swapAmountSat?: number;
  destination: string;
  destinationType?: string;
  unlockPassword?: string;
};

export type InitiateSwapRequest = {
  swapAmountSat?: number;
  destination?: string;
};

export type LSPOrderResponse = {
  invoice?: string;
  feeSat: number;
  invoiceAmountSat: number;
  incomingLiquiditySat: number;
  outgoingLiquiditySat: number;
};

export type AutoChannelRequest = {
  isPublic: boolean;
};
export type AutoChannelResponse = {
  invoice?: string;
  feeSat?: number;
  channelSizeSat: number;
};

export type RedeemOnchainFundsResponse = {
  txId: string;
};

export type LightningBalanceResponse = {
  totalSpendableSat: number;
  totalSpendableMsat: number;
  totalReceivableSat: number;
  totalReceivableMsat: number;
  nextMaxSpendableSat: number;
  nextMaxSpendableMsat: number;
  nextMaxReceivableSat: number;
  nextMaxReceivableMsat: number;
  nextMaxSpendableMPPSat: number;
  nextMaxSpendableMPPMsat: number;
  nextMaxReceivableMPPSat: number;
  nextMaxReceivableMPPMsat: number;
};

export type BalancesResponse = {
  onchain: OnchainBalanceResponse;
  lightning: LightningBalanceResponse;
};

export type Transaction = {
  id: number;
  type: "incoming" | "outgoing";
  state: "settled" | "pending" | "failed";
  appId: number | undefined;
  invoice: string;
  description: string;
  descriptionHash: string;
  preimage: string | undefined;
  paymentHash: string;
  amountSat: number;
  amountMsat: number;
  feesPaidSat: number;
  feesPaidMsat: number;
  updatedAt: string;
  createdAt: string;
  settledAt: string | undefined;
  metadata?: TransactionMetadata;
  boostagram?: Boostagram;
  failureReason: string;
};

export type TransactionMetadata = {
  comment?: string; // LUD-12
  payer_data?: {
    email?: string;
    name?: string;
    pubkey?: string;
  }; // LUD-18
  recipient_data?: {
    identifier?: string;
  }; // LUD-18
  nostr?: {
    pubkey: string;
    tags: string[][];
  }; // NIP-57
  offer?: {
    id: string;
    payer_note: string;
  }; // BOLT-12
  swap_id?: string;
  user_labels?: Record<string, string>;
} & Record<string, unknown>;

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
  valueSatTotal: number;
  valueMsatTotal: number;
};

export type OnchainTransaction = {
  amountSat: number;
  amountMsat: number;
  createdAt: number;
  type: "incoming" | "outgoing";
  state: "confirmed" | "unconfirmed";
  numConfirmations: number;
  txId: string;
};

export type ListAppsResponse = {
  apps: App[];
  totalCount: number;
  totalBalanceSat?: number;
  totalBalanceMsat?: number;
};

export type ListTransactionsResponse = {
  transactions: Transaction[];
  totalCount: number;
};

export type NewChannelOrderStatus = "pay" | "paid" | "success" | "opening";

type NewChannelOrderCommon = {
  amountSat: string;
  isPublic: boolean;
  status: NewChannelOrderStatus;
  fundingTxId?: string;
  prevChannelIds: string[];
};

export type OnchainOrder = {
  paymentMethod: "onchain";
  pubkey: string;
  host: string;
} & NewChannelOrderCommon;

export type LightningOrder = {
  paymentMethod: "lightning";
  lspType: LSPType;
  lspIdentifier: string;
} & NewChannelOrderCommon;

export type NewChannelOrder = OnchainOrder | LightningOrder;

export type AuthTokenResponse = {
  token: string;
};

export type GetForwardsResponse = {
  outboundAmountForwardedSat: number;
  outboundAmountForwardedMsat: number;
  totalFeeEarnedSat: number;
  totalFeeEarnedMsat: number;
  numForwards: number;
};
