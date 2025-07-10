import {
  Card,
  CardContent,
  CardFooter,
  CardHeader,
  CardTitle,
} from "src/components/ui/card";

import { nwc } from "@getalby/sdk";
import dayjs from "dayjs";
import { ChevronUpIcon, ZapIcon } from "lucide-react";
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
import { Separator } from "src/components/ui/separator";
import { Textarea } from "src/components/ui/textarea";
import { useToast } from "src/components/ui/use-toast";
import { PayInvoiceResponse } from "src/types";
import { request } from "src/utils/request";

// Must be a sub-wallet connection with only make invoice and list transactions permissions!
const LIGHTNING_MESSAGEBOARD_NWC_URL =
  import.meta.env.VITE_LIGHTNING_MESSAGEBOARD_NWC_URL ||
  "nostr+walletconnect://31758cb11d8060fa87ea955808dc22e3602aad7390717edd56dbbbd136c85a9b?relay=wss://relay.getalby.com/v1&secret=dce4d879ca8d875b0dc38f98425829eff71a5d213db9d5d423bf284fa75efc80";

type Message = {
  name?: string;
  message: string;
  amount: number;
  created_at: number;
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
              created_at: transaction.created_at,
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
        variant: "destructive",
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
      toast({ title: "Successfully sent message" });
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
              {messages?.map((message, index) => (
                <div key={index}>
                  <CardHeader>
                    <CardTitle className="leading-6 break-anywhere">
                      {message.message}
                    </CardTitle>
                  </CardHeader>
                  <CardFooter className="flex items-center justify-between text-sm">
                    <CardTitle className="break-all font-normal text-xs">
                      <span className="text-muted-foreground">by</span>{" "}
                      {message.name || "Anonymous"}{" "}
                      <span className="text-muted-foreground">
                        {dayjs(message.created_at * 1000).fromNow()}
                      </span>
                    </CardTitle>
                    <div>
                      <Badge className="py-1">
                        <ZapIcon className="size-4 mr-1" />{" "}
                        {new Intl.NumberFormat().format(message.amount)}
                      </Badge>
                    </div>
                  </CardFooter>
                  {index !== messages.length - 1 && <Separator />}
                </div>
              ))}
            </div>
            <form
              onSubmit={handleSubmitOpenDialog}
              className="flex items-center gap-2 mt-4"
            >
              <Input
                required
                placeholder="Type your message..."
                value={messageText}
                maxLength={140}
                onChange={(e) => setMessageText(e.target.value)}
              />
              <Button>
                <ZapIcon className="size-4 mr-2" /> Send
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
                Pay to post on the Alby Hub message board. The messages with the
                highest number of satoshis will be shown first.
              </DialogDescription>
            </DialogHeader>

            <div className="grid gap-4 py-4">
              <div className="grid grid-cols-4 items-center gap-4">
                <Label htmlFor="comment" className="text-right">
                  Your Name
                </Label>
                <div className="col-span-3">
                  <Input
                    id="sender-name"
                    value={senderName}
                    onChange={(e) => setSenderName(e.target.value)}
                    maxLength={20}
                    autoFocus
                  />
                </div>
              </div>

              <div className="grid grid-cols-4 items-center gap-4">
                <Label htmlFor="amount" className="text-right">
                  Amount (sats)
                </Label>
                <div className="col-span-2">
                  <Input
                    id="amount"
                    required
                    value={amount}
                    onChange={(e) => setAmount(e.target.value)}
                  />
                </div>
                <Button
                  type="button"
                  variant="secondary"
                  onClick={() => setAmount("" + topPlace)}
                >
                  <ChevronUpIcon className="size-4 mr-2" />
                  Top
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
                  rows={4}
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
