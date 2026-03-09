import MenuList from "../components/MenuList"

function MenuPage({ menus, loading, error, onAddToCart, onRefresh }) {
  return (
    <section className="page">
      <div className="page-head">
        <div>
          <p className="tag">One Store Only</p>
          <h2>Campus Canteen Menu</h2>
          <p>Order online and pick up at Store A when your queue status is ready.</p>
        </div>
        <button className="ghost" onClick={onRefresh}>Refresh Menu</button>
      </div>

      {loading ? <p className="empty">Loading menu...</p> : null}
      {error ? <p className="error">{error}</p> : null}

      {!loading && !error ? <MenuList menus={menus} onAddToCart={onAddToCart} /> : null}
    </section>
  )
}

export default MenuPage
