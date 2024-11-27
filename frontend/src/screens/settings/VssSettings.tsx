import Loading from "src/components/Loading";
import SettingsHeader from "src/components/SettingsHeader";
import { useInfo } from "src/hooks/useInfo";

export default function VssSettings() {
  const { data: info } = useInfo();

  if (!info) {
    return <Loading />;
  }

  return (
    <>
      <SettingsHeader title="VSS" description="Versioned Storage Service" />
      <p>
        Versioned Storage Service (VSS) provides a secure, encrypted server-side
        storage of essential lightning and onchain data, which allows you to
        recover your lightning data with your recovery phrase alone, without
        having to close your channels.
      </p>
      {!info.albyAccountConnected && (
        <p>An Alby Account is required to enable this feature.</p>
      )}

      <p>
        VSS enabled: <b>{info.ldkVssEnabled.toString()}</b>
      </p>
    </>
  );
}
