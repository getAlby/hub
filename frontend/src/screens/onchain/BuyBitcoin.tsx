import React from "react";
import AppHeader from "src/components/AppHeader";
import { MempoolAlert } from "src/components/MempoolAlert";
import { LoadingButton } from "src/components/ui/custom/loading-button";
import { Input } from "src/components/ui/input";
import { Label } from "src/components/ui/label";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "src/components/ui/select";
import { useOnchainAddress } from "src/hooks/useOnchainAddress";
import { openLink } from "src/utils/openLink";

const SUPPORTED_CURRENCIES = [
  {
    value: "usd",
    label: "USD - US Dollar",
  },
  {
    value: "ars",
    label: "ARS - Argentine Peso",
  },
  {
    value: "aud",
    label: "AUD - Australian Dollar",
  },
  {
    value: "bgn",
    label: "BGN - Bulgarian Lev",
  },
  {
    value: "brl",
    label: "BRL - Brazilian Real",
  },
  {
    value: "cad",
    label: "CAD - Canadian Dollar",
  },
  {
    value: "chf",
    label: "CHF - Swiss Franc",
  },
  {
    value: "cop",
    label: "COP - Colombian Peso",
  },
  {
    value: "czk",
    label: "CZK - Czech Koruna",
  },
  {
    value: "dkk",
    label: "DKK - Danish Krone",
  },
  {
    value: "dop",
    label: "DOP - Dominican Peso",
  },
  {
    value: "egp",
    label: "EGP - Egyptian Pound",
  },
  {
    value: "eur",
    label: "EUR - Euro",
  },
  {
    value: "gbp",
    label: "GBP - Pound Sterling",
  },
  {
    value: "hkd",
    label: "HKD - Hong Kong Dollar",
  },
  {
    value: "idr",
    label: "IDR - Indonesian Rupiah",
  },
  {
    value: "ils",
    label: "ILS - Israeli New Shekel",
  },
  {
    value: "jod",
    label: "JOD - Jordanian Dollar",
  },
  {
    value: "kes",
    label: "KES - Kenyan Shilling",
  },
  {
    value: "kwd",
    label: "KWD - Kuwaiti Dinar",
  },
  {
    value: "lkr",
    label: "LKR - Sri Lankan Rupee",
  },
  {
    value: "mxn",
    label: "MXN - Mexican Peso",
  },
  {
    value: "ngn",
    label: "NGN - Nigerian Naira",
  },
  {
    value: "nok",
    label: "NOK - Norwegian Krone",
  },
  {
    value: "nzd",
    label: "NZD - New Zealand Dollar",
  },
  {
    value: "omr",
    label: "OMR - Omani Rial",
  },
  {
    value: "pen",
    label: "PEN - Peruvian Sol",
  },
  {
    value: "pln",
    label: "PLN - Polish Złoty",
  },
  {
    value: "ron",
    label: "RON - Romanian Leu",
  },
  {
    value: "sek",
    label: "SEK - Swedish Krona",
  },
  {
    value: "thb",
    label: "THB - Thai Baht",
  },
  {
    value: "try",
    label: "TRY - Turkish Lira",
  },
  {
    value: "twd",
    label: "TWD - Taiwan Dollar",
  },
  {
    value: "vnd",
    label: "VND - Vietnamese Dong",
  },
  {
    value: "zar",
    label: "ZAR - South African Rand",
  },
];

export default function BuyBitcoin() {
  const [currency, setCurrency] = React.useState("usd");
  const [amount, setAmount] = React.useState("250");
  const { data: onchainAddress } = useOnchainAddress();

  async function launch() {
    const url = `https://getalby.com/topup?address=${onchainAddress}&amount=${amount}&currency=${currency}`;
    openLink(url);
  }

  return (
    <div className="grid gap-5">
      <AppHeader
        title="Buy Bitcoin"
        description="Use one of our partner providers to buy bitcoin and deposit it to your on-chain balance."
      />
      <MempoolAlert />
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-1 lg:gap-10">
        <div className="flex max-w-lg flex-col gap-4">
          <div className="grid gap-4">
            <p className="text-muted-foreground">
              How much bitcoin would you like to buy?
            </p>
            <div className="grid gap-1.5">
              <Label htmlFor="amount">Enter Amount</Label>
              <Input
                name="amount"
                autoFocus
                onChange={(e) => setAmount(e.target.value)}
                value={amount}
                type="text"
                placeholder="amount"
              />
            </div>

            <div className="grid gap-2">
              <Label>Currency</Label>
              <Select
                name="currency"
                value={currency}
                onValueChange={(value) => setCurrency(value)}
              >
                <SelectTrigger>
                  <SelectValue placeholder="Currency" />
                </SelectTrigger>
                <SelectContent>
                  {SUPPORTED_CURRENCIES.map((currency) => (
                    <SelectItem key={currency.value} value={currency.value}>
                      {currency.label}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>

            <LoadingButton size="sm" onClick={launch} type="button">
              Next
            </LoadingButton>
          </div>
        </div>
      </div>
    </div>
  );
}
