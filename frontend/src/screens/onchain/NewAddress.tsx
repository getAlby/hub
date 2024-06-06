import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "@radix-ui/react-tooltip";
import { Copy, QrCode, RefreshCw } from "lucide-react";
import React from "react";
import QRCode from "react-qr-code";
import { Link } from "react-router-dom";
import AppHeader from "src/components/AppHeader";
import Loading from "src/components/Loading";
import {
  Breadcrumb,
  BreadcrumbItem,
  BreadcrumbLink,
  BreadcrumbList,
  BreadcrumbPage,
  BreadcrumbSeparator,
} from "src/components/ui/breadcrumb";
import { Button } from "src/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "src/components/ui/dialog";
import { Input } from "src/components/ui/input";
import { Label } from "src/components/ui/label";
import { LoadingButton } from "src/components/ui/loading-button";
import { localStorageKeys } from "src/constants";
import { useCSRF } from "src/hooks/useCSRF";
import { copyToClipboard } from "src/lib/clipboard";
import { GetOnchainAddressResponse } from "src/types";
import { request } from "src/utils/request";

export default function NewOnchainAddress() {
  const { data: csrf } = useCSRF();
  const [onchainAddress, setOnchainAddress] = React.useState<string>();
  const [isLoading, setLoading] = React.useState(false);

  const getNewAddress = React.useCallback(async () => {
    if (!csrf) {
      return;
    }
    setLoading(true);
    try {
      const response = await request<GetOnchainAddressResponse>(
        "/api/wallet/new-address",
        {
          method: "POST",
          headers: {
            "X-CSRF-Token": csrf,
            "Content-Type": "application/json",
          },
          //body: JSON.stringify({}),
        }
      );
      if (!response?.address) {
        throw new Error("No address in response");
      }
      localStorage.setItem(localStorageKeys.onchainAddress, response.address);
      setOnchainAddress(response.address);
    } catch (error) {
      alert("Failed to request a new address: " + error);
    } finally {
      setLoading(false);
    }
  }, [csrf]);

  React.useEffect(() => {
    const existingAddress = localStorage.getItem(
      localStorageKeys.onchainAddress
    );
    if (existingAddress) {
      setOnchainAddress(existingAddress);
      return;
    }
    getNewAddress();
  }, [getNewAddress]);

  if (!onchainAddress) {
    return (
      <div className="flex justify-center">
        <Loading />
      </div>
    );
  }

  return (
    <div className="grid gap-5">
      <Breadcrumb>
        <BreadcrumbList>
          <BreadcrumbItem>
            <BreadcrumbLink asChild>
              <Link to="/channels">Liquidity</Link>
            </BreadcrumbLink>
          </BreadcrumbItem>
          <BreadcrumbSeparator />
          <BreadcrumbItem>
            <BreadcrumbPage>On-Chain Address</BreadcrumbPage>
          </BreadcrumbItem>
        </BreadcrumbList>
      </Breadcrumb>
      <AppHeader
        title="On-Chain Address"
        description="Deposit bitcoin into your wallet by sending a transaction"
      />
      <div className="grid gap-1.5 max-w-lg">
        <Label htmlFor="text">On-Chain Address</Label>
        <p className="text-xs text-muted-foreground">
          Funds will show up after one confirmation in your savings balance on
          the liquidity page.
        </p>
        <div className="flex flex-row items-center gap-2">
          <Input
            type="text"
            value={onchainAddress}
            className="flex-1"
            readOnly
          />
          <Button
            variant="secondary"
            size="icon"
            onClick={() => {
              copyToClipboard(onchainAddress);
            }}
          >
            <Copy className="w-4 h-4" />
          </Button>
          <Dialog>
            <DialogTrigger asChild>
              <Button variant="secondary" size="icon">
                <QrCode className="w-4 h-4" />
              </Button>
            </DialogTrigger>

            <DialogContent>
              <DialogHeader>
                <DialogTitle>Deposit bitcoin</DialogTitle>
                <DialogDescription>
                  Scan this QR code with your wallet to send funds.
                </DialogDescription>
              </DialogHeader>
              <div className="flex flex-row justify-center p-3">
                <a href={`bitcoin:${onchainAddress}`} target="_blank">
                  <QRCode value={onchainAddress} />
                </a>
              </div>
            </DialogContent>
          </Dialog>
          <TooltipProvider>
            <Tooltip>
              <TooltipTrigger asChild>
                <LoadingButton
                  variant="secondary"
                  size="icon"
                  onClick={getNewAddress}
                  loading={isLoading}
                >
                  <RefreshCw className="w-4 h-4" />
                </LoadingButton>
              </TooltipTrigger>
              <TooltipContent>Generate a new address</TooltipContent>
            </Tooltip>
          </TooltipProvider>
        </div>
      </div>
    </div>
  );
}
