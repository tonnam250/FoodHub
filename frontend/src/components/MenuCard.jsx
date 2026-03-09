const FALLBACK_IMAGE =
  "https://images.unsplash.com/photo-1547592180-85f173990554?auto=format&fit=crop&w=800&q=80"

function MenuCard({ item, onAddToCart }) {
  const imageSrc = item.image_url?.trim() ? item.image_url : FALLBACK_IMAGE

  return (
    <article className="menu-card">
      <div className="menu-image">
        <img
          src={imageSrc}
          alt={item.name}
          loading="lazy"
          onError={(e) => {
            e.currentTarget.src = FALLBACK_IMAGE
          }}
        />
        <span>Store A</span>
      </div>
      <div className="menu-content">
        <h3>{item.name}</h3>
        <p className="menu-price">THB {Number(item.price).toFixed(2)}</p>
        <button onClick={() => onAddToCart(item)}>Add to Cart</button>
      </div>
    </article>
  )
}

export default MenuCard
