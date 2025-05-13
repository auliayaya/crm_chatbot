import React from 'react';

interface FormErrorProps {
  message?: string;
}

const FormError: React.FC<FormErrorProps> = ({ message }) => {
  if (!message) return null;

  return (
    <div className="p-3 my-2 bg-red-50 border border-red-200 text-sm text-red-700 rounded-md">
      {message}
    </div>
  );
};

export default FormError;