import React from 'react';
import { useAuth0 } from '@auth0/auth0-react';
import { Link } from 'react-router-dom'; // Import Link from react-router-dom

function UserProfile() {
  const { user, isAuthenticated, isLoading, logout } = useAuth0();
  
  if (isLoading) {
    return <div> Loading ... </div>
  }

  return (
    <div className="user-profile">
      {isAuthenticated ? (
        <>
          <h2>User Profile</h2>
          <p>Name: {user.name}</p>
          <p>Email: {user.email}</p>
          {/* Button to create project */}
          <Link to="/createProject">
            <button>Create Project</button>
          </Link>
          {/* Button to list projects */}
          <Link to={{ pathname: '/ListProjects', state: { userId : user.email } }}>
            <button>List Projects</button>
          </Link>
          {/* Button to update user information */}
          <Link to="/updateUserInfo">
            <button>Update User Information</button>
          </Link>
          <button onClick={() => logout({ logoutParams: { returnTo: window.location.origin } })}>
            Log Out
          </button>
          <Link to="/userNotifications" state={{userId: user.email}}>
          <button>User Notifications</button>
          </Link>
        </>
      ) : (
        <p>User not authenticated</p>
      )}
    </div>
  );
}

export default UserProfile;
