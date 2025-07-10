import { SparklesIcon } from "lucide-react";
import { Button } from "src/components/ui/button";
import { UpgradeDialog } from "src/components/UpgradeDialog";

interface Props {
  title: string;
  description: string;
}

const UpgradeCard: React.FC<Props> = ({
  title: message,
  description: subMessage,
}) => {
  return (
    <div className="flex flex-1 items-center justify-center rounded-lg border border-dashed shadow-xs p-8">
      <div className="flex flex-col items-center gap-1 text-center max-w-sm">
        <SparklesIcon className="w-12 h-12" />
        <h3 className="mt-4 text-lg font-semibold">{message}</h3>
        <p className="text-sm text-muted-foreground mb-5">{subMessage}</p>
        <UpgradeDialog>
          <Button variant="premium">Upgrade</Button>
        </UpgradeDialog>
      </div>
    </div>
  );
};

export default UpgradeCard;
