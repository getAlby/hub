import React from "react";
import { useNavigate } from "react-router-dom";
import Container from "src/components/Container";
import TwoColumnLayoutHeader from "src/components/TwoColumnLayoutHeader";
import { Button } from "src/components/ui/button";
import { Input } from "src/components/ui/input";
import { Label } from "src/components/ui/label";
import { useToast } from "src/components/ui/use-toast";
import useSetupStore from "src/state/SetupStore";

export function BreezForm() {
  const { toast } = useToast();
  const navigate = useNavigate();
  const setupStore = useSetupStore();
  const [greenlightInviteCode, setGreenlightInviteCode] =
    React.useState<string>(setupStore.nodeInfo.greenlightInviteCode || "");
  const [breezApiKey, setBreezApiKey] = React.useState<string>(
    setupStore.nodeInfo.breezApiKey || ""
  );

  async function handleSubmit(data: object) {
    setupStore.updateNodeInfo({
      backendType: "BREEZ",
      ...data,
    });
    navigate("/setup/import-mnemonic");
  }

  function onSubmit(e: React.FormEvent) {
    e.preventDefault();
    if (!greenlightInviteCode || !breezApiKey) {
      toast({
        title: "Please fill out all fields",
        variant: "destructive",
      });
      return;
    }
    handleSubmit({
      greenlightInviteCode,
      breezApiKey,
    });
  }

  return (
    <Container>
      <TwoColumnLayoutHeader
        title={"Configure Breez"}
        description={"Fill out wallet details to finish setup."}
      />
      <form className="w-full grid gap-5 mt-6" onSubmit={onSubmit}>
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
        <div className="grid gap-1.5">
          <Label htmlFor="breez-api-key">Breez API Key</Label>
          <Input
            name="breez-api-key"
            onChange={(e) => setBreezApiKey(e.target.value)}
            value={breezApiKey}
            autoComplete="off"
            type="text"
            id="breez-api-key"
          />
        </div>
        <Button type="submit">Next</Button>
      </form>
    </Container>
  );
}
