import { LucideIcon } from "lucide-react";
import { Link } from "react-router-dom";
import FormattedFiatAmount from "src/components/FormattedFiatAmount";
import { Button } from "src/components/ui/button";
import {
  Card,
  CardContent,
  CardFooter,
  CardHeader,
  CardTitle,
} from "src/components/ui/card";

type BalanceCardProps = {
  title: string;
  balance: number;
  buttonTitle: string;
  buttonLink: string;
  hasChannelManagement: boolean;
  BalanceCardIcon: LucideIcon;
};

function BalanceCard({
  title,
  balance,
  buttonTitle,
  buttonLink,
  hasChannelManagement,
  BalanceCardIcon,
}: BalanceCardProps) {
  return (
    <Card className="w-full hidden md:block self-start">
      <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
        <CardTitle className="text-sm font-medium">{title}</CardTitle>
        <BalanceCardIcon className="h-4 w-4 text-muted-foreground" />
      </CardHeader>
      <CardContent>
        <div className="text-2xl font-bold balance sensitive slashed-zero">
          {new Intl.NumberFormat().format(Math.floor(balance / 1000))} sats
        </div>
        <FormattedFiatAmount
          amount={Math.floor(balance / 1000)}
          className="text-muted-foreground"
        />
      </CardContent>
      {hasChannelManagement && (
        <CardFooter className="flex justify-end">
          <Link to={buttonLink}>
            <Button variant="outline">{buttonTitle}</Button>
          </Link>
        </CardFooter>
      )}
    </Card>
  );
}

export default BalanceCard;
