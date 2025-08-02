import React from "react";
import {
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from "src/components/ui/alert-dialog";
import { Button } from "src/components/ui/button";
import { Textarea } from "src/components/ui/textarea";
import { useToast } from "src/components/ui/use-toast";
import { useInfo } from "src/hooks/useInfo";
import { request } from "src/utils/request";

type ExecuteCustomNodeCommandDialogContentProps = {
  availableCommands: string;
  setCommandResponse: (response: string) => void;
};

export function ExecuteCustomNodeCommandDialogContent({
  setCommandResponse,
  availableCommands,
}: ExecuteCustomNodeCommandDialogContentProps) {
  const { mutate: reloadInfo } = useInfo();
  const { toast } = useToast();
  const [command, setCommand] = React.useState<string>();

  let parsedAvailableCommands = availableCommands;
  try {
    parsedAvailableCommands = JSON.stringify(
      JSON.parse(availableCommands).commands,
      null,
      2
    );
  } catch (error) {
    // ignore unexpected json
  }

  async function onSubmit(e: React.FormEvent) {
    e.preventDefault();
    try {
      if (!command) {
        throw new Error("No command set");
      }
      const result = await request("/api/command", {
        method: "POST",
        body: JSON.stringify({ command }),
        headers: {
          "Content-Type": "application/json",
        },
      });
      await reloadInfo();

      const parsedResponse = JSON.stringify(result);
      setCommandResponse(parsedResponse);

      toast({ title: "Command executed", description: parsedResponse });
    } catch (error) {
      console.error(error);
      toast({
        variant: "destructive",
        title: "Something went wrong: " + error,
      });
    }
  }

  return (
    <AlertDialogContent>
      <form onSubmit={onSubmit}>
        <AlertDialogHeader>
          <AlertDialogTitle>Execute Custom Node Command</AlertDialogTitle>
          <AlertDialogDescription className="text-left">
            <Textarea
              className="h-36 font-mono"
              value={command}
              onChange={(e) => setCommand(e.target.value)}
              placeholder="commandname --arg1=value1"
            />
            <p className="mt-2">Available commands</p>
            <Textarea
              readOnly
              className="mt-2 font-mono"
              value={parsedAvailableCommands}
              rows={10}
            />
          </AlertDialogDescription>
        </AlertDialogHeader>
        <AlertDialogFooter className="mt-4">
          <AlertDialogCancel onClick={() => setCommand("")}>
            Cancel
          </AlertDialogCancel>
          <Button type="submit">Execute</Button>
        </AlertDialogFooter>
      </form>
    </AlertDialogContent>
  );
}
