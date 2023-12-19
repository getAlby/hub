/* eslint-disable react-refresh/only-export-components */
import { useEffect, useState } from "react";
import { useNavigate } from 'react-router-dom';

import Loading from '../../components/loading';
import { App, NostrEvent, ListAppsResponse } from "../../types";

function Connections() {
  const [apps, setApps] = useState<App[] | null>(null);
  const [eventsCounts, setEventsCounts] = useState<Record<App["id"], number> | null>(null);
  const [lastEvents, setLastEvents] = useState<Record<App["id"], NostrEvent> | null>(null);
  const navigate = useNavigate();

  const handleRowClick = (appId: App["nostrPubkey"]) => {
    navigate(`/apps/${appId}`);
  };

  useEffect(() => {
    const fetchAppsData = async () => {
      try {
        const response = await fetch("/api/apps");
        const data: ListAppsResponse = await response.json();
        setApps(data.apps);
        setEventsCounts(data.eventsCounts);
        setLastEvents(data.lastEvent);
      } catch (error) {
        // TODO: Show error page
        navigate("/apps?q=notfound");
        console.error('Error fetching app data:', error);
      }
    };

    fetchAppsData();
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  return(
    <>
      <div className="mb-4 flex justify-between items-center">
        <h2 className="font-bold text-2xl font-headline dark:text-white">Connected apps</h2>
        <a className="inline-flex bg-purple-700 cursor-pointer dark:text-neutral-200 duration-150 focus-visible:ring-2 focus-visible:ring-offset-2 focus:outline-none font-medium items-center justify-center px-3 md:px-6 py-2 md:py-3 rounded-lg shadow text-white transition {{if not .User}}opacity-50{{else}}hover:bg-purple-900{{end}} text-sm md:text-base" href="{{if not .User}}javascript:void(0);{{else}}/apps/new{{end}}">
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
              <th scope="col" className="px-6 py-3 w-full">Name</th>
              <th scope="col" className="px-6 py-3 w-40 hidden md:table-cell">Last used</th>
              <th scope="col" className="px-6 py-3 w-24"></th>
            </tr>
          </thead>
          <tbody className="divide-y dark:divide-white/10">
            {!apps && <tr className="bg-white dark:bg-surface-02dp">
              <td colSpan={3} className="px-6 py-16 text-gray-500 dark:text-neutral-400">
                <div className="flex w-full justify-center">
                  <Loading />
                </div>
              </td>
            </tr>}
            
            {apps && !apps.length && <tr className="bg-white dark:bg-surface-02dp">
              <td colSpan={3} className="px-6 py-16 text-center text-gray-500 dark:text-neutral-400">
                No apps connected yet.
              </td>
            </tr>}
            
            {apps && apps?.length && (<>
              {apps.map((app) => (
                <tr onClick={() => handleRowClick(app.nostrPubkey)} key={app.id} className="bg-white dark:bg-surface-02dp cursor-pointer hover:bg-purple-50 dark:hover:bg-surface-16dp">
                  {/* onClick="window.location='/apps/{{.NostrPubkey}}'"*/}
                  <td className="px-6 py-4 text-gray-500 dark:text-white">
                    {app.name}
                  </td>
                  <td className="px-6 py-4 text-gray-500 dark:text-neutral-400 hidden md:table-cell">
                    {eventsCounts && eventsCounts[app.id] > 0 ? (
                      lastEvents && lastEvents[app.id] ? (
                        lastEvents[app.id].createdAt.toLocaleString()
                      ) : (
                        '-'
                      )
                    ) : (
                      '-'
                    )}
                  </td>
                  <td className="px-6 py-4 text-purple-700 dark:text-purple-400 text-right">
                    Details
                  </td>
                </tr>
              ))}
            </>)}
          </tbody>
        </table>
      </div>
    </>
  )
}

export default Connections;