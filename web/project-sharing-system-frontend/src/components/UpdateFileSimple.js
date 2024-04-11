import React, { useState } from 'react';
import { useLocation, useNavigate, useParams } from 'react-router-dom';

const UpdateFileSimple = () => {

  const {userID} = useParams();

  const navigate = useNavigate();

  const location = useLocation();
  const [selectedFile, setSelectedFile] = useState(null);

  console.log(location.state.state.FileName)

  const handleFileChange = (event) => {
    setSelectedFile(event.target.files[0]);
    
  };

  const handleUpdateClick = async () => {
    if (!selectedFile) {
      alert('Please select a file.');
      return;
    }

    const formData = new FormData();
    formData.append('file', selectedFile);
    formData.append('file_key', location.state.state.FileName)
    formData.append('file_type', location.state.state.type)
    formData.append('file_id', location.state.state.id)
    formData.append('project_title', location.state.project)
    formData.append('user_id', userID)

    console.log(selectedFile)

    try {
      const response = await fetch('http://localhost:8080/updateFile', {
        method: 'POST',
        body: formData,
      });

      if (response.ok) {
        console.log('File updated successfully.');
        navigate(`/profile/${userID}`)
        // Handle success as needed
      } else {
        console.error('Failed to upload file.');
        // Handle error as needed
      }
    } catch (error) {
      console.error('Error uploading file:', error);
      // Handle error as needed
    }
  };

  return (
    <div>
      <input type="file" onChange={handleFileChange} />
      <button onClick={handleUpdateClick}>Update</button>
    </div>
  );
};

export default UpdateFileSimple;
