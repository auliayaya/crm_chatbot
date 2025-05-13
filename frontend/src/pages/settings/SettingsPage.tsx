import React, { useState } from 'react';
import DashboardLayout from '../../layouts/DashboardLayout';
import Card from '../../components/common/Card';
import Button from '../../components/common/Button';
import Input from '../../components/common/Input'; // Assuming you have this
// import Select from '../../components/common/Select'; // If needed for settings

// Example: Define a type for user profile data if you fetch/update it
interface UserProfile {
  firstName: string;
  lastName: string;
  email: string;
  // Add other profile fields as needed
}

// Example: Define a type for application preferences
interface AppPreferences {
  theme: 'light' | 'dark' | 'system';
  notifications: {
    email: boolean;
    inApp: boolean;
  };
}

export default function SettingsPage() {
  // Example state for profile - you would fetch this
  const [profile, setProfile] = useState<UserProfile>({
    firstName: 'John', // Replace with actual data or fetch
    lastName: 'Doe',
    email: 'john.doe@example.com',
  });

  // Example state for preferences - you would fetch this
  const [preferences, setPreferences] = useState<AppPreferences>({
    theme: 'light',
    notifications: {
      email: true,
      inApp: true,
    },
  });

  const [isSavingProfile, setIsSavingProfile] = useState(false);
  const [isSavingPreferences, setIsSavingPreferences] = useState(false);

  const handleProfileChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const { name, value } = e.target;
    setProfile(prev => ({ ...prev, [name]: value }));
  };

  const handlePreferenceChange = (e: React.ChangeEvent<HTMLInputElement | HTMLSelectElement>) => {
    const { name, value, type } = e.target;
    if (name.startsWith('notifications.')) {
      const key = name.split('.')[1] as keyof AppPreferences['notifications'];
      setPreferences(prev => ({
        ...prev,
        notifications: {
          ...prev.notifications,
          [key]: type === 'checkbox' ? (e.target as HTMLInputElement).checked : value,
        },
      }));
    } else {
      setPreferences(prev => ({ ...prev, [name]: value as AppPreferences['theme'] }));
    }
  };

  const handleSaveProfile = async (e: React.FormEvent) => {
    e.preventDefault();
    setIsSavingProfile(true);
    console.log('Saving profile:', profile);
    // TODO: Implement API call to save profile
    await new Promise(resolve => setTimeout(resolve, 1000)); // Simulate API call
    setIsSavingProfile(false);
    // TODO: Add success/error toast
  };

  const handleSavePreferences = async (e: React.FormEvent) => {
    e.preventDefault();
    setIsSavingPreferences(true);
    console.log('Saving preferences:', preferences);
    // TODO: Implement API call to save preferences
    await new Promise(resolve => setTimeout(resolve, 1000)); // Simulate API call
    setIsSavingPreferences(false);
    // TODO: Add success/error toast
  };

  return (
    <DashboardLayout>
      <h1 className="text-2xl font-semibold text-gray-900 mb-6">Settings</h1>

      <div className="space-y-8">
        {/* Profile Settings Section */}
        <Card>
          <form onSubmit={handleSaveProfile}>
            <div className="p-6">
              <h2 className="text-xl font-medium text-gray-800 mb-4">Profile</h2>
              <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                <Input
                  label="First Name"
                  name="firstName"
                  value={profile.firstName}
                  onChange={handleProfileChange}
                  disabled={isSavingProfile}
                />
                <Input
                  label="Last Name"
                  name="lastName"
                  value={profile.lastName}
                  onChange={handleProfileChange}
                  disabled={isSavingProfile}
                />
              </div>
              <div className="mt-6">
                <Input
                  label="Email Address"
                  name="email"
                  type="email"
                  value={profile.email}
                  onChange={handleProfileChange}
                  disabled // Usually email is not changed directly or requires verification
                />
              </div>
            </div>
            <div className="bg-gray-50 px-6 py-3 text-right rounded-b-lg">
              <Button type="submit" variant="default" disabled={isSavingProfile}>
                {isSavingProfile ? 'Saving...' : 'Save Profile'}
              </Button>
            </div>
          </form>
        </Card>

        {/* Application Preferences Section */}
        <Card>
          <form onSubmit={handleSavePreferences}>
            <div className="p-6">
              <h2 className="text-xl font-medium text-gray-800 mb-4">Preferences</h2>
              {/* Theme Setting - Example with Select if you have one */}
              {/* <div className="mb-4">
                <Select
                  label="Theme"
                  name="theme"
                  value={preferences.theme}
                  onChange={handlePreferenceChange}
                  disabled={isSavingPreferences}
                >
                  <option value="light">Light</option>
                  <option value="dark">Dark</option>
                  <option value="system">System Default</option>
                </Select>
              </div> */}
              <div className="mb-4">
                <label className="block text-sm font-medium text-gray-700 mb-1">Theme</label>
                <select
                  name="theme"
                  value={preferences.theme}
                  onChange={handlePreferenceChange}
                  disabled={isSavingPreferences}
                  className="mt-1 block w-full pl-3 pr-10 py-2 text-base border-gray-300 focus:outline-none focus:ring-primary-500 focus:border-primary-500 sm:text-sm rounded-md"
                >
                  <option value="light">Light</option>
                  <option value="dark">Dark</option>
                  <option value="system">System Default</option>
                </select>
              </div>

              <div>
                <h3 className="text-md font-medium text-gray-700 mb-2">Notifications</h3>
                <div className="space-y-2">
                  <label className="flex items-center">
                    <input
                      type="checkbox"
                      name="notifications.email"
                      checked={preferences.notifications.email}
                      onChange={handlePreferenceChange}
                      disabled={isSavingPreferences}
                      className="h-4 w-4 text-primary-600 border-gray-300 rounded focus:ring-primary-500"
                    />
                    <span className="ml-2 text-sm text-gray-700">Email Notifications</span>
                  </label>
                  <label className="flex items-center">
                    <input
                      type="checkbox"
                      name="notifications.inApp"
                      checked={preferences.notifications.inApp}
                      onChange={handlePreferenceChange}
                      disabled={isSavingPreferences}
                      className="h-4 w-4 text-primary-600 border-gray-300 rounded focus:ring-primary-500"
                    />
                    <span className="ml-2 text-sm text-gray-700">In-App Notifications</span>
                  </label>
                </div>
              </div>
            </div>
            <div className="bg-gray-50 px-6 py-3 text-right rounded-b-lg">
              <Button type="submit" variant="default" disabled={isSavingPreferences}>
                {isSavingPreferences ? 'Saving...' : 'Save Preferences'}
              </Button>
            </div>
          </form>
        </Card>

        {/* Security Settings Section - Placeholder */}
        <Card>
          <div className="p-6">
            <h2 className="text-xl font-medium text-gray-800 mb-4">Security</h2>
            <p className="text-sm text-gray-600 mb-4">Manage your account security settings, like changing your password or enabling two-factor authentication.</p>
            <Button variant="outline" onClick={() => console.log('Change password clicked')}>
              Change Password
            </Button>
            {/* TODO: Add 2FA settings if applicable */}
          </div>
        </Card>
      </div>
    </DashboardLayout>
  );
}