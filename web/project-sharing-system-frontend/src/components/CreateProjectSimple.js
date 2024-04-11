import React, { useState } from 'react';
//import { Link } from 'react-router-dom';
import { useNavigate, useParams } from 'react-router-dom'; // Import Redirect from react-router-dom



function CreateProjectSimple() {
  const [title, setTitle] = useState('');
  const [description, setDescription] = useState('');
  const [isPublic, setIsPublic] = useState(false);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState(null);
  
  const navigate = useNavigate();

  const {userID} = useParams();

  const handleTitleChange = (event) => {
    setTitle(event.target.value);
  };

  const handleDescriptionChange = (event) => {
    setDescription(event.target.value);
  };

  const handleRadioChange = (event) => {
    setIsPublic(event.target.value === 'public');
  };

  const handleSubmit = async (event) => {
    event.preventDefault();

    setIsLoading(true);
    setError(null);

    const projectData = {
      title: title,
      description: description,
      is_public: isPublic,
      user_id: userID
    };

    try {
      const response = await fetch('http://127.0.0.1:8082/createProject', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json'
        },
        body: JSON.stringify(projectData)
      });

      if (!response.ok) {
        throw new Error('Network response was not ok');
      }

      const result = await response.json();
      console.log('Project created successfully:', result);

      setTitle('');
      setDescription('');
      setIsPublic(false);

      navigate(`/uploadProject/${userID}`, { state: { message: result.id, title: result.title} });

    } catch (error) {
      setError(error.message);
    } finally {
      setIsLoading(false);
    }
  };


  return (
    <div>
      <h2>Create New Project {userID}</h2>
      <form onSubmit={handleSubmit}>
        <div>
          <label htmlFor="title">Title:</label>
          <input
            type="text"
            id="title"
            value={title}
            onChange={handleTitleChange}
            required
          />
        </div>
        <div>
          <label htmlFor="description">Description:</label>
          <textarea
            id="description"
            value={description}
            onChange={handleDescriptionChange}
            required
          ></textarea>
        </div>
        <div>
          <label>
            <input
              type="radio"
              value="public"
              checked={isPublic === true}
              onChange={handleRadioChange}
            />
            Public
          </label>
          <label>
            <input
              type="radio"
              value="private"
              checked={isPublic === false}
              onChange={handleRadioChange}
            />
            Private
          </label>
        </div>
        <button type="submit" disabled={isLoading}>
          {isLoading ? 'Creating...' : 'Create Project'}
        </button>
        {error && <p style={{ color: 'red' }}>{error}</p>}
      </form>
    </div>
  );
}

export default CreateProjectSimple;



// minikube service frontend