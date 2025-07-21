import { CopyIcon, SquareArrowOutUpRightIcon } from "lucide-react";
import React from "react";
import PasswordInput from "src/components/password/PasswordInput";
import SettingsHeader from "src/components/SettingsHeader";
import { Button, ExternalLinkButton } from "src/components/ui/button";
import { Input } from "src/components/ui/input";
import { Label } from "src/components/ui/label";
import { LoadingButton } from "src/components/ui/loading-button";
import { Separator } from "src/components/ui/separator";
import { useToast } from "src/components/ui/use-toast";
import { useAlbyMe } from "src/hooks/useAlbyMe";
import { copyToClipboard } from "src/lib/clipboard";
import { AuthTokenResponse } from "src/types";
import { request } from "src/utils/request";

export default function DeveloperSettings() {
  const { data: albyMe } = useAlbyMe();
  const [token, setToken] = React.useState<string>();
  const [expiryDays, setExpiryDays] = React.useState<string>("365");
  const [unlockPassword, setUnlockPassword] = React.useState<string>("");
  const [showCreateTokenForm, setShowCreateTokenForm] =
    React.useState<boolean>();
  const [loading, setLoading] = React.useState<boolean>();
  const { toast } = useToast();

  async function createToken(e: React.FormEvent) {
    e.preventDefault();
    try {
      setLoading(true);
      if (!expiryDays || !unlockPassword) {
        throw new Error("Form not filled");
      }
      const authTokenResponse = await request<AuthTokenResponse>(
        "/api/unlock",
        {
          method: "POST",
          headers: {
            "Content-Type": "application/json",
          },
          body: JSON.stringify({
            unlockPassword,
            tokenExpiryDays: +expiryDays,
          }),
        }
      );
      if (authTokenResponse) {
        setToken(authTokenResponse.token);
      }
    } catch (error) {
      console.error(error);
      toast({
        variant: "destructive",
        description: "Something went wrong: " + error,
      });
    }
    setLoading(false);
  }

  return (
    <>
      <SettingsHeader
        title="Developer"
        description="Power your apps with Alby Hub."
      />
      <div className="flex flex-col gap-4">
        <div className="flex flex-col gap-1 text-sm">
          <h3 className="font-semibold">Power your apps with NWC</h3>

          <div className="text-muted-foreground">
            Alby Hub can power your lightning apps and services in all
            environments using Nostr Wallet Connect, an open protocol to connect
            lightning wallets and apps
          </div>
        </div>
        <div>
          <ExternalLinkButton
            size={"lg"}
            variant={"secondary"}
            to="https://nwc.dev"
            className="flex-1 gap-2 items-center justify-center"
          >
            Learn More on nwc.dev{" "}
            <SquareArrowOutUpRightIcon className="w-4 h-4 mr-2" />
          </ExternalLinkButton>
        </div>
      </div>
      <Separator />
      <div className="flex flex-col gap-4">
        <div className="flex flex-col gap-1 text-sm">
          <h3 className="font-semibold">Experimental API access</h3>

          <div className="text-muted-foreground">
            You can use your auth token to access the Alby Hub internal API.
            However, whenever possible, we recommend using the NWC API directly
            for more stability. Please note that the internal API may change or
            be removed entirely in the future.
          </div>
        </div>
        {!token && !showCreateTokenForm && (
          <div>
            <Button
              size={"lg"}
              variant={"secondary"}
              onClick={() => setShowCreateTokenForm(true)}
              className="flex-1"
            >
              Configure Token
            </Button>
          </div>
        )}
        {showCreateTokenForm && !token && (
          <form
            onSubmit={createToken}
            className="w-full md:w-96 flex flex-col gap-4"
          >
            <>
              <div className="grid gap-2">
                <Label htmlFor="password">Token Expiry (Days)</Label>
                <Input
                  type="number"
                  name="token-expiry"
                  onChange={(e) => setExpiryDays(e.target.value)}
                  value={expiryDays}
                  autoFocus
                />
              </div>
              <div className="grid gap-2">
                <Label htmlFor="password">Unlock Password</Label>
                <PasswordInput
                  id="password"
                  onChange={setUnlockPassword}
                  value={unlockPassword}
                />
              </div>
              <div className="mt-4">
                <LoadingButton loading={loading}>Create Token</LoadingButton>
              </div>
            </>
          </form>
        )}
        {token && (
          <>
            <div className="flex flex-row items-center gap-2 mb-2">
              <PasswordInput readOnly value={token} className="flex-1" />
              <Button
                type="button"
                variant="secondary"
                size="icon"
                onClick={() => {
                  copyToClipboard(token, toast);
                }}
              >
                <CopyIcon className="w-4 h-4" />
              </Button>
            </div>
            <div className="my-4 border rounded-lg p-4">
              <p className="mb-2">
                To make requests from your application to{" "}
                <span className="font-mono font-semibold">
                  {window.location.origin}/api
                </span>{" "}
                add the following header{albyMe?.hub.name && "s"}:
              </p>
              <ol>
                <li>
                  <span className="font-mono font-semibold">
                    Authorization: Bearer YOUR_TOKEN
                  </span>{" "}
                </li>
                {albyMe?.hub.name && (
                  <li>
                    <span className="font-mono font-semibold">
                      AlbyHub-Name: {albyMe.hub.name}
                    </span>{" "}
                  </li>
                )}
                {albyMe?.hub.name && (
                  <li>
                    <span className="font-mono font-semibold">
                      AlbyHub-Region: {albyMe.hub.config?.region || "lax"}
                    </span>
                  </li>
                )}
              </ol>
            </div>

            <p className="text-xs">
              This token grants full access to your hub. Please keep it secure.
              If you suspect that the token has been compromised, immediately
              change your JWT_SECRET environment variable or contact
              support@getalby.com.
            </p>
          </>
        )}
      </div>
    </>
  );
}
