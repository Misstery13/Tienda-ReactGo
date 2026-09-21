import Layout from './components/Layout';
import { BrowserRouter, Routes, Route } from 'react-router-dom'; 
import Dashboard from './components/Dashboard';
import Catalogo from './components/Catalogo';
import MiRed from './components/MiRed';
import Carrito from './components/Carrito';
import Login from './components/Login';
import { CartProvider } from './context/CartContext';
import { AuthProvider } from './context/AuthContext';
import { useState } from 'react';
function App() {
 const [isCollapsed, setIsCollapsed] = useState(() => window.innerWidth < 640);

 return (
 <AuthProvider>
 <CartProvider>
 <BrowserRouter>
 <Routes>
 <Route path="/login" element={<Login />} />
 <Route
 path="/"
 element={
 <Layout
 isCollapsed={isCollapsed}
 onToggle={() => setIsCollapsed((collapsed) => !collapsed)}
 />
 }
 >
 <Route index element={<Dashboard />} />
 <Route path="catalogo" element={<Catalogo />} />
 <Route path="mi-red" element={<MiRed />} />
 <Route path="carrito" element={<Carrito />} />
 <Route path="login" element={<Login />} />
 </Route>
 </Routes>
 </BrowserRouter>
 </CartProvider>
 </AuthProvider>
 );
}
export default App;