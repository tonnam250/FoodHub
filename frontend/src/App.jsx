import { useEffect, useMemo, useState } from "react"
import Navbar from "./components/Navbar"
import MenuPage from "./pages/MenuPage"
import CartPage from "./pages/CartPage"
import OrderPage from "./pages/OrderPage"
import AdminMenuPage from "./pages/AdminMenuPage"
import AccountPage from "./pages/AccountPage"
import { createMenuItem, deleteMenuItem, getMenus, updateMenuItem } from "./api/menuApi"
import "./App.css"

const API = {
  auth: "http://localhost:3006",
  user: "http://localhost:3001",
  cart: "http://localhost:3005",
  order: "http://localhost:3003",
  payment: "http://localhost:3004",
}

function decodeJwtPayload(token) {
  try {
    const part = token.split(".")[1]
    const normalized = part.replace(/-/g, "+").replace(/_/g, "/")
    const payload = JSON.parse(atob(normalized))
    return payload
  } catch {
    return null
  }
}

async function parseResponse(res) {
  const text = await res.text()
  let body = text
  try {
    body = text ? JSON.parse(text) : null
  } catch {
    body = text
  }

  if (!res.ok) {
    const message = typeof body === "string" ? body : body?.error || "Request failed"
    throw new Error(message)
  }

  return body
}

async function apiFetch(url, options = {}) {
  const res = await fetch(url, options)
  return parseResponse(res)
}

function App() {
  const [page, setPage] = useState("menu")
  const [menus, setMenus] = useState([])
  const [cart, setCart] = useState([])
  const [orders, setOrders] = useState([])
  const [adminOrders, setAdminOrders] = useState([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState("")
  const [authMessage, setAuthMessage] = useState("")
  const [cartMessage, setCartMessage] = useState("")
  const [ordersLoading, setOrdersLoading] = useState(false)
  const [adminOrdersLoading, setAdminOrdersLoading] = useState(false)
  const [token, setToken] = useState(localStorage.getItem("foodhub_token") || "")
  const [currentUser, setCurrentUser] = useState(null)

  const cartCount = useMemo(() => cart.reduce((sum, item) => sum + item.qty, 0), [cart])

  const menuMap = useMemo(() => {
    const map = new Map()
    menus.forEach((menu) => map.set(menu.id, menu))
    return map
  }, [menus])

  const mapOrders = (allOrders) =>
    (Array.isArray(allOrders) ? allOrders : []).map((order) => ({
      id: order.id,
      userId: order.userId,
      menuId: order.menuId,
      qty: order.qty,
      status: order.status,
      menuName: menuMap.get(order.menuId)?.name || `Menu #${order.menuId}`,
    }))

  const loadMenus = async () => {
    setLoading(true)
    setError("")
    try {
      const data = await getMenus()
      setMenus(Array.isArray(data) ? data : [])
    } catch (err) {
      setError(err.message || "Failed to load menu")
    } finally {
      setLoading(false)
    }
  }

  const loadUserByEmail = async (email, defaultName = "FoodHub User") => {
    const users = await apiFetch(`${API.user}/users`)
    const existing = users.find((u) => String(u.email).toLowerCase() === String(email).toLowerCase())
    if (existing) {
      return existing
    }

    try {
      return await apiFetch(`${API.user}/users`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ name: defaultName, email }),
      })
    } catch {
      const refreshed = await apiFetch(`${API.user}/users`)
      const fallback = refreshed.find((u) => String(u.email).toLowerCase() === String(email).toLowerCase())
      if (!fallback) {
        throw new Error("Cannot map auth user to user-service")
      }
      return fallback
    }
  }

  const loadAllOrders = async () => {
    const allOrders = await apiFetch(`${API.order}/orders`)
    return mapOrders(allOrders)
  }

  const loadUserCart = async (userId) => {
    const items = await apiFetch(`${API.cart}/cart/${userId}`)
    const mapped = (Array.isArray(items) ? items : []).map((item) => {
      const menu = menuMap.get(item.menu_id) || {}
      return {
        id: item.id,
        menuId: item.menu_id,
        qty: item.quantity,
        name: menu.name || `Menu #${item.menu_id}`,
        price: Number(menu.price || 0),
      }
    })
    setCart(mapped)
  }

  const loadUserOrders = async (userId) => {
    setOrdersLoading(true)
    try {
      const mapped = await loadAllOrders()
      setOrders(mapped.filter((order) => order.userId === userId))
    } finally {
      setOrdersLoading(false)
    }
  }

  const loadAdminOrders = async () => {
    setAdminOrdersLoading(true)
    try {
      const mapped = await loadAllOrders()
      setAdminOrders(mapped)
    } finally {
      setAdminOrdersLoading(false)
    }
  }

  const bootstrapSession = async (sessionToken) => {
    if (!sessionToken) {
      setCurrentUser(null)
      setCart([])
      setOrders([])
      return
    }

    await apiFetch(`${API.auth}/auth/validate`, {
      headers: { Authorization: `Bearer ${sessionToken}` },
    })

    const payload = decodeJwtPayload(sessionToken)
    if (!payload?.email) {
      throw new Error("Invalid token payload")
    }

    const user = await loadUserByEmail(payload.email, payload.email.split("@")[0] || "FoodHub User")
    setCurrentUser(user)
    await loadUserCart(user.id)
    await loadUserOrders(user.id)
  }

  useEffect(() => {
    loadMenus()
  }, [])

  useEffect(() => {
    if (!menus.length || !currentUser) {
      return
    }
    loadUserCart(currentUser.id).catch(() => {})
    loadUserOrders(currentUser.id).catch(() => {})
  }, [menus.length])

  useEffect(() => {
    if (!menus.length || page !== "admin") {
      return
    }
    loadAdminOrders().catch(() => {})
  }, [page, menus.length])

  useEffect(() => {
    if (!token) {
      setCurrentUser(null)
      setCart([])
      setOrders([])
      return
    }

    bootstrapSession(token).catch((err) => {
      setAuthMessage(err.message || "Session expired")
      setToken("")
      localStorage.removeItem("foodhub_token")
    })
  }, [token])

  const addToCart = async (menu) => {
    if (!currentUser) {
      setAuthMessage("Please login first.")
      setPage("account")
      return
    }

    setCartMessage("")
    try {
      await apiFetch(`${API.cart}/cart`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ user_id: currentUser.id, menu_id: menu.id, quantity: 1 }),
      })
      await loadUserCart(currentUser.id)
      setCartMessage(`Added ${menu.name} to your cart.`)
    } catch (err) {
      setCartMessage(err.message || "Failed to add to cart")
    }
  }

  const increaseQty = async (item) => {
    if (!currentUser) return
    await apiFetch(`${API.cart}/cart`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ user_id: currentUser.id, menu_id: item.menuId, quantity: 1 }),
    })
    await loadUserCart(currentUser.id)
  }

  const decreaseQty = async (item) => {
    if (!currentUser) return

    if (item.qty <= 1) {
      await apiFetch(`${API.cart}/cart/${item.id}`, { method: "DELETE" })
      await loadUserCart(currentUser.id)
      return
    }

    await apiFetch(`${API.cart}/cart/${item.id}`, { method: "DELETE" })
    await apiFetch(`${API.cart}/cart`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ user_id: currentUser.id, menu_id: item.menuId, quantity: item.qty - 1 }),
    })
    await loadUserCart(currentUser.id)
  }

  const removeFromCart = async (item) => {
    if (!currentUser) return
    await apiFetch(`${API.cart}/cart/${item.id}`, { method: "DELETE" })
    await loadUserCart(currentUser.id)
  }

  const checkout = async ({ method = "CASH", simulateFail = false } = {}) => {
    if (!currentUser || !cart.length) {
      return
    }

    setCartMessage("")
    try {
      for (const item of cart) {
        const createdOrder = await apiFetch(`${API.order}/orders`, {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({
            userId: currentUser.id,
            menuId: item.menuId,
            qty: item.qty,
          }),
        })

        const payment = await apiFetch(`${API.payment}/payments`, {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({
            orderId: createdOrder.id,
            amount: Number(item.price) * item.qty,
            method,
            simulateFail,
          }),
        })

        if (payment.status !== "PAID") {
          throw new Error(`Payment failed for order #${createdOrder.id}`)
        }
      }

      for (const item of cart) {
        await apiFetch(`${API.cart}/cart/${item.id}`, { method: "DELETE" })
      }

      await loadUserCart(currentUser.id)
      await loadUserOrders(currentUser.id)
      if (page === "admin") {
        await loadAdminOrders()
      }
      setCartMessage(`Checkout successful with ${method} payment.`)
      setPage("orders")
    } catch (err) {
      setCartMessage(err.message || "Checkout failed")
    }
  }

  const handleCreateMenu = async (payload) => {
    const created = await createMenuItem(payload)
    setMenus((prev) => [...prev, created])
  }

  const handleUpdateMenu = async (id, payload) => {
    const updated = await updateMenuItem(id, payload)
    setMenus((prev) => prev.map((menu) => (menu.id === id ? updated : menu)))
  }

  const handleDeleteMenu = async (id) => {
    await deleteMenuItem(id)
    setMenus((prev) => prev.filter((menu) => menu.id !== id))
    setCart((prev) => prev.filter((item) => item.menuId !== id))
  }

  const handleAdvanceOrderStatus = async (orderId, status) => {
    await apiFetch(`${API.order}/orders/${orderId}/status`, {
      method: "PUT",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ status }),
    })

    if (currentUser) {
      await loadUserOrders(currentUser.id)
    }
    await loadAdminOrders()
  }

  const handleRegister = async ({ name, email, password }) => {
    setAuthMessage("")
    try {
      await apiFetch(`${API.auth}/auth/register`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ name, email, password }),
      })

      try {
        await apiFetch(`${API.user}/users`, {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({ name, email }),
        })
      } catch {
        // ignore duplicate email in user-service
      }

      setAuthMessage("Register success. Now login.")
    } catch (err) {
      setAuthMessage(err.message || "Register failed")
    }
  }

  const handleLogin = async ({ email, password }) => {
    setAuthMessage("")
    try {
      const data = await apiFetch(`${API.auth}/auth/login`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ email, password }),
      })

      localStorage.setItem("foodhub_token", data.token)
      setToken(data.token)
      setAuthMessage("Login success")
      setPage("menu")
    } catch (err) {
      setAuthMessage(err.message || "Login failed")
    }
  }

  const handleUpdateProfile = async ({ name, email, currentPassword, newPassword }) => {
    if (!currentUser || !token) {
      setAuthMessage("Please login first.")
      return
    }

    setAuthMessage("")
    try {
      const authPayload = await apiFetch(`${API.auth}/auth/profile`, {
        method: "PUT",
        headers: {
          "Content-Type": "application/json",
          Authorization: `Bearer ${token}`,
        },
        body: JSON.stringify({ name, email, currentPassword, newPassword }),
      })

      const updatedUser = await apiFetch(`${API.user}/users/${currentUser.id}`, {
        method: "PUT",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ name, email }),
      })

      setCurrentUser(updatedUser)

      if (authPayload?.token) {
        localStorage.setItem("foodhub_token", authPayload.token)
        setToken(authPayload.token)
      }

      setAuthMessage("Profile updated.")
    } catch (err) {
      setAuthMessage(err.message || "Failed to update profile")
    }
  }

  const handleLogout = async () => {
    try {
      if (token) {
        await apiFetch(`${API.auth}/auth/logout`, {
          method: "POST",
          headers: { Authorization: `Bearer ${token}` },
        })
      }
    } catch {
      // ignore logout errors
    }

    localStorage.removeItem("foodhub_token")
    setToken("")
    setCurrentUser(null)
    setCart([])
    setOrders([])
    setAuthMessage("Logged out")
    setPage("account")
  }

  return (
    <div className="app">
      <div className="bg-orb bg-orb-1" />
      <div className="bg-orb bg-orb-2" />

      <Navbar
        activePage={page}
        onChangePage={setPage}
        cartCount={cartCount}
        currentUser={currentUser}
        onLogout={handleLogout}
      />

      {page === "menu" ? (
        <MenuPage menus={menus} loading={loading} error={error} onAddToCart={addToCart} onRefresh={loadMenus} />
      ) : null}

      {page === "cart" ? (
        <CartPage
          items={cart}
          onIncrease={increaseQty}
          onDecrease={decreaseQty}
          onRemove={removeFromCart}
          onCheckout={checkout}
          currentUser={currentUser}
          message={cartMessage}
        />
      ) : null}

      {page === "orders" ? <OrderPage orders={orders} currentUser={currentUser} loading={ordersLoading} /> : null}

      {page === "admin" ? (
        <AdminMenuPage
          menus={menus}
          onCreate={handleCreateMenu}
          onUpdate={handleUpdateMenu}
          onDelete={handleDeleteMenu}
          onRefresh={loadMenus}
          orders={adminOrders}
          ordersLoading={adminOrdersLoading}
          onRefreshOrders={loadAdminOrders}
          onAdvanceOrderStatus={handleAdvanceOrderStatus}
        />
      ) : null}

      {page === "account" ? (
        <AccountPage
          currentUser={currentUser}
          authMessage={authMessage}
          onRegister={handleRegister}
          onLogin={handleLogin}
          onUpdateProfile={handleUpdateProfile}
        />
      ) : null}
    </div>
  )
}

export default App
