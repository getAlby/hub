import { CircleCheck, CopyIcon } from "lucide-react";
import { Link } from "react-router-dom";
import { Button } from "src/components/ui/button";
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
} from "src/components/ui/card";
import { useToast } from "src/components/ui/use-toast";
import { copyToClipboard } from "src/lib/clipboard";

type PaymentSuccessCardProps = {
  preimage: string;
  onReset: () => void;
};

function PaymentSuccessCard({ preimage, onReset }: PaymentSuccessCardProps) {
  const { toast } = useToast();

  const copy = () => {
    copyToClipboard(preimage);
    toast({ title: "Copied to clipboard." });
  };

  return (
    <>
      <Card className="w-full">
        <CardHeader>
          <CardTitle className="text-center">Payment Successful</CardTitle>
        </CardHeader>
        <CardContent className="flex flex-col items-center gap-4">
          <CircleCheck className="w-32 h-32 mb-2" />
          <Button onClick={copy} variant="outline">
            <CopyIcon className="w-4 h-4 mr-2" />
            Copy Preimage
          </Button>
        </CardContent>
      </Card>
      <Button className="mt-4 w-full" onClick={onReset}>
        Make Another Payment
      </Button>
      <Link to="/wallet">
        <Button className="mt-4 w-full" variant="secondary">
          Back To Wallet
        </Button>
      </Link>
    </>
  );
}

export default PaymentSuccessCard;
