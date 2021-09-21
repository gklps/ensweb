import React, { useState, useEffect } from 'react';
import './AlertComponent.css';
function AlertComponent(props) {
    const [modalDisplay, toggleDisplay] = useState('none');
    const [alertType, setAlert] = useState('alert-danger');
    const [message, setMessage] = useState(null);
    const openModal = () => {
        toggleDisplay('block');     
    }
    const closeModal = () => {
        toggleDisplay('none'); 
        props.hideError(null);
        props.hideSuccess(null);
    }
    useEffect(() => {
        if(props.errorMessage !== null) {
            setAlert('alert-danger')
            setMessage(props.errorMessage)
            openModal()
        } else if (props.successMessage !== null){
            setAlert('alert-success')
            setMessage(props.successMessage)
            openModal()
        } else {
            closeModal()
        }
    });
    
    return(
        <div 
            className= {"alert " + alertType + " alert-dismissable mt-4"}
            role="alert" 
            id="alertPopUp"
            style={{ display: modalDisplay }}
        >
            <div className="d-flex alertMessage">
                <span>{message}</span>
                <button type="button" className="close" aria-label="Close" onClick={() => closeModal()}>
                    <span aria-hidden="true">&times;</span>
                </button>
            </div>
            
        </div>
    )
} 

export default AlertComponent