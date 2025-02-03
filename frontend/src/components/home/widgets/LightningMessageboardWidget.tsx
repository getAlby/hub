import {
  Card,
  CardContent,
  CardFooter,
  CardHeader,
  CardTitle,
} from "src/components/ui/card";

import { nwc } from "@getalby/sdk";
import { Zap } from "lucide-react";
import React from "react";
import Loading from "src/components/Loading";
import { Badge } from "src/components/ui/badge";
import { Button } from "src/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "src/components/ui/dialog";
import { Input } from "src/components/ui/input";
import { Label } from "src/components/ui/label";
import { LoadingButton } from "src/components/ui/loading-button";
import { Textarea } from "src/components/ui/textarea";
import { useToast } from "src/components/ui/use-toast";
import { PayInvoiceResponse } from "src/types";
import { request } from "src/utils/request";

// Must be a sub-wallet connection with only make invoice and list transactions permissions!
const LIGHTNING_MESSAGEBOARD_NWC_URL =
  import.meta.env.VITE_LIGHTNING_MESSAGEBOARD_NWC_URL ||
  "nostr+walletconnect://f70c731046253fe6d53143f0e62527e08b5011fee5ab9e4c3c5f3075c21a6cb8?relay=wss://relay.getalby.com/v1&secret=e27cac72651d3733f2f195722c9c7d574a34883acf97dc89b8941a838fef43a7";

type Message = {
  name?: string;
  message: string;
  amount: number;
};

let nwcClient: nwc.NWCClient | undefined;
function getNWCClient(): nwc.NWCClient {
  if (!nwcClient) {
    nwcClient = new nwc.NWCClient({
      nostrWalletConnectUrl: LIGHTNING_MESSAGEBOARD_NWC_URL,
    });
  }
  return nwcClient;
}

export function LightningMessageboardWidget() {
  const [messageText, setMessageText] = React.useState("");
  const [senderName, setSenderName] = React.useState("");
  const [amount, setAmount] = React.useState("");
  const [messages, setMessages] = React.useState<Message[]>();
  const [isLoading, setLoading] = React.useState(false);
  const [isSubmitting, setSubmitting] = React.useState(false);
  const [dialogOpen, setDialogOpen] = React.useState(false);
  const { toast } = useToast();
  const [isOpen, setOpen] = React.useState(false);

  const loadMessages = React.useCallback(() => {
    (async () => {
      setLoading(true);
      let offset = 0;
      const _messages: Message[] = [];
      while (true) {
        try {
          const transactions = await getNWCClient().listTransactions({
            offset,
            limit: 10,
          });

          if (transactions.transactions.length === 0) {
            break;
          }

          _messages.push(
            ...transactions.transactions.map((transaction) => ({
              message: transaction.description,
              name: (
                transaction.metadata as
                  | { payer_data?: { name?: string } }
                  | undefined
              )?.payer_data?.name as string | undefined,
              amount: Math.floor(transaction.amount / 1000),
            }))
          );

          offset += transactions.transactions.length;
        } catch (error) {
          console.error(error);
          await new Promise((resolve) => setTimeout(resolve, 1000));
        }
      }
      _messages.sort((a, b) => b.amount - a.amount);
      setMessages(_messages);
      setLoading(false);
    })();
  }, []);

  const hasLoadedMessages = !!messages;

  React.useEffect(() => {
    if (isOpen && !hasLoadedMessages) {
      loadMessages();
    }
  }, [hasLoadedMessages, isOpen, loadMessages]);

  function handleSubmitOpenDialog(e: React.FormEvent) {
    e.preventDefault();
    setDialogOpen(true);
  }
  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();

    if (+amount < 1000) {
      toast({
        title: "Amount too low",
        description: "Minimum payment is 1000 sats",
      });
      return;
    }

    const amountMsat = +amount * 1000;
    setSubmitting(true);
    try {
      const transaction = await getNWCClient().makeInvoice({
        amount: amountMsat,
        description: messageText,
        metadata: {
          payer_data: {
            name: senderName,
          },
        },
      });

      const payInvoiceResponse = await request<PayInvoiceResponse>(
        `/api/payments/${transaction.invoice}`,
        {
          method: "POST",
        }
      );
      if (!payInvoiceResponse?.preimage) {
        throw new Error("No preimage in response");
      }

      setMessageText("");
      loadMessages();
      toast({ title: "Sucessfully sent message" });
      setDialogOpen(false);
    } catch (error) {
      console.error(error);
      toast({
        variant: "destructive",
        description: "Something went wrong: " + error,
      });
    }
    setSubmitting(false);
  }

  const topPlace = Math.max(
    1000,
    ...(messages?.map((message) => message.amount + 1) || [])
  );

  return (
    <>
      <Card>
        <CardHeader>
          <div className="flex justify-between items-center">
            <CardTitle className="flex items-center gap-2">
              Lightning Messageboard{isLoading && <Loading />}
            </CardTitle>
            <Button variant="secondary" onClick={() => setOpen(!isOpen)}>
              {isOpen ? "Hide" : "Show"}
            </Button>
          </div>
        </CardHeader>
        {isOpen && (
          <CardContent>
            <div className="h-96 overflow-y-visible flex flex-col gap-2 overflow-hidden">
              {messages?.map((message) => (
                <Card className="mr-2">
                  <CardHeader>
                    <CardTitle className="flex gap-2 items-start justify-start">
                      <p className="break-words">{message.message}</p>
                    </CardTitle>
                  </CardHeader>
                  <CardFooter className="flex items-center justify-between text-sm">
                    <Badge className="py-1">
                      <Zap className="w-4 h-4 mr-2" />{" "}
                      {new Intl.NumberFormat().format(message.amount)}
                    </Badge>
                    <CardTitle className="font-normal text-xs">
                      {message.name || "Anonymous"}
                    </CardTitle>
                  </CardFooter>
                </Card>
              ))}
            </div>
            <form
              onSubmit={handleSubmitOpenDialog}
              className="flex items-center gap-2 mt-4"
            >
              <Input
                required
                placeholder="type your message..."
                value={messageText}
                maxLength={140}
                onChange={(e) => setMessageText(e.target.value)}
              />
              <Button>
                <Zap className="w-4 h-4 mr-2" /> Send
              </Button>
            </form>
          </CardContent>
        )}
      </Card>
      <Dialog onOpenChange={setDialogOpen} open={dialogOpen}>
        <DialogContent className="sm:max-w-[600px]">
          <form onSubmit={handleSubmit}>
            <DialogHeader>
              <DialogTitle>Post Message</DialogTitle>
              <DialogDescription>
                Pay sats to post on the Alby Hub message board. The messages
                with the highest number of satoshis will be shown first.
              </DialogDescription>
            </DialogHeader>

            <div className="grid gap-4 py-4">
              <div className="grid grid-cols-4 items-center gap-4">
                <Label htmlFor="comment" className="text-right">
                  Your Name
                </Label>
                <Input
                  id="sender-name"
                  value={senderName}
                  onChange={(e) => setSenderName(e.target.value)}
                  className="col-span-3"
                  autoFocus
                />
              </div>

              <div className="grid grid-cols-4 items-center gap-4">
                <Label htmlFor="amount" className="text-right">
                  Amount (sats)
                </Label>
                <Input
                  id="amount"
                  required
                  value={amount}
                  onChange={(e) => setAmount(e.target.value)}
                  className="col-span-2"
                />
                <Button
                  type="button"
                  variant="secondary"
                  onClick={() => setAmount("" + topPlace)}
                >
                  Top ⚡{new Intl.NumberFormat().format(topPlace)}
                </Button>
              </div>
              <div className="grid grid-cols-4 gap-4">
                <Label htmlFor="comment" className="text-right pt-2">
                  Message
                </Label>
                <Textarea
                  id="comment"
                  value={messageText}
                  onChange={(e) => setMessageText(e.target.value)}
                  className="col-span-3"
                />
              </div>
            </div>
            <DialogFooter>
              <LoadingButton
                type="submit"
                disabled={!!isSubmitting}
                loading={isSubmitting}
              >
                Confirm Payment
              </LoadingButton>
            </DialogFooter>
          </form>
        </DialogContent>
      </Dialog>
    </>
  );
}
