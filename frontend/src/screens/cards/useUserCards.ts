import React from "react";
import {
  DEFAULT_APP_BUDGET_RENEWAL,
  DEFAULT_APP_BUDGET_SATS,
} from "src/constants";
import { createApp } from "src/requests/createApp";

export type Provider = {
  id: string;
  name: string;
  logo: string;
  initials: string;
  network: "Visa" | "Mastercard";
};

// Mirrors bitcoin-card-topup's CardConfig (label/destinationAddress/chainId/currency).
// The NWC connection secret is *not* stored — it only lives in the one-shot
// CardCreatedDialog URL right after connection.
export type SupportedCurrency = "USDC" | "USDT";

export type UserCard = {
  id: string;
  providerId: string;
  destinationAddress: string;
  chainId: number;
  currency: SupportedCurrency;
  createdAt: number;
  /** id of the NWC app this card is bound to. Optional only for legacy/sample. */
  appId?: number;
};

const STORAGE_KEY = "cardsUserCards";

const sampleCards: UserCard[] = [
  {
    id: "card-1",
    providerId: "redotpay",
    destinationAddress: "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb0",
    chainId: 42161,
    currency: "USDC",
    createdAt: Date.now() - 1000 * 60 * 60 * 24 * 3,
  },
];

function loadCards(): UserCard[] {
  try {
    const raw = localStorage.getItem(STORAGE_KEY);
    if (raw === null) {
      localStorage.setItem(STORAGE_KEY, JSON.stringify(sampleCards));
      return sampleCards;
    }
    const parsed = JSON.parse(raw);
    if (Array.isArray(parsed)) {
      return parsed;
    }
  } catch {
    // ignore parse errors, fall through to default
  }
  return sampleCards;
}

function saveCards(cards: UserCard[]) {
  localStorage.setItem(STORAGE_KEY, JSON.stringify(cards));
}

type AddCardResult = { card: UserCard; pairingUri: string };

export function useUserCards() {
  const [cards, setCards] = React.useState<UserCard[]>(() => loadCards());

  const addCard = React.useCallback(
    async (
      input: Omit<UserCard, "id" | "createdAt" | "appId"> & {
        providerName: string;
      }
    ): Promise<AddCardResult> => {
      // Mint a real NWC connection — this is the one moment we hold the
      // pairing URI; after this it's gone (write-only secret).
      const response = await createApp({
        name: `${input.providerName} Card`,
        scopes: [
          "get_info",
          "get_balance",
          "list_transactions",
          "lookup_invoice",
          "make_invoice",
          "pay_invoice",
          "notifications",
        ],
        maxAmountSat: DEFAULT_APP_BUDGET_SATS,
        budgetRenewal: DEFAULT_APP_BUDGET_RENEWAL,
        metadata: {
          app_store_app_id: "bitcoin-card-topup",
        },
      });

      const card: UserCard = {
        id: `card-${response.id}`,
        providerId: input.providerId,
        destinationAddress: input.destinationAddress,
        chainId: input.chainId,
        currency: input.currency,
        createdAt: Date.now(),
        appId: response.id,
      };

      setCards((prev) => {
        const next = [...prev, card];
        saveCards(next);
        return next;
      });

      return { card, pairingUri: response.pairingUri };
    },
    []
  );

  const removeCard = React.useCallback((id: string) => {
    setCards((prev) => {
      const next = prev.filter((c) => c.id !== id);
      saveCards(next);
      return next;
    });
  }, []);

  return { cards, addCard, removeCard };
}
