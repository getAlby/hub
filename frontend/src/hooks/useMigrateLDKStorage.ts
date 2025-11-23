import React from "react";
import { toast } from "sonner";
import { useInfo } from "src/hooks/useInfo";
import { request } from "src/utils/request";

export function useMigrateLDKStorage() {
  const [isMigratingStorage, setMigratingStorage] = React.useState(false);
  const { mutate: reloadInfo } = useInfo();

  const migrateLDKStorage = async (to: "VSS") => {
    try {
      setMigratingStorage(true);

      await request("/api/node/migrate-storage", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({
          to,
        }),
      });
      await reloadInfo();
      toast("Please unlock your hub");
    } catch (e) {
      console.error(e);
      toast.error("Could not start hub storage migration: " + e);
    }
    setMigratingStorage(false);
  };

  return { isMigratingStorage, migrateLDKStorage };
}
