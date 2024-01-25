import { Link } from "react-router-dom";

function NotFound() {
  return (
    <div className="flex justify-center">
      <div className="container max-w-screen-lg">
        <div className="bg-white rounded-md shadow my-4 px-4 lg:p-8 dark:bg-surface-02dp">
          <h3 className="font-bold text-2xl font-headline mb-4 dark:text-white">
            404 Page Not Found
          </h3>
          <p className="leading-relaxed dark:text-neutral-400 mb-4">
            The page you are looking for does not exist.
          </p>
          <Link to="/" className="text-purple-700 dark:text-purple-400">
            Return Home
          </Link>
        </div>
      </div>
    </div>
  );
}

export default NotFound;
