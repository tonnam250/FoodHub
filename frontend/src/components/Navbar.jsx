function Navbar({ activePage, onChangePage, cartCount, currentUser, onLogout }) {
  return (
    <nav className="navbar">
      <div className="brand">
        <div className="brand-badge" aria-hidden="true">
          <img src="/fh-icon.svg" alt="" />
        </div>
        <div>
          <h1>FoodHub</h1>
          <p>Campus Canteen Pickup</p>
        </div>
      </div>

      <div className="nav-links">
        <button className={activePage === "menu" ? "active" : ""} onClick={() => onChangePage("menu")}>Menu</button>
        <button className={activePage === "cart" ? "active" : ""} onClick={() => onChangePage("cart")}>Cart ({cartCount})</button>
        <button className={activePage === "orders" ? "active" : ""} onClick={() => onChangePage("orders")}>Orders</button>
        <button className={activePage === "admin" ? "active" : ""} onClick={() => onChangePage("admin")}>Vendor Admin</button>
        <button className={activePage === "account" ? "active" : ""} onClick={() => onChangePage("account")}>
          {currentUser ? currentUser.name : "Login"}
        </button>
        {currentUser ? (
          <button className="ghost" onClick={onLogout}>Logout</button>
        ) : null}
      </div>
    </nav>
  )
}

export default Navbar
