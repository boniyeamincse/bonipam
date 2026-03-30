
import React from 'react';
import clsx from 'clsx';

export const Input = ({ label, error, icon, className, ...props }) => {
  return (
    <div className="input-group">
      {label && <label className="input-label">{label}</label>}
      <div className="input-container">
        {icon && <div className="input-icon">{icon}</div>}
        <input
          className={clsx(
            'input-field',
            error ? 'error' : '',
            icon ? 'with-icon' : '',
            className
          )}
          {...props}
        />
      </div>
      {error && <span className="error-text">{error}</span>}
    </div>
  );
};
