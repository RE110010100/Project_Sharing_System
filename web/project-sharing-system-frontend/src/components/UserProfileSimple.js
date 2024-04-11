import React from 'react';
import { Link, useLocation, useParams } from 'react-router-dom'; // Import Link from react-router-dom

function UserProfileSimple() {

  const location = useLocation();
  //const username = location.state.userID
  const {userID} = useParams();

  return (
    <div className="user-profile">
          <h2>User Profile</h2>
          <p>User: {userID}</p>
          {/* Button to create project */}
          <Link to={`/createProject/${userID}`}>
            <button>Create Project</button>
          </Link>
          {/* Button to list projects */}
          <Link to={{ pathname: `/ListProjects/${userID}`, state: { userId : userID } }}>
            <button>List Projects</button>
          </Link>
          {/* Button to update user information */}
          <Link to="/updateUserInfo">
            <button>Update User Information</button>
          </Link>
          <Link to={`/profile/${userID}`}>
          <button>
            Log Out
          </button>
          </Link>
          <Link to={`/userNotifications/${userID}`} state={{userId: userID}}>
          <button>User Notifications</button>
          </Link>
    </div>
  );
}

export default UserProfileSimple;
