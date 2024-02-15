import nwcComboMark from "src/assets/images/nwc-combomark.svg";

export default function Container({ children }: React.PropsWithChildren) {
  return (
    <div className="flex flex-col items-center w-full">
      <div className="max-w-md flex flex-col items-center">
        <div className="mb-14">
          <img alt="NWC Logo" className="h-12 inline" src={nwcComboMark} />
        </div>
        {children}
      </div>
    </div>
  );
}
