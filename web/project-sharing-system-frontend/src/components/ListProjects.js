import React, { useState, useEffect } from 'react';
//import { useLocation } from 'react-router-dom';
import { useAuth0 } from '@auth0/auth0-react';
import { Link } from 'react-router-dom';

const ProjectList = () => {

  const { user } = useAuth0();


  const userId = user.email // Assuming the user ID is part of the URL path
  const [projects, setProjects] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);

  

  useEffect(() => {
    const fetchProjects = async () => {
      try {
        // Fetch projects data from the API
        const response = await fetch(`http://localhost:8083/listUserProjectWithString?userId=${userId}`);
        if (!response.ok) {
          throw new Error('Failed to fetch projects');
        }
        const data = await response.json();
        setProjects(data);
        setLoading(false);

    
      } catch (error) {
        setError(error.message);
        setLoading(false);
      }
    };

    fetchProjects();
  });

  if (loading) {
    return <div>Loading...</div>;
  }

  if (error) {
    return <div>Error: {error}</div>;
  }

  return (
    <div>
      <h2>Projects</h2>
      <ul>
        {projects.map(project => (
          <li key={project.id}>
            {project.title} - {project.description} - {project.id}
            <Link to='/ListFiles' state={{projID: project.id, projTitle: project.title}}>
            <button> go to project </button>
            </Link>
          </li>
        ))}
      </ul>
    </div>
  );
};

export default ProjectList;
