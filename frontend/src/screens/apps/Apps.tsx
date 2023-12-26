import { useNavigate } from "react-router-dom";

import Loading from "../../components/Loading";
import { App } from "../../types";
import { useApps } from "../../hooks/useApps";

function Apps() {
  const { data: apps } = useApps();
  const navigate = useNavigate();

  const handleRowClick = (appId: App["nostrPubkey"]) => {
    navigate(`/apps/${appId}`);
  };

  if (!apps) {
    return <Loading />;
  }

  return (
    <>
      <div className="mb-4 flex justify-between items-center">
        <h2 className="font-bold text-2xl font-headline dark:text-white">
          Connected apps
        </h2>
        <a
          className="inline-flex bg-purple-700 cursor-pointer dark:text-neutral-200 duration-150 focus-visible:ring-2 focus-visible:ring-offset-2 focus:outline-none font-medium items-center justify-center px-3 md:px-6 py-2 md:py-3 rounded-lg shadow text-white transition {{if not .User}}opacity-50{{else}}hover:bg-purple-900{{end}} text-sm md:text-base"
          href="/apps/new"
        >
          <img
            src="public/images/plus.svg"
            width="24"
            height="24"
            className="mr-2 text-white"
          />
          Connect app
        </a>
      </div>

      <div className="rounded-lg border border-gray-200 dark:border-white/10 overflow-hidden">
        <table className="table-fixed w-full text-sm text-left">
          <thead className="text-xs text-gray-900 uppercase bg-gray-50 dark:bg-surface-08dp dark:text-white rounded-t-lg">
            <tr>
              <th scope="col" className="px-6 py-3 w-full">
                Name
              </th>
              <th scope="col" className="px-6 py-3 w-40 hidden md:table-cell">
                Last used
              </th>
              <th scope="col" className="px-6 py-3 w-24"></th>
            </tr>
          </thead>
          <tbody className="divide-y dark:divide-white/10">
            {!apps.apps.length && (
              <tr className="bg-white dark:bg-surface-02dp">
                <td
                  colSpan={3}
                  className="px-6 py-16 text-center text-gray-500 dark:text-neutral-400"
                >
                  No apps connected yet.
                </td>
              </tr>
            )}

            {apps.apps.length && (
              <>
                {apps.apps.map((app) => (
                  <tr
                    onClick={() => handleRowClick(app.nostrPubkey)}
                    key={app.id}
                    className="bg-white dark:bg-surface-02dp cursor-pointer hover:bg-purple-50 dark:hover:bg-surface-16dp"
                  >
                    {/* onClick="window.location='/apps/{{.NostrPubkey}}'"*/}
                    <td className="px-6 py-4 text-gray-500 dark:text-white">
                      {app.name}
                    </td>
                    <td className="px-6 py-4 text-gray-500 dark:text-neutral-400 hidden md:table-cell">
                      {apps.eventsCounts && apps.eventsCounts[app.id] > 0
                        ? apps.lastEvent && apps.lastEvent[app.id]
                          ? apps.lastEvent[app.id].createdAt.toLocaleString()
                          : "-"
                        : "-"}
                    </td>
                    <td className="px-6 py-4 text-purple-700 dark:text-purple-400 text-right">
                      Details
                    </td>
                  </tr>
                ))}
              </>
            )}
          </tbody>
        </table>
      </div>
    </>
  );
}

export default Apps;
