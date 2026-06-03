import { useState } from 'react';
import './index.css';

// Componentes das Páginas
import Dashboard from './pages/Dashboard';
import Registro from './pages/Registro';

function App() {
  const [currentRoute, setCurrentRoute] = useState('dashboard');

  const navigate = (route) => {
    // Usando View Transitions API (Fallback se não suportado)
    if (!document.startViewTransition) {
      setCurrentRoute(route);
      return;
    }
    document.startViewTransition(() => {
      setCurrentRoute(route);
    });
  };

  return (
    <>
      <nav className="top-nav">
        <div className="nav-links">
          <a 
            href="#dashboard" 
            className={currentRoute === 'dashboard' ? 'active' : ''}
            onClick={(e) => { e.preventDefault(); navigate('dashboard'); }}
          >
            Dashboard de Kitnets
          </a>
          <a 
            href="#registro" 
            className={currentRoute === 'registro' ? 'active' : ''}
            onClick={(e) => { e.preventDefault(); navigate('registro'); }}
          >
            Registrar Imóvel
          </a>
        </div>
        <div style={{ fontWeight: 800, color: 'var(--color-primary)' }}>
          Cartório Chain 🧱
        </div>
      </nav>

      <main className="container">
        {currentRoute === 'dashboard' && <Dashboard onNavigate={navigate} />}
        {currentRoute === 'registro' && <Registro onNavigate={navigate} />}
      </main>
    </>
  );
}

export default App;
