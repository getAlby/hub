import { Link, useNavigate } from "react-router-dom";

import { useApps } from "src/hooks/useApps";
import { PlusIcon } from "src/components/icons/PlusIcon";
import Loading from "src/components/Loading";

function AppsList() {
  const { data: apps } = useApps();

  const navigate = useNavigate();

  if (!apps) {
    return <Loading />;
  }

  return (
    <>
      <div className="mb-4 flex justify-between items-center">
        <h2 className="font-bold text-2xl font-headline dark:text-white">
          Connected apps
        </h2>
        <Link
          className="inline-flex bg-purple-700 cursor-pointer dark:text-neutral-200 duration-150 focus-visible:ring-2 focus-visible:ring-offset-2 focus:outline-none font-medium items-center justify-center px-3 md:px-6 py-2 md:py-3 rounded-lg shadow text-white transition hover:bg-purple-900 text-sm md:text-base"
          to="/apps/new"
        >
          <PlusIcon className="mr-2 text-white w-6 h-6" />
          Connect app
        </Link>
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
            {!apps.length && (
              <tr className="bg-white dark:bg-surface-02dp">
                <td
                  colSpan={3}
                  className="px-6 py-16 text-center text-gray-500 dark:text-neutral-400"
                >
                  No apps connected yet.
                </td>
              </tr>
            )}

            {apps.map((app, index) => (
              <tr
                key={index}
                onClick={() => navigate(`/apps/${app.nostrPubkey}`)}
                className="bg-white dark:bg-surface-02dp cursor-pointer hover:bg-purple-50 dark:hover:bg-surface-16dp"
              >
                <td className="px-6 py-4 text-gray-500 dark:text-white">
                  {app.name}
                </td>
                <td className="px-6 py-4 text-gray-500 dark:text-neutral-400 hidden md:table-cell">
                  {app.lastEventAt
                    ? new Date(app.lastEventAt).toLocaleDateString()
                    : "-"}
                </td>
                <td className="px-6 py-4 text-purple-700 dark:text-purple-400 text-right">
                  <Link to={`/apps/${app.nostrPubkey}`}>Details</Link>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </>
  );
}

export default AppsList;
