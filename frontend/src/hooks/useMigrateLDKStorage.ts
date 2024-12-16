import React from "react";
import { useToast } from "src/components/ui/use-toast";
import { request } from "src/utils/request";

export function useMigrateLDKStorage() {
  const [isMigratingStorage, setMigratingStorage] = React.useState(false);
  const { toast } = useToast();

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
      toast({
        title: "Please unlock your hub",
      });
    } catch (e) {
      console.error(e);
      toast({
        title: "Could not start hub storage migration: " + e,
        variant: "destructive",
      });
    }
    setMigratingStorage(false);
  };

  return { isMigratingStorage, migrateLDKStorage };
}