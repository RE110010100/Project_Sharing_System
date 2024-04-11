import React from 'react';
import { Route, Routes, BrowserRouter } from 'react-router-dom';
import { Auth0Provider } from '@auth0/auth0-react';
import LandingPage from './components/LandingPage';
import LoginPage from './components/LoginPage'; // Import your login page component
import SignupPage from './components/SignupPage'; // Import your signup page component
import UserProfile from './components/UserProfile';
import CreateProject from './components/CreateProject'
import UploadProject from './components/UploadProject'
import ListProjects from './components/ListProjects';
import ListFiles from './components/ListFiles';
import UpdateFile from './components/UpdateFile';
import UpdateProject from './components/UpdateProject'
import UserNotifications from './components/UserNotifications'
import LoginPageSimple from './components/LoginPageSimple';
import UserProfileSimple from './components/UserProfileSimple';
import CreateProjectSimple from './components/CreateProjectSimple';
import UploadProjectSimple from './components/UploadProjectSimple';
import ListFilesSimple from './components/ListFilesSimple';
import ListProjectsSimple from './components/ListProjectsSimple';
import UpdateFileSimple from './components/UpdateFileSimple';
import UpdateProjectSimple from './components/UpdateProjectSimple';
import UserNotificationsSimple from './components/UserNotificationsSimple'
//import NotFoundPage from './components/NotFoundPage'; // Import a not found page component if needed

function App() {
  return (
    <BrowserRouter>
        <Routes>
          <Route path="/" element={<LandingPage />} />
          <Route path="/login" element={<LoginPageSimple />} />
          <Route path="/profile/:userID" element={<UserProfileSimple />} />
          <Route path="/createProject/:userID" element={<CreateProjectSimple />} />
          <Route path="/uploadProject/:userID" element={<UploadProjectSimple />} />
          <Route path="/ListProjects/:userID" element={<ListProjectsSimple/>} />
          <Route path="/ListFiles/:userID" element={<ListFilesSimple/>} />
          <Route path="/updateFile/:userID" element={<UpdateFileSimple/>} />
          <Route path="/updateProject/:userID" element={<UpdateProjectSimple/>} />
          <Route path="/userNotifications/:userID" element={<UserNotificationsSimple/>} />
        </Routes>
    </BrowserRouter>
  );
}

export default App; 





