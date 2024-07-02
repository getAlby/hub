import dayjs from "dayjs";
import { ArrowDownIcon, ArrowUpIcon, Drum } from "lucide-react";
import EmptyState from "src/components/EmptyState";

import Loading from "src/components/Loading";
import { useTransactions } from "src/hooks/useTransactions";

function TransactionsList() {
  const { data: transactions, isLoading } = useTransactions();

  if (isLoading) {
    return <Loading />;
  }

  return (
    <div>
      {!transactions?.length ? (
        <EmptyState
          icon={Drum}
          title="No transactions yet"
          description="Your most recent incoming and outgoing payments will show up here."
          buttonText="Receive Your First Payment"
          buttonLink="/wallet/receive"
        />
      ) : (
        <>
          {transactions?.map((tx, i) => {
            const type = tx.type;

            return (
              <div
                key={`tx-${i}`}
                className="p-3 mb-4 rounded-md"
                // TODO: Add modal onclick to show payment details
              >
                <div className="flex gap-3">
                  <div className="flex items-center">
                    {type == "outgoing" ? (
                      <div
                        className={
                          "flex justify-center items-center bg-orange-100 dark:bg-orange-950 rounded-full w-10 h-10 md:w-14 md:h-14"
                        }
                      >
                        <ArrowUpIcon
                          strokeWidth={3}
                          className="w-6 h-6 md:w-8 md:h-8 text-orange-400 dark:text-amber-600 stroke-orange-400 dark:stroke-amber-600"
                        />
                      </div>
                    ) : (
                      <div className="flex justify-center items-center bg-green-100 dark:bg-emerald-950 rounded-full w-10 h-10 md:w-14 md:h-14">
                        <ArrowDownIcon
                          strokeWidth={3}
                          className="w-6 h-6 md:w-8 md:h-8 text-green-500 dark:text-emerald-500 stroke-green-400 dark:stroke-emerald-500"
                        />
                      </div>
                    )}
                  </div>
                  <div className="overflow-hidden mr-3">
                    <div className="flex items-center gap-2 truncate dark:text-white">
                      <p className="text-lg md:text-xl font-semibold">
                        {type == "incoming" ? "Received" : "Sent"}
                      </p>
                      <p className="text-sm md:text-base truncate text-muted-foreground">
                        {dayjs(tx.settled_at).fromNow()}
                      </p>
                    </div>
                    <p className="text-sm md:text-base text-muted-foreground">
                      {tx.description || "Lightning invoice"}
                    </p>
                  </div>
                  <div className="flex ml-auto text-right space-x-3 shrink-0 dark:text-white">
                    <div className="flex items-center gap-2 text-xl">
                      <p
                        className={`font-semibold ${
                          type == "incoming" &&
                          "text-green-600 dark:color-green-400"
                        }`}
                      >
                        {type == "outgoing" ? "-" : "+"}{" "}
                        {Math.floor(tx.amount / 1000)}
                      </p>
                      <p className="text-muted-foreground">sats</p>

                      {/* {!!tx.totalAmountFiat && (
                        <p className="text-xs text-gray-400 dark:text-neutral-600">
                          ~{tx.totalAmountFiat}
                        </p>
                      )} */}
                    </div>
                  </div>
                </div>
              </div>
            );
          })}
          {/* {transaction && (
            <TransactionModal
              transaction={transaction}
              isOpen={modalOpen}
              onClose={() => {
                setModalOpen(false);
              }}
            />
          )} */}
        </>
      )}
    </div>
  );
}

export default TransactionsList;
