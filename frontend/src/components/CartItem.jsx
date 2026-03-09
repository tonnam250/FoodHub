function CartItem({ item, onIncrease, onDecrease, onRemove }) {
  return (
    <article className="cart-item">
      <div>
        <h4>{item.name}</h4>
        <p>THB {Number(item.price).toFixed(2)} each</p>
      </div>

      <div className="qty-control">
        <button onClick={() => onDecrease(item)}>-</button>
        <span>{item.qty}</span>
        <button onClick={() => onIncrease(item)}>+</button>
      </div>

      <div className="cart-actions">
        <strong>THB {(item.qty * Number(item.price)).toFixed(2)}</strong>
        <button className="ghost" onClick={() => onRemove(item)}>Remove</button>
      </div>
    </article>
  )
}

export default CartItem
