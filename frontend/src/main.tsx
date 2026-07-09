import { StrictMode } from 'react';
import { createRoot } from 'react-dom/client';
import dayjs from 'dayjs';
import jalaliday from 'jalaliday';
import { App } from './App';
import './index.css';

dayjs.extend(jalaliday);

createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <App />
  </StrictMode>,
);
