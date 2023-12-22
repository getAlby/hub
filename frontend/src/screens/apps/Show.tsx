import { useState, useEffect } from 'react';
import { ShowAppResponse } from '../../types';
import Loading from '../../components/Loading';
import { useNavigate, useParams } from 'react-router-dom';
import { useUser } from '../../context/UserContext';

function Show() {
  const { info } = useUser();
  const { pubkey } = useParams() as { pubkey: string };
  const navigate = useNavigate();
  const [appData, setAppData] = useState<ShowAppResponse | null>(null);

  const handleDelete = async (event: React.FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    if (!appData || !info) return
    try {
      // Here you'd handle form submission. For example:
      await fetch(`/api/apps/${appData.app.nostrPubkey}`, {
        method: 'DELETE',
        headers: {
          'Content-Type': 'application/json',
          'X-CSRF-Token': info.csrf,
        }
      })
      navigate("/apps?q=appdeleted");
    } catch (error) {
      console.error('Error deleting app:', error);
    }
  };

  useEffect(() => {
    const fetchAppData = async () => {
      try {
        const response = await fetch(`/api/apps/${pubkey}`);
        const data: ShowAppResponse = await response.json();
        setAppData(data);
      } catch (error) {
        console.error('Error fetching app data:', error);
        // TODO: Show error page
        navigate("/apps?q=notfound");
      }
    };

    fetchAppData();
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  return (
    <div>
      <h2 className="font-bold text-2xl font-headline dark:text-white">{appData ? appData.app.name : "Fetching app..."}</h2>
      <p className="text-gray-600 dark:text-neutral-400 text-sm">{appData ? appData.app.description : ""}</p>
      <a className="ml-1 mt-1 mb-4 block dark:text-white text-xs" href="/apps">
        &slarr;
        Back to overview
      </a>

      <div className="bg-white rounded-md shadow p-4 lg:p-8 dark:bg-surface-02dp">
        {!appData && <div>
          <Loading />
        </div>}
        {appData &&
          <>
            <div className="divide-y divide-gray-200 dark:divide-white/10 dark:bg-surface-02dp">
              <div className="pb-4">
                <table>
                  <tr>
                    <td className="align-top w-32 font-medium dark:text-white">Public Key</td>
                    <td className="text-gray-600 dark:text-neutral-400 break-all">{appData.app.nostrPubkey}</td>
                  </tr>
                  <tr>
                    <td className="align-top font-medium dark:text-white">Last used</td>
                    <td className="text-gray-600 dark:text-neutral-400">{appData.eventsCount && appData.lastEvent ? appData.lastEvent.createdAt : 'never'}</td>
                  </tr>
                  <tr>
                    <td className="align-top font-medium dark:text-white">Expires at</td>
                    <td className="text-gray-600 dark:text-neutral-400">{appData.expiresAtFormatted || 'never'}</td>
                  </tr>
                </table>
              </div>

              <div className="py-4">
                <h3 className="text-xl font-headline dark:text-white">Permissions</h3>
                <ul className="mt-2 text-gray-600 dark:text-neutral-400">
                  {appData.requestMethods.map((method, index) => (
                    <li key={index} className="mb-2 relative pl-6">
                      <span className="absolute left-0 text-green-500">✓</span>
                      {method}
                    </li>
                  ))}
                </ul>
                {appData.paySpecificPermission && appData.paySpecificPermission.maxAmount > 0 && (
                  <div className="pl-6">
                    <table className="text-gray-600 dark:text-neutral-400">
                      <tr>
                        <td className="font-medium">Budget</td>
                        <td>{appData.paySpecificPermission.maxAmount} sats ({appData.budgetUsage} sats used)</td>
                      </tr>
                      <tr>
                        <td className="font-medium pr-3">Renews in</td>
                        <td>{appData.renewsIn} (set to {appData.paySpecificPermission.budgetRenewal})</td>
                      </tr>
                    </table>
                  </div>
                )}
              </div>

              <div className="pt-4">
                <h3 className="text-xl font-headline mb-2 dark:text-white">⚠️ Danger zone</h3>
                <p className="text-gray-600 dark:text-neutral-400 mb-4">
                  This will revoke the permission and will no longer allow calls from this public key.
                </p>
              </div>
            </div>

            <form method="post" onSubmit={handleDelete}>
              <input type="hidden" name="_csrf" value={appData.csrf} />
              <button type="submit"
                className="inline-flex bg-white border border-red-400 cursor-pointer dark:bg-surface-02dp dark:hover:bg-surface-16dp duration-150 focus-visible:ring-2 focus-visible:ring-offset-2 focus:outline-none font-medium hover:bg-gray-50 items-center justify-center px-5 py-3 rounded-md shadow text-gray-700 dark:text-neutral-300 transition w-full sm:w-[250px] sm:mr-8 mt-8 sm:mt-0 order-last sm:order-first">Disconnect</button>
            </form>
          </>
        }
      </div>
    </div>
  );
}

export default Show;