import SettingsHeader from "src/components/SettingsHeader";
import { Button } from "src/components/ui/button";
import { Input } from "src/components/ui/input";
import { Label } from "src/components/ui/label";
import { useToast } from "src/components/ui/use-toast";

function Swaps() {
  const { toast } = useToast();

  function save() {
    toast({ title: "Saved successfully." });
  }

  return (
    <>
      <SettingsHeader title="Swaps" description={""} />
      <div className="grid gap-1.5">
        <Label>Spending balance threshold</Label>
        <Input placeholder="2M sats" />
      </div>
      <div className="grid gap-1.5">
        <Label>Swap amount</Label>
        <Input placeholder="1M sats" />
      </div>
      <div className="grid gap-1.5">
        <Label>Bitcoin address</Label>
        <Input placeholder="bc1..." />
      </div>
      <div>
        <Button onClick={save}>Save</Button>
      </div>
    </>
  );
}

export default Swaps;
