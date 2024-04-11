// src/components/LoginPage.js

import React from 'react';
import { useAuth0 } from '@auth0/auth0-react';

function LoginPage() {
  const { loginWithRedirect } = useAuth0();

  const handleLogin = () => {
    loginWithRedirect({
      returnTo: '/profile', // Specify the URL of the profile page
    });
  };

  return (
    <div className="login-page">
      <h2>Login</h2>
      <p>Click the button below to log in with Auth0:</p>
      <button onClick={handleLogin}>Login with Auth0</button>
    </div>
  );
}

export default LoginPage;

