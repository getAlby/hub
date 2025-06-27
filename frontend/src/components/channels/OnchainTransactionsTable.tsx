import dayjs from "dayjs";
import { ArrowDownIcon, ArrowUpIcon } from "lucide-react";
import FormattedFiatAmount from "src/components/FormattedFiatAmount";
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
} from "src/components/ui/card";
import { Table, TableBody, TableCell, TableRow } from "src/components/ui/table";
import { useInfo } from "src/hooks/useInfo";
import { useOnchainTransactions } from "src/hooks/useOnchainTransactions";
import { cn } from "src/lib/utils";

export function OnchainTransactionsTable() {
  const { data: info } = useInfo();
  const { data: transactions } = useOnchainTransactions();

  if (!transactions?.length) {
    return null;
  }

  return (
    <Card className="mt-6">
      <CardHeader>
        <CardTitle className="text-2xl">On-Chain Transactions</CardTitle>
      </CardHeader>
      <CardContent>
        <Table>
          <TableBody>
            {transactions.map((tx) => {
              const Icon = tx.type == "outgoing" ? ArrowUpIcon : ArrowDownIcon;
              return (
                <TableRow
                  key={tx.txId}
                  className="cursor-pointer"
                  onClick={() => {
                    window.open(`${info?.mempoolUrl}/tx/${tx.txId}`, "_blank");
                  }}
                >
                  <TableCell className="flex items-center gap-2">
                    <div
                      className={cn(
                        "flex justify-center items-center rounded-full w-10 h-10 relative",
                        tx.state === "unconfirmed"
                          ? "bg-blue-100 dark:bg-sky-950 animate-pulse"
                          : tx.type === "outgoing"
                            ? "bg-orange-100 dark:bg-amber-950"
                            : "bg-green-100 dark:bg-emerald-950"
                      )}
                      title={`${tx.numConfirmations} confirmations`}
                    >
                      <Icon
                        strokeWidth={3}
                        className={cn(
                          "w-6 h-6",
                          tx.state === "unconfirmed"
                            ? "stroke-blue-500 dark:stroke-sky-500"
                            : tx.type === "outgoing"
                              ? "stroke-orange-500 dark:stroke-amber-500"
                              : "stroke-green-500 dark:stroke-teal-500"
                        )}
                      />
                    </div>
                    <div className="md:flex md:gap-2 md:items-center">
                      <p className="font-semibold text-lg">
                        {tx.type == "outgoing"
                          ? tx.state === "confirmed"
                            ? "Sent"
                            : "Sending"
                          : tx.state === "confirmed"
                            ? "Received"
                            : "Receiving"}
                      </p>
                      <p
                        className="text-muted-foreground"
                        title={dayjs(tx.createdAt * 1000)
                          .local()
                          .format("D MMMM YYYY, HH:mm")}
                      >
                        {dayjs(tx.createdAt * 1000)
                          .local()
                          .fromNow()}
                      </p>
                    </div>
                  </TableCell>

                  <TableCell>
                    <div className="flex flex-col items-end">
                      <div className="flex flex-row gap-1">
                        <p
                          className={cn(
                            tx.type == "incoming" &&
                              "text-green-600 dark:text-emerald-500"
                          )}
                        >
                          {tx.type == "outgoing" ? "-" : "+"}
                          <span className="font-medium">
                            {new Intl.NumberFormat().format(tx.amountSat)}
                          </span>
                        </p>
                        <p className="text-muted-foreground">
                          {tx.amountSat == 1 ? "sat" : "sats"}
                        </p>
                      </div>
                      <FormattedFiatAmount
                        className="text-xs"
                        amount={tx.amountSat}
                      />
                    </div>
                  </TableCell>
                </TableRow>
              );
            })}
          </TableBody>
        </Table>
      </CardContent>
    </Card>
  );
}
