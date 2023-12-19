function NotFound() {
  return(
    <div className="bg-white rounded-md shadow px-4 lg:p-8 dark:bg-surface-02dp">
      <h3 className="font-bold text-2xl font-headline mb-4 dark:text-white">404 Page Not Found</h3>
      <p className="leading-relaxed dark:text-neutral-400">
        The page you are looking for does not exist.
      </p>
    </div>
  )
}

export default NotFound;