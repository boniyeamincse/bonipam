
import React from 'react';
import clsx from 'clsx';

export const Button = ({ children, variant = 'primary', className, ...props }) => {
  return (
    <button
      className={clsx(
        variant === 'primary' ? 'btn-primary' : 'btn-outline',
        className
      )}
      {...props}
    >
      {children}
    </button>
  );
};
