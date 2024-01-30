import { Link } from "react-router-dom";

export function Welcome() {
  return (
    <div className="flex flex-col justify-center items-center gap-10">
      <p>Welcome to NWC!</p>
      <p>NWC is ... and allows you to ...</p>
      <Link to="/about" className="text-purple-500">
        Learn more
      </Link>

      <Link to="/setup" className="text-green-500">
        Get Started
      </Link>
    </div>
  );
}
