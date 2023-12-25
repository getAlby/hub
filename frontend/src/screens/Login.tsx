import { useNavigate } from "react-router-dom";
import { useUser } from "../context/UserContext";
import { useEffect } from "react";
import nwcLogo from "../assets/nwc-logo.svg"

function Login() {
  const { info } = useUser()
  const navigate = useNavigate()

  useEffect(() => {
    if (info?.user) {
      navigate('/');
    }
  }, [navigate, info?.user]);

  return(
    <div className="text-center">
      <img
        alt="Nostr Wallet Connect logo"
        className="mx-auto my-4"
        width="128"
        height="120"
        src={nwcLogo}
      />

      <h1 className="font-headline text-3xl sm:text-4xl mb-2 dark:text-white">
        Nostr Wallet Connect
      </h1>

      <h2 className="text-md text-gray-700 dark:text-neutral-300">
        Securely connect your Alby Account to Nostr clients and applications.
      </h2>

      <p className="my-8">
        <a className=" inline-flex cursor-pointer items-center justify-center rounded-md transition-all px-10 py-4 text-black bg-gradient-to-r from-[#ffde6e] to-[#f8c455]" href="/alby/auth">
        {/* style="background: linear-gradient(180deg, #ffde6e 63.72%, #f8c455 95.24%);" */}
          <img
            src="/public/images/alby-head.svg"
            width="400"
            height="400"
            className="w-[24px] mr-2"
          />
          Log in with Alby Account
        </a>
      </p>
    
      <p>
        <a href="/about" className="text-purple-700 dark:text-purple-400">How does it work?</a>
      </p>
    </div>
  )
}

export default Login;