import MenuCard from "./MenuCard"

function MenuList({ menus, onAddToCart }) {
  if (!menus.length) {
    return <p className="empty">No menu items yet. Vendor can add items in Admin page.</p>
  }

  return (
    <div className="menu-grid">
      {menus.map((item) => (
        <MenuCard key={item.id} item={item} onAddToCart={onAddToCart} />
      ))}
    </div>
  )
}

export default MenuList
