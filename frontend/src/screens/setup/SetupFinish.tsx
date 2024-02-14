import { useEffect, useState } from "react";
import { Link, useNavigate } from "react-router-dom";
import Container from "src/components/Container";
import { useInfo } from "src/hooks/useInfo";
import useSetupStore from "src/state/SetupStore";
import { useCSRF } from "src/hooks/useCSRF";
import { handleRequestError } from "src/utils/handleRequestError";
import { request } from "src/utils/request";
import Loading from "src/components/Loading";

export function SetupFinish() {
  const navigate = useNavigate();
  const { nodeInfo } = useSetupStore();
  const [isLoading, setIsLoading] = useState(true);

  const { mutate: refetchInfo } = useInfo();
  const { data: csrf, isLoading: csrfLoading } = useCSRF();

  useEffect(() => {
    if (csrfLoading) {
      return;
    }

    const postNodeInfo = async () => {
      try {
        if (!csrf) {
          throw new Error("info not loaded");
        }
        console.log(nodeInfo);
        await request("/api/setup", {
          method: "POST",
          headers: {
            "X-CSRF-Token": csrf,
            "Content-Type": "application/json",
          },
          body: JSON.stringify({
            ...nodeInfo,
          }),
        });
      } catch (error) {
        handleRequestError("Failed to connect", error);
      } finally {
        setIsLoading(false);
      }
    };

    postNodeInfo();
  }, [csrf, csrfLoading, nodeInfo, refetchInfo, navigate]);

  if (isLoading) {
    return <Loading />;
  }

  return (
    <Container>
      <h1 className="font-semibold text-2xl font-headline mb-2 dark:text-white">
        Setup Finished
      </h1>
      <p className="text-center text-md leading-relaxed dark:text-neutral-400 mb-14">
        Happy Zapping!
      </p>

      <div className="w-full flex flex-row justify-between">
        <Link
          to="/about"
          className="flex-row w-1/2 first:mr-2 last:ml-2 px-0 py-2 bg-white text-gray-700 dark:bg-surface-02dp dark:text-neutral-200 dark:border-neutral-800 hover:bg-gray-50 dark:hover:bg-surface-16dp cursor-pointer inline-flex justify-center items-center font-medium bg-origin-border shadow rounded-md focus:outline-none focus-visible:ring-2 focus-visible:ring-offset-2 focus-visible:ring-purple-700 transition duration-150"
        >
          Learn more
        </Link>
        <Link
          to="/apps"
          className="flex-row w-1/2 first:mr-2 last:ml-2 px-0 py-2 bg-purple-700 border-2 border-transparent text-white hover:bg-purple-800 cursor-pointer inline-flex justify-center items-center font-medium bg-origin-border shadow rounded-md focus:outline-none focus-visible:ring-2 focus-visible:ring-offset-2 focus-visible:ring-purple-500 transition duration-150"
        >
          See Apps
        </Link>
      </div>
    </Container>
  );
}
