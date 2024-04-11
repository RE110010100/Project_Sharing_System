import React, { useState, useEffect } from 'react';
import { useLocation } from 'react-router-dom';
import { Link, useNavigate, useParams } from 'react-router-dom';

const ListFilesSimple = () => {

const {userID} = useParams();

const location = useLocation();
//const projID  = state.projID;
//const {projID} = state;

//console.log(projID)

const navigate = useNavigate();

  const [files, setFiles] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [flag, setFlag] = useState(false);
  //const [projectID, setProjectID] = useState(null);

  //setProjectID(projectID)

  const [isLoading, setIsLoading] = useState(false);
  const [error1, setError1] = useState(null);
  const [error2, setError2] = useState(null);

  const handleClick = () => {
    setIsLoading(true);
  
    // Second API call
    fetch(`http://localhost:8080/deleteProjectHandler?projectID=${location.state.projID}&UserID=${userID}&project_title=${location.state.projTitle}`, {
      method: 'GET',
    })
      .then(response => {
        if (!response.ok) {
          throw new Error('Failed to fetch data from another endpoint');
        }
        return response.json();
      })
      .then(anotherData => {
        // Process anotherData as needed
        console.log('Data from another endpoint:', anotherData);
  
        // First API call
        fetch(`http://localhost:8082/deleteProject?projectID=${location.state.projID}`, {
          method: 'POST',
        })
          .then(response => {
            if (!response.ok) {
              throw new Error('Failed to delete project record');
            }
            return response.json();
          })
          .then(data => {
            alert(data.message); // Display success message
            // Navigate or perform other actions after the second API call
            navigate(`/profile/${userID}`);
          })
          .catch(error => {
            setError1(error.message); // Display error message for the first API call
          });
      })
      .catch(error => {
        setError2(error.message); // Display error message for the second API call
      })
      .finally(() => {
        setIsLoading(false);
      });
  };
  

  useEffect(() => {
    const fetchFiles = async () => {
      try {
        // Fetch files data from the API
        const response = await fetch(`http://localhost:8080/listProjectFiles?project_id=${location.state.projID}`);
        if (!response.ok) {
          throw new Error('Failed to fetch files');
        }
        const data = await response.json();
        setFiles(data);
        setLoading(false);
        setFlag(true)

        console.log(location.state.projID)
      } catch (error) {
        setError(error.message);
        setLoading(false);
      }
    };

    fetchFiles();
  }, [location]);

  const handleButton = () => {
    // Replace 'your-folder-key' with the actual folder key
    //const folderKey = 'project2';

    // Make a GET request to the server endpoint
    fetch(`http://localhost:8080/downloadZip?UserID=${userID}&projectID=${location.state.projID}`)
      .then(response => {
        // Check if response is successful
        if (!response.ok) {
          throw new Error('Failed to download zip file');
        }
        console.log(userID)
        // Convert response to blob
        return response.blob();
      })
      .then(blob => {
        // Create a temporary URL for the blob
        const url = window.URL.createObjectURL(new Blob([blob]));
        // Create a temporary link element
        const link = document.createElement('a');
        // Set link properties

        const filename = location.state.projTitle + ".zip"
        link.href = url;
        link.setAttribute('download', filename);
        // Append link to the document body
        document.body.appendChild(link);
        // Click the link to trigger download
        link.click();
        // Remove the link from the document body
        document.body.removeChild(link);
      })
      .catch(error => {
        console.error('Error downloading zip file:', error);
        // Handle error here
      });
  };

  const handleUpdate = (fileId) => {
    console.log('Update file:', fileId);
    // Add your update logic here
  };

  const handleDelete = (fileId) => {
    console.log('Delete file:', fileId);
    // Add your delete logic here
  };

  if (loading) {
    return <div>Loading...</div>;
  }

  if (error) {
    return <div>Error: {error}</div>;
  }

  return (
    <div>
      <h2>Files List</h2>
      <ul>
        {flag && files.map(file => (
          <li key={file.id}>
            {file.FileName} - {file.FileSize} bytes
            <Link to="/updateFile" state={{state: file, project: location.state.projTitle}}>
            <button onClick={() => handleUpdate(file.id)}>Update</button>
            </Link>
            <button onClick={() => handleDelete(file.id)}>Delete</button>
          </li>
        ))}
      </ul>
      <div>
        <button onClick={handleButton}>Download Zip</button>
      </div>
      <Link to='/updateProject' state={{ message : location.state.projID}}>
        <button>Update Project</button>
      </Link>
      <button onClick={handleClick} disabled={isLoading}>
        {isLoading ? 'Deleting...' : 'Delete Project'}
      </button>
      {error1 && <p>Error: {error1}</p>}
    </div>
  );
};

export default ListFilesSimple;
