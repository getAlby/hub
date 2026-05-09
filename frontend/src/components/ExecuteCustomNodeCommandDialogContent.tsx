import React from "react";
import { toast } from "sonner";
import {
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from "src/components/ui/alert-dialog";
import { Textarea } from "src/components/ui/textarea";
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
  const [command, setCommand] = React.useState<string>();

  let parsedAvailableCommands = availableCommands;
  try {
    parsedAvailableCommands = JSON.stringify(
      JSON.parse(availableCommands).commands,
      null,
      2
    );
  } catch {
    // ignore unexpected json
  }

  async function executeCommand() {
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

      toast("Command executed", { description: parsedResponse });
    } catch (error) {
      console.error(error);
      toast.error("Something went wrong: " + error);
    }
  }

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    executeCommand();
  };

  return (
    <AlertDialogContent>
      <AlertDialogHeader>
        <AlertDialogTitle>Execute Custom Node Command</AlertDialogTitle>
        <AlertDialogDescription className="text-left">
          <form id="execute-command-form" onSubmit={handleSubmit}>
            <Textarea
              className="h-36 font-mono"
              value={command}
              onChange={(e) => setCommand(e.target.value)}
              placeholder="commandname --arg1=value1"
            />
          </form>
          <p className="mt-2">Available commands</p>
          <Textarea
            readOnly
            className="mt-2 font-mono"
            value={parsedAvailableCommands}
            rows={10}
          />
        </AlertDialogDescription>
      </AlertDialogHeader>
      <AlertDialogFooter>
        <AlertDialogCancel onClick={() => setCommand("")}>
          Cancel
        </AlertDialogCancel>
        <AlertDialogAction type="submit" form="execute-command-form">
          Execute
        </AlertDialogAction>
      </AlertDialogFooter>
    </AlertDialogContent>
  );
}
