import React from "react";
import { Link, useNavigate } from "react-router-dom";
import SettingsHeader from "src/components/SettingsHeader";
import { Button } from "src/components/ui/button";
import { Input } from "src/components/ui/input";
import { Label } from "src/components/ui/label";
import { LoadingButton } from "src/components/ui/loading-button";
import { useToast } from "src/components/ui/use-toast";
import { useInfo } from "src/hooks/useInfo";
import { MnemonicResponse } from "src/types";
import { request } from "src/utils/request";

export default function Backup() {
  const navigate = useNavigate();

  const { toast } = useToast();
  const { hasNodeBackup, hasMnemonic } = useInfo();

  const [unlockPassword, setUnlockPassword] = React.useState("");

  const [loading, setLoading] = React.useState(false);

  const onSubmitPassword = async (e: React.FormEvent) => {
    e.preventDefault();
    try {
      setLoading(true);
      const result = await request<MnemonicResponse>("/api/mnemonic", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({
          unlockPassword,
        }),
      });

      if (result?.mnemonic) {
        navigate("/settings/mnemonic-backup");
      }
    } catch (error) {
      toast({
        title: "Incorrect password",
        description: "Failed to decrypt mnemonic.",
        variant: "destructive",
      });
    } finally {
      setLoading(false);
    }
  };

  return (
    <>
      <SettingsHeader
        title="Backup"
        description="Backup your wallet recovery phrase and your channels state."
      />

      {hasMnemonic && (
        <div className="mb-12">
          <div>
            <h3 className="text-lg font-medium">Wallet Keys Backup</h3>
            <p className="text-sm text-muted-foreground">
              Key recovery phrase is a group of 12 random words that are the
              only way to recover access to your wallet on another machine or
              when you lose your unlock password.
            </p>
          </div>
          <form
            onSubmit={onSubmitPassword}
            className="max-w-md flex flex-col gap-3 mt-8"
          >
            <div className="grid gap-1.5 mb-8">
              <Label htmlFor="password">Password</Label>
              <Input
                type="password"
                name="password"
                onChange={(e) => setUnlockPassword(e.target.value)}
                value={unlockPassword}
                placeholder="Password"
              />
              <p className="text-sm text-muted-foreground">
                Enter your unlock password to view your recovery phrase.
              </p>
            </div>
            <div className="flex justify-start">
              <LoadingButton loading={loading} variant="secondary">
                View Recovery Phrase
              </LoadingButton>
            </div>
          </form>
        </div>
      )}
      {hasNodeBackup && (
        <div>
          <h3 className="text-lg font-medium">Channels Backup</h3>
          <p className="text-sm text-muted-foreground mb-8">
            If you’d like to import or host your Hub on another host or machine,
            you’ll need your channels’ backup file to import your channels
            state. This instance of Hub will be stopped.
          </p>
          <Link to="/settings/node-backup">
            <Button variant="secondary">Create Channels Backup</Button>
          </Link>
        </div>
      )}

      {!hasMnemonic && !hasNodeBackup && (
        <p className="text-sm text-muted-foreground">
          No wallet recovery phrase or channel state backup present.
        </p>
      )}
    </>
  );
}
