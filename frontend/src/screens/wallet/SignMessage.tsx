import { CopyIcon } from "lucide-react";
import React from "react";
import { toast } from "sonner";
import AppHeader from "src/components/AppHeader";
import { Button } from "src/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "src/components/ui/card";
import { LoadingButton } from "src/components/ui/custom/loading-button";
import { Input } from "src/components/ui/input";
import { Label } from "src/components/ui/label";

import { copyToClipboard } from "src/lib/clipboard";
import { SignMessageResponse } from "src/types";
import { request } from "src/utils/request";

export default function SignMessage() {
  const [isLoading, setLoading] = React.useState(false);
  const [message, setMessage] = React.useState("");
  const [signature, setSignature] = React.useState("");
  const [signatureMessage, setSignatureMessage] = React.useState("");

  const handleSubmit = async (event: React.FormEvent<HTMLFormElement>) => {
    event.preventDefault();

    try {
      setLoading(true);
      const signMessageResponse = await request<SignMessageResponse>(
        "/api/wallet/sign-message",
        {
          method: "POST",
          headers: {
            "Content-Type": "application/json",
          },
          body: JSON.stringify({ message: message.trim() }),
        }
      );
      setSignatureMessage(message);
      setMessage("");
      if (signMessageResponse) {
        setSignature(signMessageResponse.signature);
        toast("Successfully signed message");
      }
    } catch (e) {
      toast.error("Failed to sign message", {
        description: "" + e,
      });
      console.error(e);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="grid gap-5">
      <AppHeader
        title="Sign Message"
        description="Manually sign a message with your node's key (e.g. to proof ownership of your node)"
      />
      <div className="max-w-lg">
        <form onSubmit={handleSubmit} className="grid gap-5">
          <div className="grid gap-2">
            <Label htmlFor="message">Message</Label>
            <Input
              id="message"
              type="text"
              value={message}
              placeholder=""
              onChange={(e) => {
                setMessage(e.target.value);
                setSignature("");
              }}
            />
          </div>
          <div>
            <LoadingButton
              loading={isLoading}
              type="submit"
              disabled={!message}
            >
              Sign
            </LoadingButton>
          </div>
          {signature && (
            <Card>
              <CardHeader>
                <CardTitle>Signed Message</CardTitle>
                <CardDescription>{signatureMessage}</CardDescription>
              </CardHeader>
              <CardContent>
                <div className="flex flex-row items-center gap-2">
                  <Input
                    type="text"
                    value={signature}
                    className="flex-1"
                    readOnly
                  />
                  <Button
                    type="button"
                    variant="secondary"
                    size="icon"
                    onClick={() => {
                      copyToClipboard(signature);
                    }}
                  >
                    <CopyIcon className="size-4" />
                  </Button>
                </div>
              </CardContent>
            </Card>
          )}
        </form>
      </div>
    </div>
  );
}
