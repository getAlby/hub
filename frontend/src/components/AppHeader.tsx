type Props = {
  headerLeft?: React.ReactNode;
  headerRight?: React.ReactNode;
  children: React.ReactNode;
};

function AppHeader({ children, headerLeft, headerRight }: Props) {
  return (
    <div className="bg-white py-1 border-b border-gray-200 dark:bg-surface-01dp dark:border-neutral-700">
      <div className="flex justify-between items-center container max-w-screen-lg mx-auto">
        <div className="w-8 h-8 mr-3">{headerLeft}</div>
        <h1 className="text-lg font-medium dark:text-white overflow-hidden">
          {children}
        </h1>
        <div className="w-8 h-8 ml-3">{headerRight}</div>
      </div>
    </div>
  );
}

export default AppHeader;
