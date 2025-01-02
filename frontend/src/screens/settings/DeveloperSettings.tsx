import { Copy, Rocket, TriangleAlert } from "lucide-react";
import React from "react";
import SettingsHeader from "src/components/SettingsHeader";
import { Alert, AlertDescription, AlertTitle } from "src/components/ui/alert";
import { Button, ExternalLinkButton } from "src/components/ui/button";
import { Input } from "src/components/ui/input";
import { Label } from "src/components/ui/label";
import { LoadingButton } from "src/components/ui/loading-button";
import { useToast } from "src/components/ui/use-toast";
import { useAlbyMe } from "src/hooks/useAlbyMe";
import { copyToClipboard } from "src/lib/clipboard";
import { AuthTokenResponse } from "src/types";
import { request } from "src/utils/request";

export default function DeveloperSettings() {
  const { data: albyMe } = useAlbyMe();
  const [token, setToken] = React.useState<string>();
  const [expiryDays, setExpiryDays] = React.useState<string>("365");
  const [unlockPassword, setUnlockPassword] = React.useState<string>();
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
        description="Power your apps with Alby Hub"
      />
      <Alert>
        <Rocket className="h-4 w-4" />
        <AlertTitle>Power your Apps with NWC</AlertTitle>
        <AlertDescription>
          <div className="mb-2">
            Alby Hub can power your lightning apps and services in all
            environments using the amazing open protocol{" "}
            <span className="font-semibold">Nostr Wallet Connect</span>.
          </div>
          <ExternalLinkButton size={"sm"} to="https://nwc.dev">
            Learn More
          </ExternalLinkButton>
        </AlertDescription>
      </Alert>
      <Alert>
        <TriangleAlert className="h-4 w-4" />
        <AlertTitle>Experimental</AlertTitle>
        <AlertDescription>
          <div className="mb-2">
            You can use your auth token to access the Alby Hub internal API.
            However, whenever possible, we recommend using the NWC API directly
            for more stability. Please note that the internal API may change or
            be removed entirely in the future.
          </div>
          {!token && !showCreateTokenForm && (
            <Button size={"sm"} onClick={() => setShowCreateTokenForm(true)}>
              Configure Token
            </Button>
          )}
          {showCreateTokenForm && !token && (
            <form onSubmit={createToken} className="w-full flex flex-col gap-3">
              <>
                <div className="grid gap-1.5">
                  <Label htmlFor="password">Token Expiry (Days)</Label>
                  <Input
                    type="number"
                    name="token-expiry"
                    onChange={(e) => setExpiryDays(e.target.value)}
                    value={expiryDays}
                  />
                </div>
                <div className="grid gap-1.5">
                  <Label htmlFor="password">Unlock Password</Label>
                  <Input
                    type="password"
                    name="password"
                    onChange={(e) => setUnlockPassword(e.target.value)}
                    value={unlockPassword}
                    placeholder="Password"
                  />
                </div>
                <LoadingButton loading={loading}>Create Token</LoadingButton>
              </>
            </form>
          )}
          {token && (
            <>
              <div className="flex flex-row items-center gap-2 mb-2">
                <Input
                  type="password"
                  value={token}
                  className="flex-1"
                  readOnly
                />
                <Button
                  type="button"
                  variant="secondary"
                  size="icon"
                  onClick={() => {
                    copyToClipboard(token, toast);
                  }}
                >
                  <Copy className="w-4 h-4" />
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
                </ol>
              </div>
              <p className="text-xs">
                This token grants full access to your hub. Please keep it
                secure. If you suspect that the token has been compromised,
                immediately change your JWT_SECRET environment variable or
                contact support@getalby.com.
              </p>
            </>
          )}
        </AlertDescription>
      </Alert>
    </>
  );
}
