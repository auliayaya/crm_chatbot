import React from 'react';
import { Link } from 'react-router-dom';
import Button from '../components/common/Button'; // Assuming you have this

const NotFoundPage: React.FC = () => {
  return (
    <div className="flex flex-col items-center justify-center min-h-screen bg-gray-100 text-gray-800 p-4">
      <div className="text-center">
        <h1 className="text-6xl md:text-9xl font-bold text-primary-600 mb-4">404</h1>
        <h2 className="text-2xl md:text-4xl font-semibold mb-6">Page Not Found</h2>
        <p className="text-md md:text-lg text-gray-600 mb-8">
          Oops! The page you are looking for does not exist. It might have been moved or deleted.
        </p>
        <Button variant="default" asChild>
          <Link to="/dashboard">Go to Dashboard</Link>
        </Button>
      </div>
      <div className="mt-12 text-sm text-gray-500">
        If you believe this is an error, please contact support.
      </div>
    </div>
  );
};

export default NotFoundPage;