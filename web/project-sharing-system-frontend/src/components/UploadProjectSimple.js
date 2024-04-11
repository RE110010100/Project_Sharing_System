import React, { useState } from 'react';
import {useNavigate, useLocation, useParams } from 'react-router-dom'

const UploadProject = () => {


  const {userID} = useParams();

  // Use useLocation to access the location object
  const location = useLocation();
  
  // Access the data passed during navigation
  const message = location.state.message;

  console.log(message)

  const navigate = useNavigate();
  const [zipFile, setZipFile] = useState(null);
  //const [projID, setProjID] = useState(null)

  //setProjID(message)

  const handleFileUpload = (event) => {
    const uploadedFile = event.target.files[0];
    console.log(event.target.files[0])
    if (uploadedFile && uploadedFile.name.endsWith('.zip')) {
      setZipFile(uploadedFile);
      
    } else {
      alert('Please select a valid ZIP file.');
    }
  };

  const uploadZipFile = async () => {
    if (!zipFile) {
      alert('Please select a ZIP file to upload.');
      return;
    }

    console.log(zipFile)
    //console.log(userID)

    const formData = new FormData();
    formData.append('zipFile', zipFile);
    formData.append('project_id', message)
    formData.append('user_id', userID)
    formData.append('project_title',location.state.title)

    try {
      const response = await fetch('http://localhost:8080/upload', {
        method: 'POST',
        body: formData,
      });

      if (!response.ok) {
        throw new Error('Failed to upload ZIP file.');
      }

      navigate(`/profile'/${userID}`)
      alert('ZIP file uploaded successfully!');
      
    } catch (error) {
      console.error('Error uploading ZIP file:', error);
      alert('Failed to upload ZIP file. Please try again later.');
    }
  };

  return (
    <div>
      <h2>Upload a ZIP File</h2>
      <input type="file" accept=".zip" onChange={handleFileUpload} />
      {zipFile && (
        <div>
          <h3>Uploaded File Information:</h3>
          <p>Name: {zipFile.name}</p>
          <p>Type: {zipFile.type}</p>
          <p>Size: {zipFile.size} bytes</p>
          <button onClick={uploadZipFile}>Upload ZIP File</button>
        </div>
      )}
    </div>
  );
};

export default UploadProject;
