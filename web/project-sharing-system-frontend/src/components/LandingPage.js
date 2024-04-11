// src/components/LandingPage.js

import React from 'react';
import { Link } from 'react-router-dom';

function LandingPage() {
  return (
    <div className="landing-page">
      <header>
        <h1>Welcome to Project Sharing</h1>
        <p>A platform to collaborate and share projects with others.</p>
      </header>
      <main>
        <section>
          <h2>About Us</h2>
          <p>Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed viverra est et efficitur consequat.</p>
        </section>
        <section>
          <h2>How It Works</h2>
          <p>Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed viverra est et efficitur consequat.</p>
        </section>
        <section>
          <h2>Get Started</h2>
          <p>Sign up or log in to start sharing your projects or collaborate with others!</p>
          <div className="buttons">
            <Link to="/signup"><button>Sign Up</button></Link>
            <Link to="/login"><button>Login</button></Link>
          </div>
        </section>
      </main>
      <footer>
        <p>&copy; 2024 Project Sharing</p>
      </footer>
    </div>
  );
}

export default LandingPage;
