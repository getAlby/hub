import { useEffect } from "react";
import { useNavigate } from "react-router-dom";
import Container from "src/components/Container";
import { useInfo } from "src/hooks/useInfo";
import useSetupStore from "src/state/SetupStore";
import { useCSRF } from "src/hooks/useCSRF";
import { handleRequestError } from "src/utils/handleRequestError";
import { request } from "src/utils/request";
import Loading from "src/components/Loading";
import { NodeInfo } from "src/types";

let hasFetched = false;
export function SetupFinish() {
  const navigate = useNavigate();
  const { nodeInfo, unlockPassword } = useSetupStore();

  const { mutate: refetchInfo } = useInfo();
  const { data: csrf } = useCSRF();

  useEffect(() => {
    // ensure setup call is only called once
    if (!csrf || hasFetched) {
      return;
    }
    hasFetched = true;

    (async () => {
      if (await finishSetup(csrf, nodeInfo, unlockPassword)) {
        await refetchInfo();
        navigate("/");
      }
    })();
  }, [csrf, nodeInfo, refetchInfo, navigate, unlockPassword]);

  return (
    <Container>
      <h1 className="font-semibold text-lg font-headline mt-16 mb-8 dark:text-white">
        Connecting...
      </h1>

      <Loading />
    </Container>
  );
}

const finishSetup = async (
  csrf: string,
  nodeInfo: NodeInfo,
  unlockPassword: string
): Promise<boolean> => {
  try {
    if (!csrf) {
      throw new Error("info not loaded");
    }
    await request("/api/setup", {
      method: "POST",
      headers: {
        "X-CSRF-Token": csrf,
        "Content-Type": "application/json",
      },
      body: JSON.stringify({
        ...nodeInfo,
        unlockPassword,
      }),
    });
    await request("/api/start", {
      method: "POST",
      headers: {
        "X-CSRF-Token": csrf,
        "Content-Type": "application/json",
      },
      body: JSON.stringify({
        unlockPassword,
      }),
    });
    return true;
  } catch (error) {
    handleRequestError("Failed to connect", error);
    return false;
  }
};
