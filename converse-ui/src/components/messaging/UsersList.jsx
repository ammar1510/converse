import React, { useState } from 'react';
import { generateAvatarUrl } from '../../utils/formatUtils';

const UsersList = ({ users, onSelectUser }) => {
  const [activeUser, setActiveUser] = useState(null);
  
  const handleSelectUser = (user) => {
    setActiveUser(user.id);
    onSelectUser(user);
  };

  return (
    <div className="users-list">
      <h3>All Users</h3>
      {users.length === 0 ? (
        <p className="no-users">No users found</p>
      ) : (
        <ul>
          {users.map(user => (
            <li 
              key={user.id} 
              className={`user-item ${activeUser === user.id ? 'active' : ''}`}
              onClick={() => handleSelectUser(user)}
            >
              <div className="user-avatar">
                <img 
                  src={user.avatar_url || generateAvatarUrl(user.username)} 
                  alt={user.username} 
                />
                <span className="online-indicator"></span>
              </div>
              <div className="user-content">
                <div className="user-name">{user.display_name || user.username}</div>
                <div className="user-email">{user.email}</div>
              </div>
            </li>
          ))}
        </ul>
      )}
    </div>
  );
};

export default UsersList; 