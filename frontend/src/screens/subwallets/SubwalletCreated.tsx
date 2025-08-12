import {
  AlertCircle,
  AlertTriangleIcon,
  CopyIcon,
  ExternalLinkIcon,
  TriangleAlert,
} from "lucide-react";
import React from "react";
import QRCode from "react-qr-code";
import { Link, useLocation, useNavigate } from "react-router-dom";
import AppHeader from "src/components/AppHeader";
import ExternalLink from "src/components/ExternalLink";
import { AppleIcon } from "src/components/icons/Apple";
import { PlayStoreIcon } from "src/components/icons/PlayStore";
import { ZapStoreIcon } from "src/components/icons/ZapStore";
import { IsolatedAppTopupDialog } from "src/components/IsolatedAppTopupDialog";
import {
  Accordion,
  AccordionContent,
  AccordionItem,
  AccordionTrigger,
} from "src/components/ui/accordion";
import { Alert, AlertDescription, AlertTitle } from "src/components/ui/alert";
import { Button, ExternalLinkButton } from "src/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from "src/components/ui/card";
import { Input } from "src/components/ui/input";
import { Label } from "src/components/ui/label";
import { LoadingButton } from "src/components/ui/loading-button";
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from "src/components/ui/popover";
import { Textarea } from "src/components/ui/textarea";
import { useToast } from "src/components/ui/use-toast";
import { UpgradeDialog } from "src/components/UpgradeDialog";
import { useAlbyMe } from "src/hooks/useAlbyMe";
import { useApp } from "src/hooks/useApp";
import { useCreateLightningAddress } from "src/hooks/useCreateLightningAddress";
import { useNodeConnectionInfo } from "src/hooks/useNodeConnectionInfo";
import { copyToClipboard } from "src/lib/clipboard";
import { ConnectAppCard } from "src/screens/apps/AppCreated";
import { CreateAppResponse } from "src/types";

export function SubwalletCreated() {
  const { data: nodeConnectionInfo } = useNodeConnectionInfo();
  const { toast } = useToast();

  const { state } = useLocation();
  const navigate = useNavigate();
  const createAppResponse = state as CreateAppResponse | undefined;
  const { data: app } = useApp(createAppResponse?.id, true);
  const [intendedLightningAddress, setIntendedLightningAddress] =
    React.useState(
      createAppResponse?.name.toLowerCase().replace(/[^a-z0-9]/g, "") || ""
    );

  const { data: albyMe } = useAlbyMe();

  const { createLightningAddress, creatingLightningAddress } =
    useCreateLightningAddress(createAppResponse?.id);

  const [step, setStep] = React.useState(1);

  if (!createAppResponse?.pairingUri) {
    navigate("/");
    return null;
  }

  const name = createAppResponse.name;
  let connectionSecret = createAppResponse.pairingUri;
  if (app?.metadata?.lud16) {
    connectionSecret += `&lud16=${app.metadata.lud16}`;
  }

  const albyAccountUrl = `https://getalby.com/nwc/new#${connectionSecret}`;
  const valueTag = `<podcast:value type="lightning" method="keysend">
    <podcast:valueRecipient name="${name}" type="node" address="${nodeConnectionInfo?.pubkey}" customKey="696969"  customValue="${app?.id}" split="100"/>
  </podcast:value>`;

  return (
    <div className="grid gap-5">
      <AppHeader title={`Connect ${name}`} description="" />
      <div className="max-w-lg">
        <div className="flex flex-col col-span-3 gap-5 items-start">
          {step === 1 && app && (
            <div className="grid gap-5">
              <div>
                Configure this sub-wallet by creating a lightning address and
                topping it up. That way it's ready to use as soon as it's
                connected.
              </div>
              <div className="grid gap-5">
                {!app.metadata?.lud16 && (
                  <Card>
                    <CardHeader>
                      <CardTitle>Lightning address</CardTitle>
                      <CardDescription>
                        Create a lightning address for this sub-wallet
                      </CardDescription>
                    </CardHeader>
                    <CardContent>
                      <Input
                        type="text"
                        value={intendedLightningAddress}
                        onChange={(e) =>
                          setIntendedLightningAddress(e.target.value)
                        }
                        required
                        autoComplete="off"
                        endAdornment={
                          <span className="mr-1 text-muted-foreground text-xs">
                            @getalby.com
                          </span>
                        }
                      />
                    </CardContent>
                    <CardFooter className="flex flex-row justify-end">
                      {!albyMe?.subscription.plan_code ? (
                        <UpgradeDialog>
                          <Button size="sm" variant="secondary">
                            Create Lightning Address
                          </Button>
                        </UpgradeDialog>
                      ) : (
                        <LoadingButton
                          loading={creatingLightningAddress}
                          onClick={() =>
                            createLightningAddress(intendedLightningAddress)
                          }
                          size="sm"
                          variant="secondary"
                        >
                          Create Lightning Address
                        </LoadingButton>
                      )}
                    </CardFooter>
                  </Card>
                )}
                {app.metadata?.lud16 && (
                  <Card>
                    <CardHeader>
                      <CardTitle>Lightning address</CardTitle>
                      <CardDescription>
                        Your lightning address for this sub-account
                      </CardDescription>
                    </CardHeader>
                    <CardContent>
                      <p className="font-semibold">{app.metadata.lud16}</p>
                    </CardContent>
                    <CardFooter className="flex flex-row justify-end">
                      <Button
                        onClick={() => {
                          if (app.metadata?.lud16) {
                            copyToClipboard(app.metadata.lud16, toast);
                          }
                        }}
                        size="sm"
                        variant="secondary"
                      >
                        Copy
                      </Button>
                    </CardFooter>
                  </Card>
                )}
                <Card>
                  <CardHeader>
                    <CardTitle>{name}</CardTitle>
                    <CardDescription>
                      Balance:{" "}
                      {new Intl.NumberFormat().format(
                        Math.floor(app.balance / 1000)
                      )}{" "}
                      sats
                    </CardDescription>
                  </CardHeader>
                  <CardFooter className="flex flex-row justify-end">
                    <IsolatedAppTopupDialog appId={app.id}>
                      <Button size="sm" variant="secondary">
                        Top Up
                      </Button>
                    </IsolatedAppTopupDialog>
                  </CardFooter>
                </Card>
                <Button onClick={() => setStep(2)}>Next</Button>
              </div>
            </div>
          )}
          {step === 2 && (
            <div className="grid gap-5">
              <div>
                Select which apps to connect to this sub-wallet — whether for
                yourself or someone you're inviting. Connect to as many services
                as you want:
              </div>
              <Alert variant="destructive">
                <AlertCircle className="h-4 w-4" />
                <AlertTitle>Important</AlertTitle>
                <AlertDescription>
                  For your security, these connection details are only visible
                  now and{" "}
                  <span className="font-semibold">
                    cannot be retrieved later
                  </span>
                  . If needed, you can store them in a password manager for
                  future reference.
                </AlertDescription>
              </Alert>
              <Accordion type="single" collapsible>
                <AccordionItem value="mobile">
                  <AccordionTrigger>Alby Go</AccordionTrigger>
                  <AccordionContent>
                    <ul className="flex flex-col gap-1 list-inside list-decimal mb-6">
                      <li>
                        Download Alby Go from the app store
                        <div className="flex flex-row gap-3 my-2">
                          <Popover>
                            <PopoverTrigger>
                              <Button variant="outline">
                                <PlayStoreIcon className="w-4 h-4 mr-2" />
                                Play Store
                              </Button>
                            </PopoverTrigger>
                            <PopoverContent className="flex flex-col items-center gap-3">
                              <QRCode value="https://play.google.com/store/apps/details?id=com.getalby.mobile" />
                              <ExternalLinkButton
                                variant="link"
                                to="https://play.google.com/store/apps/details?id=com.getalby.mobile"
                              >
                                Open
                                <ExternalLinkIcon className="w-4 h-4 ml-2" />
                              </ExternalLinkButton>
                            </PopoverContent>
                          </Popover>
                          <Popover>
                            <PopoverTrigger>
                              <Button variant="outline">
                                <AppleIcon className="w-4 h-4 mr-2" />
                                Apple App Store
                              </Button>
                            </PopoverTrigger>
                            <PopoverContent className="flex flex-col items-center gap-3">
                              <QRCode value="https://apps.apple.com/us/app/alby-go/id6471335774" />
                              <ExternalLinkButton
                                variant="link"
                                to="https://apps.apple.com/us/app/alby-go/id6471335774"
                              >
                                Open
                                <ExternalLinkIcon className="w-4 h-4 ml-2" />
                              </ExternalLinkButton>
                            </PopoverContent>
                          </Popover>
                          <Popover>
                            <PopoverTrigger>
                              <Button variant="outline">
                                <ZapStoreIcon className="w-4 h-4 mr-2" />
                                Zapstore
                              </Button>
                            </PopoverTrigger>
                            <PopoverContent className="flex flex-col items-center gap-3">
                              <div className="text-center text-xs text-muted-foreground">
                                Install Zapstore on your Android device and
                                search for{" "}
                                <span className="font-semibold">Alby Go</span>
                              </div>
                              <QRCode value="https://zapstore.dev" />
                              <ExternalLinkButton
                                variant="link"
                                to="https://zapstore.dev"
                              >
                                Open
                                <ExternalLinkIcon className="w-4 h-4 ml-2" />
                              </ExternalLinkButton>
                            </PopoverContent>
                          </Popover>
                        </div>
                      </li>
                      <li>Open Alby Go and scan this QR code</li>
                    </ul>
                    {app && (
                      <ConnectAppCard app={app} pairingUri={connectionSecret} />
                    )}
                  </AccordionContent>
                </AccordionItem>
                <AccordionItem value="account">
                  <AccordionTrigger>Alby Account</AccordionTrigger>
                  <AccordionContent>
                    <ul className="flex flex-col gap-1 list-inside list-decimal mb-6">
                      <li>
                        If the recipient doesn't have an Alby Account yet, ask
                        them to create an Alby Account on getalby.com or send
                        them an{" "}
                        <ExternalLink
                          to="https://getalby.com/invite_codes"
                          className="underline"
                        >
                          invite
                        </ExternalLink>
                      </li>
                      <li>
                        Send your recipient this connection link after they've
                        created an Alby account. The link will automatically
                        connect the wallet to their account — no extra steps
                        needed.
                      </li>
                    </ul>
                    <div className="flex gap-2">
                      <Input
                        disabled
                        readOnly
                        type="password"
                        value={albyAccountUrl}
                      />
                      <Button
                        onClick={() => copyToClipboard(albyAccountUrl, toast)}
                        variant="outline"
                      >
                        <CopyIcon className="w-4 h-4 mr-2" />
                        Copy URL
                      </Button>
                    </div>
                    <Alert className="mt-5">
                      <AlertTriangleIcon className="h-4 w-4" />
                      <AlertTitle>
                        Use separate browsers or profiles when managing multiple
                        Alby accounts
                      </AlertTitle>
                      <AlertDescription>
                        Opening a connection URL while logged into a different
                        Alby account will override your currently connected
                        wallet.
                      </AlertDescription>
                    </Alert>
                  </AccordionContent>
                </AccordionItem>
                <AccordionItem value="extension">
                  <AccordionTrigger>Alby Browser Extension</AccordionTrigger>
                  <AccordionContent>
                    <ul className="list-inside list-decimal flex flex-col gap-1 mb-6">
                      <li>
                        Install Alby Browser Extension from{" "}
                        <ExternalLink
                          className="underline"
                          to="https://chromewebstore.google.com/detail/alby-bitcoin-wallet-for-l/iokeahhehimjnekafflcihljlcjccdbe"
                        >
                          Chrome Web Store
                        </ExternalLink>{" "}
                        or{" "}
                        <ExternalLink
                          className="underline"
                          to="https://addons.mozilla.org/en-US/firefox/addon/alby/"
                        >
                          Firefox Add-ons
                        </ExternalLink>
                      </li>
                      <li>
                        In the extension, choose{" "}
                        <span className="font-medium">
                          Bring Your Own Wallet
                        </span>{" "}
                        while connecting a wallet
                      </li>
                      <li>
                        Choose{" "}
                        <span className="font-medium">
                          Nostr Wallet Connect
                        </span>{" "}
                        option
                      </li>
                      <li>
                        Copy & paste the connection secret below to connect the
                        sub-wallet to the extension.
                      </li>
                    </ul>
                    <div className="flex gap-2">
                      <Input
                        disabled
                        readOnly
                        type="password"
                        value={connectionSecret}
                      />
                      <Button
                        onClick={() => copyToClipboard(connectionSecret, toast)}
                        variant="outline"
                      >
                        <CopyIcon className="w-4 h-4 mr-2" />
                        Copy
                      </Button>
                    </div>
                  </AccordionContent>
                </AccordionItem>
                <AccordionItem value="other">
                  <AccordionTrigger>Other apps</AccordionTrigger>
                  <AccordionContent>
                    <ul className="list-inside list-decimal flex flex-col gap-1 mb-6">
                      <li>
                        Open the app you wish to connect sub-wallet to and look
                        for a way to attach a wallet (most apps provide this
                        option in settings).
                      </li>
                      <li>Copy and paste the connection secret from below</li>
                    </ul>
                    <div className="grid gap-1.5">
                      <Label htmlFor="connectionSecret">
                        Connection Secret
                      </Label>
                      <div className="flex gap-2">
                        <Input
                          id="connectionSecret"
                          disabled
                          readOnly
                          type="password"
                          value={connectionSecret}
                        />
                        <Button
                          onClick={() =>
                            copyToClipboard(connectionSecret, toast)
                          }
                          variant="outline"
                        >
                          <CopyIcon className="w-4 h-4 mr-2" />
                          Copy
                        </Button>
                      </div>
                    </div>
                  </AccordionContent>
                </AccordionItem>
                <AccordionItem value="podcasting">
                  <AccordionTrigger>Podcasting 2.0</AccordionTrigger>
                  <AccordionContent>
                    <p className="text-muted-foreground text-sm mb-5">
                      To allow receiving podcasting 2.0 payments to the
                      sub-wallet, make sure to share the podcast:value tag that
                      needs to be added to recipient's RSS feed:
                    </p>
                    <div className="flex flex-col gap-2">
                      <div className="grid gap-1.5">
                        <Label htmlFor="podcastValueTag">
                          podcast:value tag
                        </Label>
                        <Textarea
                          id="podcastValueTag"
                          readOnly
                          className="h-32"
                          value={valueTag}
                        />
                      </div>
                      <Button
                        onClick={() => copyToClipboard(valueTag, toast)}
                        variant="outline"
                      >
                        <CopyIcon className="w-4 h-4 mr-2" />
                        Copy
                      </Button>
                      <Alert>
                        <AlertTitle className="flex flex-row gap-2">
                          <TriangleAlert className="w-4 h-4" />
                          Make sure you also connect other options
                        </AlertTitle>
                        <AlertDescription>
                          This connection will allow them to receive payments
                          from their podcast audience. To allow them to spend
                          the funds, make sure they are connected to other
                          options like Alby Go or Alby Account.
                        </AlertDescription>
                      </Alert>
                    </div>
                  </AccordionContent>
                </AccordionItem>
              </Accordion>
              <div className="flex gap-2">
                <Button onClick={() => setStep(1)} variant="secondary">
                  Back
                </Button>
                <Link to="/sub-wallets">
                  <Button>Finish</Button>
                </Link>
              </div>
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
