import { useState } from "react";
import { useNavigate, useParams } from "react-router-dom";
import AppHeader from "src/components/AppHeader";
import { suggestedApps } from "src/components/SuggestedAppData";
import { Button } from "src/components/ui/button";
import { Separator } from "src/components/ui/separator";
import { toast } from "src/components/ui/use-toast";
import { useCSRF } from "src/hooks/useCSRF";
import { AppPermissions, CreateAppResponse, NIP_47_NOTIFICATIONS_PERMISSION, PermissionType, nip47MethodDescriptions } from "src/types";
import { handleRequestError } from "src/utils/handleRequestError";
import { request } from "src/utils/request";
import Permissions from "../../components/Permissions";

// TODO: merge with nip47MethodDescriptions
export const nip47PermissionDescriptions: Record<PermissionType, string> = {
  ...nip47MethodDescriptions,
  [NIP_47_NOTIFICATIONS_PERMISSION]: "Receive wallet notifications",
};

export default function AppDetail() {
  const params = useParams();
  const navigate = useNavigate();
  const { data: csrf } = useCSRF();

  const app = suggestedApps.find(x => x.id == params.id);

  const methods = Object.keys(nip47PermissionDescriptions)
  const requestMethodsSet = new Set<PermissionType>(
    methods as PermissionType[]
  );

  const [permissions, setPermissions] = useState<AppPermissions>({
    requestMethods: requestMethodsSet,
    maxAmount: 100000,
    budgetRenewal: "monthly",
  });

  const handleSubmit = async (event: React.FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    if (!csrf) {
      throw new Error("No CSRF token");
    }

    try {
      const createAppResponse = await request<CreateAppResponse>("/api/apps", {
        method: "POST",
        headers: {
          "X-CSRF-Token": csrf,
          "Content-Type": "application/json",
        },
        body: JSON.stringify({
          name: app?.title,
          ...permissions,
          requestMethods: [...permissions.requestMethods].join(" "),
          expiresAt: permissions.expiresAt?.toISOString(),
        }),
      });

      if (!createAppResponse) {
        throw new Error("no create app response received");
      }

      navigate("/appstore/" + app?.id + "/connect", {
        state: createAppResponse,
      });
      toast({ title: "App created" });
    } catch (error) {
      handleRequestError(toast, "Failed to create app", error);
    }
  };

  if (!app)
    return;

  return (
    <>
      <AppHeader
        title={`Connect to ${app.title}`}
        description="Configure wallet permissions for the app and follow instructions to finalise the connection"
      />
      <form onSubmit={handleSubmit} acceptCharset="UTF-8"
        className="flex flex-col items-start gap-5 max-w-lg">

        <div className="flex flex-row items-center gap-3">
          <img src={app.logo} className="h-12 w-12" />
          <h2 className="font-semibold text-lg">{app.title}</h2>
        </div>

        <div className="flex flex-col gap-2 w-full">
          <p className="font-medium text-sm">Authorize the app to:</p>
          <Permissions
            initialPermissions={permissions}
            onPermissionsChange={setPermissions}
            isEditing
            isNew
          />
        </div>

        <Separator />

        <Button type="submit">
          Create Connection
        </Button>

      </form>
    </>
  );
}
