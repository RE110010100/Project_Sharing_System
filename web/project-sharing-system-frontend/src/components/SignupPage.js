// src/components/SignupPage.js

import React from 'react';
import { useAuth0 } from '@auth0/auth0-react';

function SignupPage() {
  const { loginWithRedirect } = useAuth0();

  return (
    <div className="signup-page">
      <h2>Sign Up</h2>
      <p>Click the button below to sign up with Auth0:</p>
      <button onClick={() => loginWithRedirect({ screen_hint: 'signup' })}>Sign Up with Auth0</button>
    </div>
  );
}

export default SignupPage;
