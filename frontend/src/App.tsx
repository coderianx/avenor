import logo from './assets/logo.png'

function App() {
  return (
    <main className="welcome">
      <h1 className="welcome-title">Welcome to Avenor</h1>
      <img src={logo} className="welcome-logo" alt="Avenor logo" />
      <button type="button" className="welcome-button">
        Get started <span className="arrow" aria-hidden="true">→</span>
      </button>
    </main>
  )
}

export default App
