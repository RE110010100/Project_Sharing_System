import React, { useState, useEffect } from 'react';
import { useLocation, useParams } from 'react-router-dom';


function UserNotificationsSimple() {

    const {userID} = useParams();
    const [notifications, setNotifications] = useState([]);

    useEffect(() => {
        // Fetch notifications when the component mounts
        fetchNotifications();
    }, []);

    const fetchNotifications = async () => {
        try {
            // Fetch notifications from the API
            const response = await fetch(`http://localhost:8085/fetch_notifications?userID=${userID}`);
            if (!response.ok) {
                throw new Error('Failed to fetch notifications');
            }
            const data = await response.json();
            // Update state with the fetched notifications
            setNotifications(data);
            console.log(data)
        } catch (error) {
            console.error(error);
        }
    };

    return (
        <div>
            <h1>Notifications</h1>
            <ul>
                {notifications.map((notification, index) => (
                    <li key={index}>
                        <p>{notification.Message}</p>
                        <p>{notification.Time}</p>
                    </li>
                ))}
            </ul>
        </div>
    );
}

export default UserNotificationsSimple;
