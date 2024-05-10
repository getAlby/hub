import { PowerCircleIcon } from "lucide-react";
import React, { ChangeEvent, useState } from "react";
import { useNavigate } from "react-router-dom";
import Loading from "src/components/Loading";
import TwoColumnLayoutHeader from "src/components/TwoColumnLayoutHeader";
import { Input } from "src/components/ui/input";
import { Label } from "src/components/ui/label";
import { LoadingButton } from "src/components/ui/loading-button";
import { useToast } from "src/components/ui/use-toast";
import { useCSRF } from "src/hooks/useCSRF";
import { useInfo } from "src/hooks/useInfo";
import { handleRequestError } from "src/utils/handleRequestError";
import { request } from "src/utils/request";

export function RestoreNode() {
  const navigate = useNavigate();
  const { toast } = useToast();
  const { data: csrf } = useCSRF();
  const [unlockPassword, setUnlockPassword] = useState("");
  const [file, setFile] = useState<File | null>(null);

  const [loading, setLoading] = useState(false);
  const [restored, setRestored] = useState(false);
  const { data: info } = useInfo(restored);
  const isHttpMode = window.location.protocol.startsWith("http");

  React.useEffect(() => {
    if (restored && info?.setupCompleted) {
      navigate("/");
    }
  }, [info?.setupCompleted, navigate, restored]);

  if (restored) {
    return (
      <div className="flex flex-col gap-5 items-center">
        <TwoColumnLayoutHeader
          title="Restart your Hub"
          description="Alby Hub needs to restart to finish restoring your node"
        />
        <PowerCircleIcon className="w-32 h-32" />
        <p className="max-w-sm text-center">
          If you're running in the cloud, your Alby Hub will restart
          automatically. Otherwise, please manually restart your Alby Hub to
          finish the restore process.
        </p>
        <div className="flex items-center gap-2 text-muted-foreground">
          <Loading /> <p>Waiting for restart...</p>
        </div>
      </div>
    );
  }

  const onSubmit = async (e: React.FormEvent) => {
    alert(
      "As part of the node restore process your Alby Hub will be shut down. If you're running in the cloud, your Alby Hub will restart automatically. Otherwise, please manually restart your Alby Hub to finish the restore process."
    );

    e.preventDefault();
    if (!csrf) {
      throw new Error("No CSRF token");
    }

    try {
      setLoading(true);

      if (isHttpMode) {
        const formData = new FormData();
        formData.append("unlockPassword", unlockPassword);
        if (file !== null) {
          formData.append("backup", file);
        }
        await request("/api/restore", {
          method: "POST",
          headers: {
            "X-CSRF-Token": csrf,
          },
          body: formData,
        });
      } else {
        await request("/api/restore", {
          method: "POST",
          headers: {
            "X-CSRF-Token": csrf,
          },
          body: JSON.stringify({
            unlockPassword,
          }),
        });
      }

      setRestored(true);
    } catch (error) {
      handleRequestError(toast, "Failed to restore backup", error);
    } finally {
      setLoading(false);
    }
  };

  const handleChangeFile = (e: ChangeEvent<HTMLInputElement>) => {
    const files = e.currentTarget.files;
    if (files) setFile(files[0]);
  };

  return (
    <>
      <form
        onSubmit={onSubmit}
        className="flex flex-col gap-5 mx-auto max-w-2xl text-sm"
      >
        <TwoColumnLayoutHeader
          title="Restore Node Backup"
          description="Restore a previously created backup to recover your node."
        />
        <div className="grid gap-2">
          <Label htmlFor="password">Unlock Password</Label>
          <Input
            type="password"
            name="password"
            required
            onChange={(e) => setUnlockPassword(e.target.value)}
            value={unlockPassword}
            placeholder="Unlock Password"
          />
        </div>
        {isHttpMode && (
          <div className="grid gap-2">
            <Label htmlFor="backup">Backup File</Label>
            <Input
              type="file"
              required
              name="backup"
              accept=".bkp"
              onChange={handleChangeFile}
              className="cursor-pointer"
            />
          </div>
        )}
        <LoadingButton loading={loading}>Restore Node</LoadingButton>
      </form>
    </>
  );
}
