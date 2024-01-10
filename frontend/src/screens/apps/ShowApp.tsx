import { Navigate, useNavigate, useParams } from "react-router-dom";

import { RequestMethodType, nip47MethodDescriptions } from "@types";
import { useInfo } from "@hooks/useInfo";
import { useApp } from "@hooks/useApp";
import { useCSRF } from "@hooks/useCSRF";
import toast from "@components/Toast";
import Loading from "@components/Loading";
import { handleFetchError, validateFetchResponse } from "@utils/fetch";

function ShowApp() {
  const { data: info } = useInfo();
  const { data: csrf } = useCSRF();
  const { pubkey } = useParams() as { pubkey: string };
  const { data: app } = useApp(pubkey);
  const navigate = useNavigate();

  if (!app || !info) {
    return <Loading />;
  }

  if (app && "error" in app) {
    handleFetchError("Failed to fetch", app.message);
    return <Navigate to="/apps" />;
  }

  const handleDelete = async (event: React.FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    try {
      if (!csrf) {
        throw new Error("No CSRF token");
      }
      const response = await fetch(`/api/apps/${app.nostrPubkey}`, {
        method: "DELETE",
        headers: {
          "Content-Type": "application/json",
          "X-CSRF-Token": csrf,
        },
      });
      await validateFetchResponse(response);
      navigate("/apps");
      toast.success("App disconnected");
    } catch (error) {
      await handleFetchError("Failed to delete app", error);
    }
  };

  return (
    <div>
      <h2 className="font-bold text-2xl font-headline dark:text-white">
        {app ? app.name : "Fetching app..."}
      </h2>
      <p className="text-gray-600 dark:text-neutral-400 text-sm">
        {app ? app.description : ""}
      </p>
      <a className="ml-1 mt-1 mb-4 block dark:text-white text-xs" href="/apps">
        {"<"} Back to overview
      </a>

      <div className="bg-white rounded-md shadow p-4 lg:p-8 dark:bg-surface-02dp">
        <div className="divide-y divide-gray-200 dark:divide-white/10 dark:bg-surface-02dp">
          <div className="pb-4">
            <table>
              <tr>
                <td className="align-top w-32 font-medium dark:text-white">
                  Public Key
                </td>
                <td className="text-gray-600 dark:text-neutral-400 break-all">
                  {app.nostrPubkey}
                </td>
              </tr>
              <tr>
                <td className="align-top font-medium dark:text-white">
                  Last used
                </td>
                <td className="text-gray-600 dark:text-neutral-400">
                  {app.lastEventAt
                    ? new Date(app.lastEventAt).toLocaleDateString()
                    : "never"}
                </td>
              </tr>
              <tr>
                <td className="align-top font-medium dark:text-white">
                  Expires at
                </td>
                <td className="text-gray-600 dark:text-neutral-400">
                  {app.expiresAt
                    ? new Date(app.expiresAt).toLocaleDateString()
                    : "never"}
                </td>
              </tr>
            </table>
          </div>

          <div className="py-4">
            <h3 className="text-xl font-headline dark:text-white">
              Permissions
            </h3>
            <ul className="mt-2 text-gray-600 dark:text-neutral-400">
              {app.requestMethods.map((method: string, index: number) => (
                <li key={index} className="mb-2 relative pl-6">
                  <span className="absolute left-0 text-green-500">✓</span>
                  {nip47MethodDescriptions[method as RequestMethodType]}
                </li>
              ))}
            </ul>
            {app.maxAmount > 0 && (
              <div className="pl-6">
                <table className="text-gray-600 dark:text-neutral-400">
                  <tr>
                    <td className="font-medium">Budget</td>
                    <td>
                      {app.maxAmount} sats ({app.budgetUsage} sats used)
                    </td>
                  </tr>
                  <tr>
                    <td className="font-medium pr-3">Renews</td>
                    <td>{app.budgetRenewal}</td>
                  </tr>
                </table>
              </div>
            )}
          </div>

          <div className="pt-4">
            <h3 className="text-xl font-headline mb-2 dark:text-white">
              ⚠️ Danger zone
            </h3>
            <p className="text-gray-600 dark:text-neutral-400 mb-4">
              This will revoke the permission and will no longer allow calls
              from this public key.
            </p>
          </div>
        </div>

        <form method="post" onSubmit={handleDelete}>
          <button
            type="submit"
            className="inline-flex bg-white border border-red-400 cursor-pointer dark:bg-surface-02dp dark:hover:bg-surface-16dp duration-150 focus-visible:ring-2 focus-visible:ring-offset-2 focus:outline-none font-medium hover:bg-gray-50 items-center justify-center px-5 py-3 rounded-md shadow text-gray-700 dark:text-neutral-300 transition w-full sm:w-[250px] sm:mr-8 mt-8 sm:mt-0 order-last sm:order-first"
          >
            Disconnect
          </button>
        </form>
      </div>
    </div>
  );
}

export default ShowApp;
