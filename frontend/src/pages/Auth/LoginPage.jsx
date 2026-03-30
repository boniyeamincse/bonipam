
import React, { useState } from 'react';
import { Button } from '../../components/UI/Button';
import { Input } from '../../components/UI/Input';

import { Lock, Mail, Code, Globe, Shield } from 'lucide-react';
import { motion } from 'framer-motion';

const LoginPage = () => {
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [isLoading, setIsLoading] = useState(false);
  const [errors, setErrors] = useState({});

  console.log('LoginPage is rendering with icons');

  const handleLogin = async (e) => {
    e.preventDefault();
    setIsLoading(true);
    setErrors({});

    const newErrors = {};
    if (!email) newErrors.email = 'Email is required';
    if (!password) newErrors.password = 'Password is required';
    
    if (Object.keys(newErrors).length > 0) {
      setErrors(newErrors);
      setIsLoading(false);
      return;
    }

    try {
      const response = await fetch('http://localhost:8081/api/v1/auth/login', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ email, password }),
      });

      const result = await response.json();

      if (!response.ok) {
        throw new Error(result.error?.message || 'Login failed');
      }

      console.log('Login successful!', result.data);
      localStorage.setItem('access_token', result.data.access_token);
      alert('Login successful! Welcome to Boni PAM.');
    } catch (err) {
      console.error('Login error:', err);
      setErrors({ form: err.message });
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <div className="min-h-screen flex items-center justify-center p-4 bg-[#0f172a]">
      <motion.div 
        initial={{ opacity: 0, y: 20 }}
        animate={{ opacity: 1, y: 0 }}
        className="glass-card w-full max-w-md p-8 flex flex-col gap-8"
      >
        <div className="flex flex-col items-center gap-2">
          <div className="w-12 h-12 bg-indigo-600 rounded-xl flex items-center justify-center shadow-lg shadow-indigo-500/20 text-white">
            <Shield size={28} />
          </div>
          <h1 className="text-2xl font-bold text-white">Boni PAM</h1>
          <p className="text-slate-400">Enterprise Access Management</p>
        </div>

        <form onSubmit={handleLogin} className="flex flex-col gap-5">
          {errors.form && (
            <div className="p-3 rounded-lg bg-red-500/10 border border-red-500/50 text-red-500 text-sm">
              {errors.form}
            </div>
          )}
          <Input 
            label="Email Address"
            type="email"
            placeholder="name@company.com"
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            error={errors.email}
            icon={<Mail size={18} />}
          />
          <Input 
            label="Password"
            type="password"
            placeholder="••••••••"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            error={errors.password}
            icon={<Lock size={18} />}
          />
          <Button type="submit" className="w-full mt-2" disabled={isLoading}>
            {isLoading ? 'Signing In...' : 'Sign In'}
          </Button>
        </form>

        <div className="grid grid-cols-2 gap-4">
          <Button variant="outline" className="flex items-center justify-center gap-2 text-white border-slate-700">
            <Code size={18} /> GitHub
          </Button>
          <Button variant="outline" className="flex items-center justify-center gap-2 text-white border-slate-700">
            <Globe size={18} /> Google
          </Button>
        </div>

        <p className="text-center text-sm text-slate-400">
          Don't have an account? <a href="#" className="text-indigo-400 hover:underline font-medium">Contact administrator</a>
        </p>
      </motion.div>
    </div>
  );
};

export default LoginPage;
