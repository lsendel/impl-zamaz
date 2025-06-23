import React from 'react';

interface User {
  id: string;
  username: string;
  email: string;
  roles: string[];
}

interface UserInfoProps {
  user: User | null;
}

const UserInfo: React.FC<UserInfoProps> = ({ user }) => {
  if (!user) {
    return (
      <div className="card">
        <h2>👤 User Information</h2>
        <p>Loading user information...</p>
      </div>
    );
  }

  return (
    <div className="card">
      <h2>👤 User Information</h2>
      <div className="user-info">
        <p><strong>Username:</strong> {user.username}</p>
        <p><strong>Email:</strong> {user.email}</p>
        <p><strong>Roles:</strong> {user.roles.join(', ')}</p>
      </div>
    </div>
  );
};

export default UserInfo;