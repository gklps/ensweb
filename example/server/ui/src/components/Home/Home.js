import React,{ useEffect } from 'react';
import {useState} from 'react';
import { withRouter } from 'react-router-dom';
import { ACCESS_TOKEN_NAME, API_BASE_URL } from '../../constants/apiConstants';
import axios from 'axios'
function Home(props) {
  const [message, setMessage] = useState('');
   useEffect(() => {
        axios.get(API_BASE_URL+'/api/home')
        .then(function (response) {
            if(response.status !== 200){
              redirectToLogin()
            }
            console.log(response.data)
            setMessage(response.data.message)
        })
        .catch(function (error) {
          props.showError(error.response.data.Message)
          redirectToLogin()
        });
      })
    function redirectToLogin() {
    props.history.push('/login');
    }
    return(
        <div className="mt-2">
            User Logged in : {message}
        </div>
    )
}

export default withRouter(Home);