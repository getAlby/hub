import React from "react";
import { useNavigate } from "react-router-dom";
import Container from "src/components/Container";
import TwoColumnLayoutHeader from "src/components/TwoColumnLayoutHeader";
import { Button } from "src/components/ui/button";
import { Input } from "src/components/ui/input";
import { Label } from "src/components/ui/label";
import { useToast } from "src/components/ui/use-toast";
import useSetupStore from "src/state/SetupStore";

export function GreenlightForm() {
  const { toast } = useToast();
  const navigate = useNavigate();
  const setupStore = useSetupStore();
  const [greenlightInviteCode, setGreenlightInviteCode] =
    React.useState<string>(setupStore.nodeInfo.greenlightInviteCode || "");

  async function handleSubmit(data: object) {
    setupStore.updateNodeInfo({
      backendType: "GREENLIGHT",
      ...data,
    });
    navigate("/setup/import-mnemonic");
  }

  function onSubmit(e: React.FormEvent) {
    e.preventDefault();
    if (!greenlightInviteCode) {
      toast({
        title: "Please fill out all fields",
        variant: "destructive",
      });
      return;
    }
    handleSubmit({
      greenlightInviteCode,
    });
  }

  return (
    <Container>
      <TwoColumnLayoutHeader
        title="Configure Greenlight"
        description="Fill out wallet details to finish setup."
      />
      <form onSubmit={onSubmit} className="w-full grid gap-5 mt-6">
        <div className="grid gap-1.5">
          <Label htmlFor="greenlight-invite-code">Greenlight Invite Code</Label>
          <Input
            name="greenlight-invite-code"
            onChange={(e) => setGreenlightInviteCode(e.target.value)}
            value={greenlightInviteCode}
            type="text"
            id="greenlight-invite-code"
            placeholder="XXXX-YYYY"
          />
        </div>
        <Button>Next</Button>
      </form>
    </Container>
  );
}
