import { useState, useEffect } from 'react';

const New = () => {
  const [formData, setFormData] = useState({
    csrf: '',
    pubkey: '',
    returnTo: '',
    name: '',
    requestMethods: '',
    expiresAt: '',
    maxAmount: '',
    budgetRenewal: 'monthly',
    customRequestMethods: '',
    // requestMethodHelper: [],
    userEmail: '',
  });

  useEffect(() => {
    // Replace this with your actual API call
    const fetchData = async () => {
      const mockApiResponse = {
        // Mock data structure, replace with actual API response structure
        csrf: 'csrf-token',
        pubkey: 'public-key',
        returnTo: 'return-url',
        name: 'App Name',
        requestMethods: 'default-methods',
        expiresAt: 'expiry-date',
        maxAmount: '10000',
        budgetRenewal: 'monthly',
        customRequestMethods: '',
        // requestMethodHelper: [
        //   { key: 'method1', description: 'Method 1', icon: 'icon1', checked: true },
        //   { key: 'method2', description: 'Method 2', icon: 'icon2', checked: false },
        //   // Add more methods as needed
        // ],
        userEmail: 'user@example.com',
      };
      setFormData({ ...formData, ...mockApiResponse });
    };

    fetchData();
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  const handleInputChange = (event: React.ChangeEvent<HTMLInputElement>) => {
    setFormData({ ...formData, [event.target.name]: event.target.value });
  };

  const handleFormSubmit = async (event: React.FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    const response = await fetch('/apps', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(formData),
    });
    console.log(response)
    // Handle response
  };

  // Functionality for checkbox, expiry date, and other interactions...
  // You will need to implement these based on your specific needs

  return (
    <form onSubmit={handleFormSubmit} acceptCharset="UTF-8">
      <h2 className="font-bold text-2xl font-headline mb-4 dark:text-white">
        {formData.name ? `Connect to ${formData.name}` : 'Connect a new app'}
      </h2>
      <input type="hidden" name="_csrf" value={formData.csrf} />
      <input type="hidden" name="pubkey" value={formData.pubkey} />
      <input type="hidden" name="returnTo" value={formData.returnTo} />
      {!formData.name && (
        <>
          <label htmlFor="name" className="block font-medium text-gray-900 dark:text-white">
            Name
          </label>
          <input
            readOnly={!!formData.name}
            type="text"
            name="name"
            value={formData.name}
            id="name"
            required
            autoComplete="off"
            className="bg-gray-50 border border-gray-300 text-gray-900 focus:ring-purple-700 dark:focus:ring-purple-600 dark:ring-offset-gray-800 focus:ring-2 text-sm rounded-lg block w-full p-2.5 dark:bg-surface-00dp dark:border-gray-700 dark:placeholder-gray-400 dark:text-white"
            onChange={handleInputChange}
          />
          <p className="mt-1 mb-6 text-xs text-gray-500 dark:text-gray-400">
            Name of the app or purpose of the connection
          </p>
        </>
      )}
      {formData.name && <input type="hidden" name="name" value={formData.name} />}
      <input type="hidden" name="RequestMethods" value={formData.requestMethods} />
      <input type="hidden" name="ExpiresAt" value={formData.expiresAt} />
      <input type="hidden" name="MaxAmount" value={formData.maxAmount} />
      <input type="hidden" name="BudgetRenewal" value={formData.budgetRenewal} />

      {/* Additional form fields and interactions... */}
      {/* Implement the logic for editing permissions, setting budget amounts, and expiry dates as per your application's needs */}

      {formData.userEmail && (
        <p className="mt-8 pt-4 border-t border-gray-300 dark:border-gray-700 text-sm text-gray-500 dark:text-neutral-300 text-center">
          You're logged in as <span className="font-mono">{formData.userEmail}</span><br />
        </p>
      )}

      <div className="mt-6 flex flex-col sm:flex-row sm:justify-center">
        {!formData.pubkey && (
          <a
            href="/apps"
            className="inline-flex p-4 underline cursor-pointer duration-150 items-center justify-center text-gray-700 dark:text-neutral-300 w-full sm:w-[250px] order-last sm:order-first"
          >
            Cancel
          </a>
        )}
        <button
          type="submit"
          className="inline-flex w-full sm:w-[250px] bg-purple-700 cursor-pointer dark:text-neutral-200 duration-150 focus-visible:ring-2 focus-visible:ring-offset-2 focus:outline-none font-medium hover:bg-purple-900 items-center justify-center px-5 py-3 rounded-md shadow text-white transition"
        >
          {formData.pubkey ? 'Connect' : 'Next'}
        </button>
      </div>
    </form>
  );
};

export default New;