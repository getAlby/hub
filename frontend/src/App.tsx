import { useEffect, useState } from 'react'
import reactLogo from './assets/react.svg'
import viteLogo from '/vite.svg'
import { BrowserRouter, Routes, Route } from "react-router-dom";
import './App.css'

function App() {
  const [messageFromServer, setMessageFromServer] = useState('')

  useEffect(() => {
    const fetchData = async () => {
      const response = await fetch('/api/message')
      const data = await response.json()
      setMessageFromServer(data.message)
    }

    fetchData().catch((e) => console.error(e))
  }, [])

  return (
    <BrowserRouter>
      <Routes>
        <Route
          path="/"
          element={
            <div className="container">
              <div>
                <a href='https://vitejs.dev' target='_blank'>
                  <img src={viteLogo} className='logo' alt='Vite logo' />
                </a>
                <a href='https://react.dev' target='_blank'>
                  <img src={reactLogo} className='logo react' alt='React logo' />
                </a>
              </div>
              <h1>Golang + Vite + React</h1>
              <h2>{messageFromServer}</h2>
              <a href='/aboutt' target='_blank'>
                About page
              </a>
            </div>
          }
        />
        <Route
          path="/alby"
          element={
            <div className="container">
              ALBY
            </div>
          }
        />
        <Route
          path="/lnd"
          element={
            <div className="container">
              LND
            </div>
          }
        />
      </Routes>
    </BrowserRouter>
  )
}

export default App
