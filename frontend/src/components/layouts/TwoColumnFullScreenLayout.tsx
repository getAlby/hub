import { Outlet } from "react-router-dom";

export default function TwoColumnFullScreenLayout() {
  return (
    <div className="w-full lg:grid lg:min-h-screen lg:grid-cols-2 items-stretch">
      <div className="hidden lg:flex flex-col bg-[#18181B] justify-end p-5">
        <p className="text-neutral-500">
          This isn't about nation-states anymore. This isn't about who adopts
          bitcoin first or who adopts cryptocurrencies first, because the
          internet is adopting cryptocurrencies, and the internet is the world's
          largest economy. It is the first transnational economy, and it needs a
          transnational currency.
        </p>
        <p className="text-white text-sm">Andreas M. Antonopoulos</p>
      </div>
      <div className="flex items-center justify-center py-12">
        <Outlet />
      </div>
    </div>
  );
}
