import './app/bootstrap' // должен быть первым: см. комментарий в bootstrap.ts

import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'

import { App } from './app/app'

import './global/index.css'

const container = document.getElementById('root')
if (!container) throw new Error('Корневой элемент #root не найден')

createRoot(container).render(
  <StrictMode>
    <App />
  </StrictMode>,
)
