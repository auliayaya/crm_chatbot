export default function Footer() {
  const currentYear = new Date().getFullYear()

  return (
    <footer className="bg-white border-t border-gray-200">
      <div className="max-w-7xl mx-auto py-4 px-4 sm:px-6 md:px-8">
        <div className="flex flex-col md:flex-row justify-between items-center">
          <p className="text-gray-500 text-sm">
            &copy; {currentYear} CRM Chatbot. All rights reserved.
          </p>
          <div className="flex space-x-6 mt-2 md:mt-0">
            <a
              href="#terms"
              className="text-gray-500 text-sm hover:text-gray-700"
            >
              Terms of Service
            </a>
            <a
              href="#privacy"
              className="text-gray-500 text-sm hover:text-gray-700"
            >
              Privacy Policy
            </a>
            <a
              href="#help"
              className="text-gray-500 text-sm hover:text-gray-700"
            >
              Help Center
            </a>
          </div>
        </div>
      </div>
    </footer>
  )
}
