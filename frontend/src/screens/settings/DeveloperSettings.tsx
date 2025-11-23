import { CopyIcon, SquareArrowOutUpRightIcon } from "lucide-react";
import React from "react";
import { toast } from "sonner";
import PasswordInput from "src/components/password/PasswordInput";
import SettingsHeader from "src/components/SettingsHeader";
import { Button } from "src/components/ui/button";
import { ExternalLinkButton } from "src/components/ui/custom/external-link-button";
import { LoadingButton } from "src/components/ui/custom/loading-button";
import { Input } from "src/components/ui/input";
import { Label } from "src/components/ui/label";
import { RadioGroup, RadioGroupItem } from "src/components/ui/radio-group";
import { Separator } from "src/components/ui/separator";
import { useAlbyMe } from "src/hooks/useAlbyMe";
import { copyToClipboard } from "src/lib/clipboard";
import { AuthTokenResponse } from "src/types";
import { request } from "src/utils/request";

export default function DeveloperSettings() {
  const { data: albyMe } = useAlbyMe();
  const [token, setToken] = React.useState<string>();
  const [tokenPermission, setTokenPermission] = React.useState<string>();
  const [expiryDays, setExpiryDays] = React.useState<string>("365");
  const [unlockPassword, setUnlockPassword] = React.useState<string>("");
  const [permission, setPermission] = React.useState<"full" | "readonly">(
    "full"
  );
  const [showCreateTokenForm, setShowCreateTokenForm] =
    React.useState<boolean>();
  const [loading, setLoading] = React.useState<boolean>();

  async function createToken(e: React.FormEvent) {
    e.preventDefault();
    try {
      setLoading(true);
      if (!expiryDays || !unlockPassword || !permission) {
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
            permission,
          }),
        }
      );
      if (authTokenResponse) {
        setToken(authTokenResponse.token);
        setTokenPermission(permission);
      }
    } catch (error) {
      console.error(error);
      toast.error("Something went wrong", {
        description: "" + error,
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
            Learn More on nwc.dev
            <SquareArrowOutUpRightIcon />
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
              <div className="grid gap-3">
                <Label>Token Type</Label>
                <RadioGroup
                  value={permission}
                  onValueChange={(v) => {
                    if (v != "readonly" && v !== "full") {
                      throw new Error("Unknown permission type");
                    }
                    setPermission(v);
                  }}
                  className="mt-4 gap-4"
                >
                  <div className="flex items-start space-x-2">
                    <RadioGroupItem value="full" id="full" />
                    <Label
                      htmlFor="full"
                      className="flex-1 flex flex-col justify-center items-start cursor-pointer"
                    >
                      <div className="font-medium shrink-0">Full Access</div>
                      <div className="text-sm text-muted-foreground">
                        Complete control over your hub - can read data and
                        perform all operations (send payments, manage apps,
                        etc.)
                      </div>
                    </Label>
                  </div>
                  <div className="flex items-start space-x-2">
                    <RadioGroupItem value="readonly" id="readonly" />
                    <Label
                      htmlFor="readonly"
                      className="flex-1  flex flex-col justify-center items-start cursor-pointer"
                    >
                      <div className="font-medium">Read-Only Access</div>
                      <div className="text-sm text-muted-foreground">
                        View-only access - can read balances, transactions, and
                        other data but cannot perform operations
                      </div>
                    </Label>
                  </div>
                </RadioGroup>
              </div>
              <div className="grid gap-2">
                <Label htmlFor="token-expiry">Token Expiry (Days)</Label>
                <Input
                  type="number"
                  name="token-expiry"
                  id="token-expiry"
                  onChange={(e) => setExpiryDays(e.target.value)}
                  value={expiryDays}
                />
              </div>
              <div className="grid gap-2">
                <Label htmlFor="password">Unlock Password</Label>
                <PasswordInput
                  id="password"
                  onChange={setUnlockPassword}
                  value={unlockPassword}
                  autoFocus
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
                  copyToClipboard(token);
                }}
              >
                <CopyIcon className="size-4" />
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
              {tokenPermission === "readonly" ? (
                <>
                  This is a read-only token that can view data but cannot
                  perform operations like sending payments. Please keep it
                  secure.
                </>
              ) : (
                <>
                  This token grants full access to your hub. Please keep it
                  secure.
                </>
              )}{" "}
              If you suspect that the token has been compromised, immediately
              change your unlock password.
            </p>
          </>
        )}
      </div>
    </>
  );
}
