import React from "react";
import AppHeader from "src/components/AppHeader";
import AppCard from "src/components/connections/AppCard";
import { Button } from "src/components/ui/button";
import {
  Card,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from "src/components/ui/card";
import { useToast } from "src/components/ui/use-toast";
import { useApps } from "src/hooks/useApps";
import { createApp } from "src/requests/createApp";
import { CreateAppRequest } from "src/types";
import { handleRequestError } from "src/utils/handleRequestError";

export function Zapplanner() {
  const { data: apps, mutate: reloadApps } = useApps();
  const { toast } = useToast();
  const [, setLoading] = React.useState(false);
  const handleSubmit = async (name: string) => {
    setLoading(true);

    try {
      if (apps?.some((existingApp) => existingApp.name === name)) {
        throw new Error("A connection with the same name already exists.");
      }

      const createAppRequest: CreateAppRequest = {
        name,
        scopes: [
          "get_balance",
          "get_info",
          "list_transactions",
          "lookup_invoice",
          "make_invoice",
          "notifications",
          "pay_invoice",
        ],
        isolated: true,
        metadata: {
          app_store_app_id: "zapplanner",
        },
      };

      await createApp(createAppRequest);
      toast({
        title: "Created subscription",
        description: "The first payment is scheduled for 1st of October.",
      });

      reloadApps();
    } catch (error) {
      handleRequestError(toast, "Failed to create app", error);
    }
    setLoading(false);
  };

  const zapplannerApps = apps?.filter(
    (app) => app.metadata?.app_store_app_id === "zapplanner"
  );

  return (
    <div className="grid gap-5">
      <AppHeader title="Zapplanner" description="Schedule payments" />
      <h2 className="font-semibold text-xl">Inspiration</h2>
      <p className="text-muted-foreground -mt-5">
        Be the change you want to see in the world and do your part
      </p>
      <div className="grid grid-cols-3 gap-3">
        <Card>
          <CardHeader>
            <CardTitle>Alby</CardTitle>
            <CardDescription>description</CardDescription>
          </CardHeader>
          <CardFooter className="flex justify-end">
            <Button
              onClick={() => {
                handleSubmit("Alby");
              }}
            >
              Support with 5$ / month
            </Button>
          </CardFooter>
        </Card>
        <Card>
          <CardHeader>
            <CardTitle>HRF</CardTitle>
            <CardDescription>description</CardDescription>
          </CardHeader>
          <CardFooter className="flex justify-end">
            <Button>Support with 5$ / month</Button>
          </CardFooter>
        </Card>
        <Card>
          <CardHeader>
            <CardTitle>opensats</CardTitle>
            <CardDescription>description</CardDescription>
          </CardHeader>
          <CardFooter className="flex justify-end">
            <Button>Support with 5$ / month</Button>
          </CardFooter>
        </Card>
      </div>
      {/* {!connectionSecret && (
        <>
          <form
            onSubmit={handleSubmit}
            className="flex flex-col items-start gap-5 max-w-lg"
          >
            <div className="w-full grid gap-1.5">
              <Label htmlFor="name">Name of friend or family member</Label>
              <Input
                autoFocus
                type="text"
                name="name"
                value={name}
                id="name"
                onChange={(e) => setName(e.target.value)}
                required
                autoComplete="off"
              />
            </div>
            <LoadingButton loading={isLoading} type="submit">
              Create Subaccount
            </LoadingButton>
          </form> */}

      {!!zapplannerApps?.length && (
        <>
          <h2 className="font-semibold text-xl">Recurring Payments</h2>
          <div className="grid grid-cols-1 lg:grid-cols-2 gap-4 items-stretch app-list">
            {zapplannerApps.map((app, index) => (
              <AppCard key={index} app={app} />
            ))}
          </div>
        </>
      )}
    </div>
  );
}
