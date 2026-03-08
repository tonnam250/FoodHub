import { useState } from "react"

function AccountPage({ currentUser, authMessage, onRegister, onLogin, onUpdateProfile }) {
  const [registerForm, setRegisterForm] = useState({ name: "", email: "", password: "" })
  const [loginForm, setLoginForm] = useState({ email: "", password: "" })

  const submitRegister = async (e) => {
    e.preventDefault()
    await onRegister(registerForm)
    setRegisterForm({ name: "", email: "", password: "" })
  }

  const submitLogin = async (e) => {
    e.preventDefault()
    await onLogin(loginForm)
  }

  const submitProfile = async (e) => {
    e.preventDefault()
    const formData = new FormData(e.currentTarget)
    await onUpdateProfile({
      name: String(formData.get("name") || ""),
      email: String(formData.get("email") || ""),
      currentPassword: String(formData.get("currentPassword") || ""),
      newPassword: String(formData.get("newPassword") || ""),
    })

    const currentPasswordInput = e.currentTarget.querySelector('input[name="currentPassword"]')
    const newPasswordInput = e.currentTarget.querySelector('input[name="newPassword"]')
    if (currentPasswordInput) currentPasswordInput.value = ""
    if (newPasswordInput) newPasswordInput.value = ""
  }

  return (
    <section className="page">
      <div className="page-head">
        <div>
          <p className="tag">User Account</p>
          <h2>{currentUser ? "Profile Settings" : "Login / Register"}</h2>
          <p>
            {currentUser
              ? `Logged in as ${currentUser.name} (${currentUser.email})`
              : "Login to bind cart and orders to your account."}
          </p>
        </div>
      </div>

      {authMessage ? <p className="hint-text">{authMessage}</p> : null}

      {currentUser ? (
        <form className="auth-form" onSubmit={submitProfile} key={`${currentUser.id}-${currentUser.email}`}>
          <h3>Update Profile</h3>
          <input required name="name" placeholder="Name" defaultValue={currentUser.name || ""} />
          <input required name="email" type="email" placeholder="Email" defaultValue={currentUser.email || ""} />
          <input
            name="currentPassword"
            type="password"
            placeholder="Current Password (required if changing password)"
          />
          <input name="newPassword" type="password" placeholder="New Password (optional)" />
          <button type="submit">Save Changes</button>
        </form>
      ) : (
        <div className="auth-grid">
          <form className="auth-form" onSubmit={submitRegister}>
            <h3>Register</h3>
            <input
              required
              placeholder="Name"
              value={registerForm.name}
              onChange={(e) => setRegisterForm((prev) => ({ ...prev, name: e.target.value }))}
            />
            <input
              required
              type="email"
              placeholder="Email"
              value={registerForm.email}
              onChange={(e) => setRegisterForm((prev) => ({ ...prev, email: e.target.value }))}
            />
            <input
              required
              type="password"
              placeholder="Password"
              value={registerForm.password}
              onChange={(e) => setRegisterForm((prev) => ({ ...prev, password: e.target.value }))}
            />
            <button type="submit">Register</button>
          </form>

          <form className="auth-form" onSubmit={submitLogin}>
            <h3>Login</h3>
            <input
              required
              type="email"
              placeholder="Email"
              value={loginForm.email}
              onChange={(e) => setLoginForm((prev) => ({ ...prev, email: e.target.value }))}
            />
            <input
              required
              type="password"
              placeholder="Password"
              value={loginForm.password}
              onChange={(e) => setLoginForm((prev) => ({ ...prev, password: e.target.value }))}
            />
            <button type="submit">Login</button>
          </form>
        </div>
      )}
    </section>
  )
}

export default AccountPage
