import { useState } from "react"
import CartItem from "../components/CartItem"

function CartPage({ items, onIncrease, onDecrease, onRemove, onCheckout, currentUser, message }) {
  const [paymentMethod, setPaymentMethod] = useState("CASH")
  const [simulatePaymentFail, setSimulatePaymentFail] = useState(false)
  const total = items.reduce((sum, item) => sum + Number(item.price) * item.qty, 0)

  const handleCheckout = async () => {
    await onCheckout({ method: paymentMethod, simulateFail: simulatePaymentFail })
    setSimulatePaymentFail(false)
  }

  return (
    <section className="page">
      <div className="page-head">
        <div>
          <p className="tag">Pickup Cart</p>
          <h2>Your Selected Items</h2>
          <p>
            {currentUser
              ? `Ordering as ${currentUser.name} (${currentUser.email})`
              : "Login required before adding to cart and checkout."}
          </p>
        </div>
      </div>

      {message ? <p className="hint-text">{message}</p> : null}
      {!items.length ? <p className="empty">Your cart is empty. Add food from Menu page.</p> : null}

      <div className="stack">
        {items.map((item) => (
          <CartItem key={item.id} item={item} onIncrease={onIncrease} onDecrease={onDecrease} onRemove={onRemove} />
        ))}
      </div>

      {items.length ? (
        <div className="auth-form" style={{ marginTop: 14 }}>
          <h3>Payment</h3>
          <select value={paymentMethod} onChange={(e) => setPaymentMethod(e.target.value)}>
            <option value="CASH">Cash</option>
            <option value="QR">QR</option>
            <option value="CARD">Card</option>
          </select>
          <label style={{ display: "flex", gap: 8, alignItems: "center", color: "#94a3b8" }}>
            <input
              type="checkbox"
              checked={simulatePaymentFail}
              onChange={(e) => setSimulatePaymentFail(e.target.checked)}
              style={{ width: 16, height: 16 }}
            />
            Simulate payment failure (demo)
          </label>
        </div>
      ) : null}

      <div className="checkout-bar">
        <div>
          <p>Total</p>
          <strong>THB {total.toFixed(2)}</strong>
        </div>
        <button disabled={!items.length || !currentUser} onClick={handleCheckout}>Checkout</button>
      </div>
    </section>
  )
}

export default CartPage
